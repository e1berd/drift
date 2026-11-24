package workspace

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/typeid"
)

const (
	defaultMessagePageSize = 64
	maxMessagePageSize     = 256
)

// listMessagesResponse is a contiguous, chronological slice of the transcript.
type listMessagesResponse struct {
	Messages   []Message `json:"messages"`
	HasOlder   bool      `json:"hasOlder"`
	HasNewer   bool      `json:"hasNewer"`
	LastPrompt *Message  `json:"lastPrompt,omitempty"`
}

type messageWindow struct {
	q       *db.Queries
	session pgtype.UUID
	size    int32
	start   *db.GetMessagePositionRow
	end     *db.GetMessagePositionRow
}

func messagePageSize(c *echo.Context) int32 {
	n, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil || n <= 0 {
		return defaultMessagePageSize
	}
	return int32(min(n, maxMessagePageSize))
}

func (h *Handler) messagePosition(ctx context.Context, session pgtype.UUID, raw string) (*db.GetMessagePositionRow, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := typeid.ParseMessageID(raw)
	if err != nil {
		return nil, badRequest("invalid message id " + raw)
	}
	position, err := h.q.GetMessagePosition(ctx, db.GetMessagePositionParams{ID: id, SessionID: session})
	if err != nil {
		return nil, notFound(err)
	}
	return &position, nil
}

func (h *Handler) listMessages(c *echo.Context) error {
	s, err := h.session(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	window := messageWindow{q: h.q, session: s.ID, size: messagePageSize(c)}
	if window.start, err = h.messagePosition(ctx, s.ID, c.QueryParam("start")); err != nil {
		return err
	}
	if window.end, err = h.messagePosition(ctx, s.ID, c.QueryParam("end")); err != nil {
		return err
	}
	out, err := window.read(ctx)
	if err != nil {
		return err
	}
	if out.LastPrompt, err = h.lastPrompt(ctx, s.ID); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) lastPrompt(ctx context.Context, session pgtype.UUID) (*Message, error) {
	m, err := h.q.GetLastPrompt(ctx, session)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	prompt := messageDTO(m)
	return &prompt, nil
}

func (w messageWindow) read(ctx context.Context) (listMessagesResponse, error) {
	if w.start != nil {
		return w.readForward(ctx)
	}
	return w.readBackward(ctx)
}

func (w messageWindow) readForward(ctx context.Context) (listMessagesResponse, error) {
	params := db.ListMessagesFromParams{
		SessionID: w.session, Limit: w.size + 1, StartCreated: w.start.CreatedAt, StartID: w.start.ID,
	}
	if w.end != nil {
		params.EndCreated, params.EndID = w.end.CreatedAt, w.end.ID
	}
	rows, err := w.q.ListMessagesFrom(ctx, params)
	if err != nil {
		return listMessagesResponse{}, err
	}
	out := listMessagesResponse{HasNewer: len(rows) > int(w.size)}
	rows = rows[:min(len(rows), int(w.size))]
	if !out.HasNewer && w.end != nil {
		if out.HasNewer, err = w.hasAfter(ctx, *w.end); err != nil {
			return out, err
		}
	}
	out.HasOlder, err = w.q.HasMessagesBefore(ctx, db.HasMessagesBeforeParams{
		SessionID: w.session, Created: w.start.CreatedAt, ID: w.start.ID,
	})
	out.Messages = messageDTOs(rows)
	return out, err
}

func (w messageWindow) readBackward(ctx context.Context) (listMessagesResponse, error) {
	params := db.ListMessagesUpToParams{SessionID: w.session, Limit: w.size + 1}
	if w.end != nil {
		params.EndCreated, params.EndID = w.end.CreatedAt, w.end.ID
	}
	rows, err := w.q.ListMessagesUpTo(ctx, params)
	if err != nil {
		return listMessagesResponse{}, err
	}
	out := listMessagesResponse{HasOlder: len(rows) > int(w.size)}
	rows = rows[:min(len(rows), int(w.size))]
	slices.Reverse(rows)
	if w.end != nil {
		if out.HasNewer, err = w.hasAfter(ctx, *w.end); err != nil {
			return out, err
		}
	}
	out.Messages = messageDTOs(rows)
	return out, nil
}

func (w messageWindow) hasAfter(ctx context.Context, position db.GetMessagePositionRow) (bool, error) {
	return w.q.HasMessagesAfter(ctx, db.HasMessagesAfterParams{
		SessionID: w.session, Created: position.CreatedAt, ID: position.ID,
	})
}

func messageDTOs(rows []db.Message) []Message {
	out := make([]Message, 0, len(rows))
	for _, m := range rows {
		out = append(out, messageDTO(m))
	}
	return out
}
