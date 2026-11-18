package workspace

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/e1berd/drift/app/internal/askuser"
	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

// ErrNotContinuable means the limit message is not in a state its continuation can change.
var ErrNotContinuable = errors.New("the limit can no longer be continued this way")

type answerInput struct {
	runOptions
	Answers []askuser.Answer `json:"answers"`
}

// messageCustomMethod serves AIP-136 custom methods of transcript cards:
// POST /sessions/{id}/messages/{id}:answer, :continue, :scheduleContinue, :cancelContinue.
func (h *Handler) messageCustomMethod(c *echo.Context) error {
	s, err := h.session(c)
	if err != nil {
		return err
	}
	raw, method, _ := strings.Cut(c.Param("message"), ":")
	var message pgtype.UUID
	if err := message.Scan(raw); err != nil || h.runner == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	owner := identity.From(c)
	switch method {
	case "answer":
		return h.answerQuestion(c, s, message)
	case "continue":
		return continuationResult(c, h.runner.ContinueNow(owner, s.ID, message))
	case "scheduleContinue":
		return continuationResult(c, h.runner.ScheduleContinue(owner, s.ID, message))
	case "cancelContinue":
		return continuationResult(c, h.runner.CancelContinue(owner, s.ID, message))
	}
	return echo.NewHTTPError(http.StatusNotFound, "not found")
}

func continuationResult(c *echo.Context, err error) error {
	if errors.Is(err, ErrNotContinuable) {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}
	if err != nil {
		return notFound(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) answerQuestion(c *echo.Context, s db.Session, id pgtype.UUID) error {
	var in answerInput
	if err := c.Bind(&in); err != nil {
		return badRequest("invalid body")
	}
	if err := in.validate(); err != nil {
		return err
	}
	ctx := c.Request().Context()
	m, err := h.q.GetSessionMessage(ctx, db.GetSessionMessageParams{ID: id, SessionID: s.ID})
	if err != nil {
		return notFound(err)
	}
	var body askuser.Card
	if m.Kind != "question" || json.Unmarshal(m.Body, &body) != nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if body.Status != askuser.CardPending {
		return echo.NewHTTPError(http.StatusConflict, "the question is already closed")
	}
	if err := askuser.CheckAnswers(body.Questions, in.Answers); err != nil {
		return badRequest(err.Error())
	}
	body.Answers, body.Status = in.Answers, askuser.CardAnswered
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	updated, err := h.q.UpdateMessageBody(ctx, db.UpdateMessageBodyParams{ID: m.ID, Body: raw})
	if err != nil {
		return err
	}
	out := messageDTO(updated)
	h.publish(c, "message.updated", out)
	h.startRun(c, s.ID, askuser.Prompt(body.Questions, body.Answers), nil, in.runOptions)
	return c.JSON(http.StatusOK, out)
}
