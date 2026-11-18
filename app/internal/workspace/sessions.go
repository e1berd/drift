package workspace

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

const maxMessageBytes = 256 << 10

var agents = map[string]bool{"claude": true, "codex": true, "antigravity": true, "kimi": true, "mimo": true}

type listSessionsResponse struct {
	Sessions      []Session `json:"sessions"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

func sessionPage(rows []db.Session, size int32) listSessionsResponse {
	out := listSessionsResponse{Sessions: make([]Session, 0, min(len(rows), int(size)))}
	for i, s := range rows {
		if int32(i) == size {
			last := rows[i-1]
			out.NextPageToken = encodeCursor(last.UpdatedAt, last.ID)
			break
		}
		out.Sessions = append(out.Sessions, sessionDTO(s))
	}
	return out
}

func (h *Handler) listSessions(c *echo.Context) error {
	id, err := parseID(c, "project")
	if err != nil {
		return err
	}
	before, beforeID, err := decodeCursor(c)
	if err != nil {
		return err
	}
	size := pageSize(c)
	rows, err := h.q.ListSessions(c.Request().Context(), db.ListSessionsParams{
		ProjectID: id, OwnerID: identity.From(c), Limit: size + 1, BeforeUpdated: before, BeforeID: beforeID,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sessionPage(rows, size))
}

type sessionInput struct {
	runOptions
	Title         string   `json:"title"`
	Agent         string   `json:"agent"`
	Text          string   `json:"text"`
	AttachmentIDs []string `json:"attachmentIds"`
}

type createSessionResponse struct {
	Session Session  `json:"session"`
	Message *Message `json:"message,omitempty"`
}

func titleFrom(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	if utf8.RuneCountInString(line) <= 60 {
		return line
	}
	return string([]rune(line)[:60]) + "…"
}

func (h *Handler) createSession(c *echo.Context) error {
	p, err := h.project(c)
	if err != nil {
		return err
	}
	return h.insertSession(c, p.ID, "new:"+uuidString(p.ID))
}

func (h *Handler) insertSession(c *echo.Context, projectID pgtype.UUID, newDraftKey string) error {
	var in sessionInput
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	if !agents[in.Agent] {
		return badRequest("unknown agent")
	}
	if err := in.validate(); err != nil {
		return err
	}
	text := strings.TrimSpace(in.Text)
	if len(text) > maxMessageBytes {
		return badRequest("message is too long")
	}
	attachmentIDs, err := parseAttachmentIDs(in.AttachmentIDs)
	if err != nil {
		return err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = titleFrom(text)
	}
	if title == "" && len(attachmentIDs) > 0 {
		title = defaultAttachmentName
	}
	if title, ok := cleanName(title, 200); !ok || title == "" {
		return badRequest("title or text is required")
	}

	owner := identity.From(c)
	ctx := c.Request().Context()
	var out createSessionResponse
	var sessionID pgtype.UUID
	var draftsDeleted int64
	err = pgx.BeginFunc(ctx, h.pool, func(tx pgx.Tx) error {
		q := h.q.WithTx(tx)
		s, err := q.CreateSession(ctx, db.CreateSessionParams{ProjectID: projectID, OwnerID: owner, Title: title, Agent: in.Agent})
		if err != nil {
			return err
		}
		out.Session = sessionDTO(s)
		sessionID = s.ID
		if text != "" || len(attachmentIDs) > 0 {
			m, err := q.CreateMessage(ctx, db.CreateMessageParams{SessionID: s.ID, Role: "user", Kind: "text", Body: textBody(text)})
			if err != nil {
				return err
			}
			if m, err = claimAttachments(ctx, q, owner, m, text, attachmentIDs); err != nil {
				return err
			}
			dto := messageDTO(m)
			out.Message = &dto
		}
		draftsDeleted, err = q.DeleteDraft(ctx, db.DeleteDraftParams{OwnerID: owner, Key: newDraftKey})
		return err
	})
	if err != nil {
		return err
	}

	h.publish(c, "session.created", out.Session)
	if out.Message != nil {
		h.publish(c, "message.created", out.Message)
	}
	if draftsDeleted > 0 {
		h.publish(c, "draft.deleted", deleted{Key: newDraftKey})
	}
	h.startRun(c, sessionID, text, attachmentIDs, in.runOptions)
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) session(c *echo.Context) (db.Session, error) {
	id, err := parseID(c, "session")
	if err != nil {
		return db.Session{}, err
	}
	s, err := h.q.GetSession(c.Request().Context(), db.GetSessionParams{ID: id, OwnerID: identity.From(c)})
	return s, notFound(err)
}

func (h *Handler) getSession(c *echo.Context) error {
	s, err := h.session(c)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sessionDTO(s))
}

type sessionUpdate struct {
	Title *string `json:"title"`
	Agent *string `json:"agent"`
}

func (u sessionUpdate) params(id pgtype.UUID, owner string) (db.UpdateSessionParams, error) {
	out := db.UpdateSessionParams{ID: id, OwnerID: owner}
	if u.Title == nil && u.Agent == nil {
		return out, badRequest("title or agent is required")
	}
	if u.Title != nil {
		title, ok := cleanName(*u.Title, 200)
		if !ok {
			return out, badRequest("title must be 1–200 characters")
		}
		out.Title = pgtype.Text{String: title, Valid: true}
	}
	if u.Agent != nil {
		if !agents[*u.Agent] {
			return out, badRequest("unknown agent")
		}
		out.Agent = pgtype.Text{String: *u.Agent, Valid: true}
	}
	return out, nil
}

func (h *Handler) updateSession(c *echo.Context) error {
	id, err := parseID(c, "session")
	if err != nil {
		return err
	}
	var in sessionUpdate
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	params, err := in.params(id, identity.From(c))
	if err != nil {
		return err
	}
	s, err := h.q.UpdateSession(c.Request().Context(), params)
	if err != nil {
		return notFound(err)
	}
	dto := sessionDTO(s)
	h.publish(c, "session.updated", dto)
	return c.JSON(http.StatusOK, dto)
}

func (h *Handler) deleteSession(c *echo.Context) error {
	s, err := h.session(c)
	if err != nil {
		return err
	}
	owner := identity.From(c)
	ctx := c.Request().Context()
	if h.runner != nil {
		h.runner.Stop(owner, s.ID)
	}
	if _, err := h.q.DeleteSession(ctx, db.DeleteSessionParams{ID: s.ID, OwnerID: owner}); err != nil {
		return err
	}
	if h.terminals != nil {
		h.terminals.Close(owner, uuidString(s.ID))
	}
	if !s.ProjectID.Valid {
		if err := h.folders.RemoveChat(s.ID); err != nil {
			h.log.Warn("delete chat folder", "err", err)
		}
	}
	if err := h.q.DeleteDraftsForSession(ctx, db.DeleteDraftsForSessionParams{OwnerID: owner, SessionID: uuidString(s.ID)}); err != nil {
		h.log.Warn("delete session drafts", "err", err)
	}
	h.publish(c, "session.deleted", deleted{ID: uuidString(s.ID), ProjectID: uuidString(s.ProjectID)})
	return c.NoContent(http.StatusNoContent)
}

type listMessagesResponse struct {
	Messages      []Message `json:"messages"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

func (h *Handler) listMessages(c *echo.Context) error {
	s, err := h.session(c)
	if err != nil {
		return err
	}
	before, beforeID, err := decodeCursor(c)
	if err != nil {
		return err
	}
	size := pageSize(c)
	rows, err := h.q.ListMessages(c.Request().Context(), db.ListMessagesParams{
		SessionID: s.ID, Limit: size + 1, BeforeCreated: before, BeforeID: beforeID,
	})
	if err != nil {
		return err
	}
	out := listMessagesResponse{Messages: make([]Message, 0, len(rows))}
	for i, m := range rows {
		if int32(i) == size {
			last := rows[i-1]
			out.NextPageToken = encodeCursor(last.CreatedAt, last.ID)
			break
		}
		out.Messages = append(out.Messages, messageDTO(m))
	}
	return c.JSON(http.StatusOK, out)
}

type messageInput struct {
	runOptions
	Text          string   `json:"text"`
	AttachmentIDs []string `json:"attachmentIds"`
}

func (h *Handler) createMessage(c *echo.Context) error {
	s, err := h.session(c)
	if err != nil {
		return err
	}
	var in messageInput
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	text := strings.TrimSpace(in.Text)
	attachmentIDs, err := parseAttachmentIDs(in.AttachmentIDs)
	if err != nil {
		return err
	}
	if text == "" && len(attachmentIDs) == 0 || len(text) > maxMessageBytes {
		return badRequest("text must be 1 byte to 256 KiB or come with attachments")
	}
	if err := in.validate(); err != nil {
		return err
	}
	owner := identity.From(c)
	ctx := c.Request().Context()
	var msg Message
	var sess Session
	var draftsDeleted int64
	err = pgx.BeginFunc(ctx, h.pool, func(tx pgx.Tx) error {
		q := h.q.WithTx(tx)
		m, err := q.CreateMessage(ctx, db.CreateMessageParams{SessionID: s.ID, Role: "user", Kind: "text", Body: textBody(text)})
		if err != nil {
			return err
		}
		if m, err = claimAttachments(ctx, q, owner, m, text, attachmentIDs); err != nil {
			return err
		}
		msg = messageDTO(m)
		if err := q.TouchSession(ctx, s.ID); err != nil {
			return err
		}
		updated, err := q.GetSession(ctx, db.GetSessionParams{ID: s.ID, OwnerID: owner})
		if err != nil {
			return err
		}
		sess = sessionDTO(updated)
		draftsDeleted, err = q.DeleteDraft(ctx, db.DeleteDraftParams{OwnerID: owner, Key: "session:" + uuidString(s.ID)})
		return err
	})
	if err != nil {
		return err
	}
	h.publish(c, "message.created", msg)
	h.publish(c, "session.updated", sess)
	if draftsDeleted > 0 {
		h.publish(c, "draft.deleted", deleted{Key: "session:" + uuidString(s.ID)})
	}
	h.startRun(c, s.ID, text, attachmentIDs, in.runOptions)
	return c.JSON(http.StatusCreated, msg)
}

func textBody(text string) []byte {
	b, _ := json.Marshal(userMessageBody{Text: text})
	return b
}
