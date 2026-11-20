package workspace

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/gitrepo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

// NoteRemoved is the code of the system note written when a file or worktree is removed from the info bar.
const NoteRemoved = "removed"

// RemovalNote is the body of that note; Delivered turns true once the agent has been told about it.
type RemovalNote struct {
	Code      string `json:"code"`
	Target    string `json:"target"`
	Path      string `json:"path"`
	Branch    string `json:"branch,omitempty"`
	SourceID  string `json:"sourceId"`
	Delivered bool   `json:"delivered"`
}

type activityTool struct {
	Name    string `json:"name"`
	Input   string `json:"input"`
	Created bool   `json:"created"`
}

// ActivityJournal reads and writes what the info bar shows about a session; the session id must
// already be checked against its owner, as SessionFolder does.
type ActivityJournal struct{ h *Handler }

// ActivityJournal returns the journal the git handler records removals in.
func (h *Handler) ActivityJournal() ActivityJournal { return ActivityJournal{h: h} }

// History returns the session's base commit, git and file tool calls and removals, oldest first.
func (j ActivityJournal) History(ctx context.Context, session string) (gitrepo.History, error) {
	id, err := parseSessionID(session)
	if err != nil {
		return gitrepo.History{}, err
	}
	base, err := j.h.q.GetSessionGitBase(ctx, id)
	if err != nil {
		return gitrepo.History{}, notFound(err)
	}
	rows, err := j.h.q.ListSessionActivity(ctx, id)
	if err != nil {
		return gitrepo.History{}, err
	}
	history := gitrepo.History{Base: base}
	for _, row := range rows {
		readActivityRow(&history, row)
	}
	return history, nil
}

func readActivityRow(history *gitrepo.History, row db.ListSessionActivityRow) {
	if row.Kind == "note" {
		var note RemovalNote
		if json.Unmarshal(row.Body, &note) == nil {
			history.Removals = append(history.Removals, gitrepo.Removal{
				SourceID: note.SourceID, Target: note.Target, Name: note.Path, Branch: note.Branch, Time: row.CreatedAt.Time,
			})
		}
		return
	}
	var tool activityTool
	if json.Unmarshal(row.Body, &tool) == nil {
		history.Calls = append(history.Calls, gitrepo.ToolCall{
			MessageID: uuidString(row.ID), Name: tool.Name, Input: tool.Input, Created: tool.Created,
		})
	}
}

// RecordRemoval writes the removal into the transcript as a system note and pushes it to the owner's clients.
func (j ActivityJournal) RecordRemoval(ctx context.Context, owner, session string, removal gitrepo.Removal) error {
	id, err := parseSessionID(session)
	if err != nil {
		return err
	}
	body, err := json.Marshal(RemovalNote{
		Code: NoteRemoved, Target: removal.Target, Path: removal.Name, Branch: removal.Branch, SourceID: removal.SourceID,
	})
	if err != nil {
		return err
	}
	m, err := j.h.q.CreateMessage(ctx, db.CreateMessageParams{SessionID: id, Role: "system", Kind: "note", Body: body})
	if err != nil {
		return err
	}
	j.h.hub.Publish(owner, "message.created", MessageOf(m), "")
	return nil
}

func parseSessionID(session string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(session); err != nil {
		return id, echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	return id, nil
}
