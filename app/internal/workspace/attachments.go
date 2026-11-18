package workspace

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

const (
	maxAttachmentBytes    = 5 << 20
	maxMessageAttachments = 10
	defaultAttachmentName = "image"
)

var attachableTypes = map[string]bool{
	"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true,
}

// Attachment is an image uploaded for a user message. It belongs to no message until a
// message or session created with its id claims it; unclaimed uploads are dropped after a day.
type Attachment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int32  `json:"size"`
	Type string `json:"type"`
}

type userMessageBody struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

func attachmentName(raw string) string {
	name := path.Base(strings.ReplaceAll(strings.TrimSpace(raw), `\`, "/"))
	if name, ok := cleanName(name, 255); ok && name != "." && name != "/" {
		return name
	}
	return defaultAttachmentName
}

func (h *Handler) uploadAttachment(c *echo.Context) error {
	data, err := io.ReadAll(io.LimitReader(c.Request().Body, maxAttachmentBytes+1))
	if err != nil {
		return badRequest("could not read the upload")
	}
	if len(data) == 0 {
		return badRequest("the file is empty")
	}
	if len(data) > maxAttachmentBytes {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "an image must be at most 5 MB")
	}
	mediaType := http.DetectContentType(data)
	if !attachableTypes[mediaType] {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "only PNG, JPEG, GIF and WebP images can be attached")
	}
	ctx := c.Request().Context()
	if _, err := h.q.DeleteStaleAttachments(ctx); err != nil {
		h.log.Warn("delete stale attachments", "err", err)
	}
	row, err := h.q.CreateAttachment(ctx, db.CreateAttachmentParams{
		OwnerID: identity.From(c), Name: attachmentName(c.QueryParam("name")),
		MediaType: mediaType, Size: int32(len(data)), Data: data,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, Attachment{ID: uuidString(row.ID), Name: row.Name, Size: row.Size, Type: row.MediaType})
}

func (h *Handler) attachmentContent(c *echo.Context) error {
	id, err := parseID(c, "attachment")
	if err != nil {
		return err
	}
	row, err := h.q.GetAttachmentContent(c.Request().Context(), db.GetAttachmentContentParams{ID: id, OwnerID: identity.From(c)})
	if err != nil {
		return notFound(err)
	}
	header := c.Response().Header()
	header.Set("Cache-Control", "private, max-age=31536000, immutable")
	header.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	return c.Blob(http.StatusOK, row.MediaType, row.Data)
}

func parseAttachmentIDs(raw []string) ([]pgtype.UUID, error) {
	if len(raw) > maxMessageAttachments {
		return nil, badRequest("at most 10 attachments per message")
	}
	ids := make([]pgtype.UUID, 0, len(raw))
	seen := make(map[pgtype.UUID]bool, len(raw))
	for _, value := range raw {
		var id pgtype.UUID
		if err := id.Scan(value); err != nil {
			return nil, badRequest("invalid attachment id")
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func claimAttachments(ctx context.Context, q *db.Queries, owner string, message db.Message, text string, ids []pgtype.UUID) (db.Message, error) {
	if len(ids) == 0 {
		return message, nil
	}
	rows, err := q.ClaimAttachments(ctx, db.ClaimAttachmentsParams{MessageID: message.ID, OwnerID: owner, Ids: ids})
	if err != nil {
		return message, err
	}
	if len(rows) != len(ids) {
		return message, badRequest("unknown or already sent attachment")
	}
	byID := make(map[pgtype.UUID]db.ClaimAttachmentsRow, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	body := userMessageBody{Text: text, Attachments: make([]Attachment, 0, len(ids))}
	for _, id := range ids {
		row := byID[id]
		body.Attachments = append(body.Attachments, Attachment{ID: uuidString(row.ID), Name: row.Name, Size: row.Size, Type: row.MediaType})
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return message, err
	}
	return q.UpdateMessageBody(ctx, db.UpdateMessageBodyParams{ID: message.ID, Body: raw})
}
