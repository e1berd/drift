package workspace

import (
	"encoding/json"
	"time"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/typeid"
)

// Project is the API view of a project (a directory under the workspace root).
type Project struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	SessionCount int64     `json:"sessionCount"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

// Session is one agent conversation inside a project.
type Session struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"projectId,omitempty"`
	Title      string    `json:"title"`
	Agent      string    `json:"agent"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

// Message is one entry of a session transcript. Body depends on Kind.
type Message struct {
	ID         string          `json:"id"`
	SessionID  string          `json:"sessionId"`
	Role       string          `json:"role"`
	Kind       string          `json:"kind"`
	Body       json.RawMessage `json:"body"`
	CreateTime time.Time       `json:"createTime"`
}

// Draft is unsent composer text, keyed by "session:<id>", "new:<project id>" or "new:chat".
type Draft struct {
	Key        string    `json:"key"`
	Text       string    `json:"text"`
	DeviceID   string    `json:"deviceId"`
	UpdateTime time.Time `json:"updateTime"`
}

type deleted struct {
	ID        string `json:"id,omitempty"`
	ProjectID string `json:"projectId,omitempty"`
	Key       string `json:"key,omitempty"`
}

func projectDTO(p db.Project, sessions int64) Project {
	return Project{
		ID: uuidString(p.ID), Name: p.Name, Path: p.Path, SessionCount: sessions,
		CreateTime: p.CreatedAt.Time, UpdateTime: p.UpdatedAt.Time,
	}
}

func sessionDTO(s db.Session) Session {
	return Session{
		ID: uuidString(s.ID), ProjectID: uuidString(s.ProjectID), Title: s.Title, Agent: s.Agent,
		Status: s.Status, CreateTime: s.CreatedAt.Time, UpdateTime: s.UpdatedAt.Time,
	}
}

func messageDTO(m db.Message) Message {
	return Message{
		ID: typeid.MessageID(m.ID), SessionID: uuidString(m.SessionID), Role: m.Role, Kind: m.Kind,
		Body: m.Body, CreateTime: m.CreatedAt.Time,
	}
}

func draftDTO(d db.Draft) Draft {
	return Draft{Key: d.Key, Text: d.Text, DeviceID: d.DeviceID, UpdateTime: d.UpdatedAt.Time}
}
