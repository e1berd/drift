package workspace

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestSessionUpdateParams(t *testing.T) {
	text := func(s string) *string { return &s }
	cases := []struct {
		name    string
		in      sessionUpdate
		title   pgtype.Text
		agent   pgtype.Text
		wantErr bool
	}{
		{name: "empty", in: sessionUpdate{}, wantErr: true},
		{name: "title", in: sessionUpdate{Title: text("  Plan  ")}, title: pgtype.Text{String: "Plan", Valid: true}},
		{name: "blank title", in: sessionUpdate{Title: text("  ")}, wantErr: true},
		{name: "agent", in: sessionUpdate{Agent: text("codex")}, agent: pgtype.Text{String: "codex", Valid: true}},
		{name: "unknown agent", in: sessionUpdate{Agent: text("gpt")}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.in.params(pgtype.UUID{}, "owner")
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got.Title != tc.title || got.Agent != tc.agent || got.OwnerID != "owner" {
				t.Errorf("params = %+v", got)
			}
		})
	}
}
