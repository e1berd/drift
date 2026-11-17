package workspace

import (
	"net/http"
	"net/url"
	"regexp"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/labstack/echo/v5"
)

const maxDraftBytes = 64 << 10

var draftKey = regexp.MustCompile(`^(session|new):[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type listDraftsResponse struct {
	Drafts []Draft `json:"drafts"`
}

func (h *Handler) listDrafts(c *echo.Context) error {
	rows, err := h.q.ListDrafts(c.Request().Context(), identity.From(c))
	if err != nil {
		return err
	}
	out := listDraftsResponse{Drafts: make([]Draft, 0, len(rows))}
	for _, d := range rows {
		out.Drafts = append(out.Drafts, draftDTO(d))
	}
	return c.JSON(http.StatusOK, out)
}

type draftInput struct {
	Text string `json:"text"`
}

func pathKey(c *echo.Context) (string, error) {
	key, err := url.PathUnescape(c.Param("key"))
	if err != nil || !draftKey.MatchString(key) {
		return "", badRequest("invalid draft key")
	}
	return key, nil
}

func (h *Handler) putDraft(c *echo.Context) error {
	key, err := pathKey(c)
	if err != nil {
		return err
	}
	var in draftInput
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	if len(in.Text) > maxDraftBytes {
		return badRequest("draft is too long")
	}
	if in.Text == "" {
		return h.deleteDraft(c)
	}
	d, err := h.q.UpsertDraft(c.Request().Context(), db.UpsertDraftParams{
		OwnerID: identity.From(c), Key: key, Text: in.Text, DeviceID: device(c),
	})
	if err != nil {
		return err
	}
	dto := draftDTO(d)
	h.publish(c, "draft.updated", dto)
	return c.JSON(http.StatusOK, dto)
}

func (h *Handler) deleteDraft(c *echo.Context) error {
	key, err := pathKey(c)
	if err != nil {
		return err
	}
	n, err := h.q.DeleteDraft(c.Request().Context(), db.DeleteDraftParams{OwnerID: identity.From(c), Key: key})
	if err != nil {
		return err
	}
	if n > 0 {
		h.publish(c, "draft.deleted", deleted{Key: key})
	}
	return c.NoContent(http.StatusNoContent)
}
