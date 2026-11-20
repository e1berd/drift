package workspace

import (
	"context"
	"encoding/json"
	"mime"
	"net/http"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

const (
	maxMessageAttachments = 10
	defaultAttachmentName = "file"
)

var inlineMediaTypes = map[string]bool{
	"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true,
}

// Attachment is a file uploaded for a user message over tus (/uploads). It belongs to no message
// until a message or session created with its id claims it; unclaimed uploads are dropped after a week.
type Attachment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

type userMessageBody struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
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
	file, err := h.folders.Attachments.Open(row.Path.String)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "the file is gone")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	header := c.Response().Header()
	header.Set("Cache-Control", "private, max-age=31536000, immutable")
	header.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	header.Set("Content-Disposition", contentDisposition(row.Name, row.MediaType))
	header.Set("Content-Type", servedMediaType(row.MediaType))
	http.ServeContent(c.Response(), c.Request(), "", info.ModTime(), file)
	return nil
}

func contentDisposition(name, mediaType string) string {
	disposition := "attachment"
	if inlineMediaTypes[mediaType] {
		disposition = "inline"
	}
	return mime.FormatMediaType(disposition, map[string]string{"filename": name})
}

func servedMediaType(mediaType string) string {
	if inlineMediaTypes[mediaType] {
		return mediaType
	}
	return "application/octet-stream"
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
