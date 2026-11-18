package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestFoldersGiveStandaloneSessionsTheirOwnFolder(t *testing.T) {
	folders := Folders{Workspace: t.TempDir(), Chats: filepath.Join(t.TempDir(), ".chats")}
	var session pgtype.UUID
	if err := session.Scan("0b6f3f1e-8c1a-4d55-9a57-3d0f2f3f8a10"); err != nil {
		t.Fatal(err)
	}

	dir, err := folders.Session(pgtype.Text{}, session)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(folders.Chats, "0b6f3f1e-8c1a-4d55-9a57-3d0f2f3f8a10"); dir != want {
		t.Fatalf("chat folder %q, want %q", dir, want)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("chat folder was not created: %v", err)
	}

	if err := folders.RemoveChat(session); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("chat folder still exists: %v", err)
	}
	if err := folders.RemoveChat(pgtype.UUID{}); err == nil {
		t.Fatal("removing a chat without an id must fail")
	}
}

func TestFoldersKeepProjectSessionsInsideTheWorkspace(t *testing.T) {
	folders := Folders{Workspace: "/ws", Chats: "/chats"}
	dir, err := folders.Session(pgtype.Text{String: "alpha", Valid: true}, pgtype.UUID{})
	if err != nil || dir != "/ws/alpha" {
		t.Fatalf("got %q, %v", dir, err)
	}
	if _, err := folders.Session(pgtype.Text{String: "../etc", Valid: true}, pgtype.UUID{}); err == nil {
		t.Fatal("a path outside the workspace must be refused")
	}
}
