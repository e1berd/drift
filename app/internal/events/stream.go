package events

import (
	"bufio"
	"encoding/json"
	"net/http"
	"time"

	"github.com/e1berd/drift/app/internal/identity"
	"github.com/labstack/echo/v5"
)

const heartbeat = 15 * time.Second

type wireEvent struct {
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data,omitempty"`
	Origin string          `json:"origin,omitempty"`
}

// Stream serves GET /api/events as text/event-stream for the signed-in user.
// Every change is a default "message" event whose JSON carries its type. Named "ready" (with the
// resume cursor as its id, after any replay) and periodic "ping" events let the client tell a live
// connection from a silently dead one.
func (h *Hub) Stream(c *echo.Context) error {
	r := c.Request()
	lastID := r.Header.Get("Last-Event-ID")
	if lastID == "" {
		lastID = c.QueryParam("last_event_id")
	}

	sub := h.Subscribe(identity.From(c), lastID)
	defer sub.Cancel()

	w := c.Response()
	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-store")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	bw := bufio.NewWriterSize(w, 4096)
	flush := func() error {
		if err := bw.Flush(); err != nil {
			return err
		}
		return rc.Flush()
	}

	if _, err := bw.WriteString("retry: 3000\n\n"); err != nil {
		return nil
	}
	if sub.Resync {
		if err := writeEvent(bw, "", wireEvent{Type: "resync"}); err != nil {
			return nil
		}
	}
	for _, ev := range sub.Missed {
		if err := writeEvent(bw, ev.ID, wireEvent{Type: ev.Type, Data: ev.Data, Origin: ev.Origin}); err != nil {
			return nil
		}
	}
	if _, err := bw.WriteString("id: " + sub.Cursor + "\nevent: ready\ndata: {}\n\n"); err != nil {
		return nil
	}
	if err := flush(); err != nil {
		return nil
	}

	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return nil
		case <-ticker.C:
			if _, err := bw.WriteString("event: ping\ndata: {}\n\n"); err != nil {
				return nil
			}
		case ev, ok := <-sub.Events:
			if !ok {
				return nil
			}
			if err := writeEvent(bw, ev.ID, wireEvent{Type: ev.Type, Data: ev.Data, Origin: ev.Origin}); err != nil {
				return nil
			}
		}
		if err := flush(); err != nil {
			return nil
		}
	}
}

func writeEvent(w *bufio.Writer, id string, ev wireEvent) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	if id != "" {
		if _, err := w.WriteString("id: " + id + "\n"); err != nil {
			return err
		}
	}
	if _, err := w.WriteString("data: "); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	_, err = w.WriteString("\n\n")
	return err
}
