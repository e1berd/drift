// Package events fans out per-user change events to every open tab over Server-Sent Events.
// Each user keeps a small ring of recent events, so a reconnecting tab replays what it missed
// by Last-Event-ID instead of refetching everything.
package events

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ringSize   = 256
	subBuffer  = 64
	maxIdleAge = 10 * time.Minute
)

// Event is one change notification. Origin is the device id that caused it, so that device
// can skip echoes of its own writes.
type Event struct {
	ID     string          `json:"-"`
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data"`
	Origin string          `json:"origin,omitempty"`
	seq    uint64
}

type stream struct {
	subs     map[chan Event]struct{}
	ring     []Event
	lastSeen time.Time
}

// Hub keeps the subscribers and recent events of every user in memory.
type Hub struct {
	mu    sync.Mutex
	boot  string
	seq   uint64
	users map[string]*stream
}

// NewHub returns an empty hub. Event ids carry a boot marker, so ids from a previous process
// are recognised and answered with a resync.
func NewHub() *Hub {
	return &Hub{boot: strconv.FormatInt(time.Now().UnixNano(), 36), users: make(map[string]*stream)}
}

// Publish records the event for the owner and sends it to every subscriber without blocking.
// A subscriber whose buffer is full is dropped; its tab reconnects and replays from the ring.
func (h *Hub) Publish(owner, typ string, data any, origin string) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.seq++
	ev := Event{ID: h.boot + "-" + strconv.FormatUint(h.seq, 10), Type: typ, Data: raw, Origin: origin, seq: h.seq}
	s := h.streamLocked(owner)
	s.ring = append(s.ring, ev)
	if len(s.ring) > ringSize {
		s.ring = s.ring[len(s.ring)-ringSize:]
	}
	for ch := range s.subs {
		select {
		case ch <- ev:
		default:
			delete(s.subs, ch)
			close(ch)
		}
	}
}

// Subscribe registers a subscriber and returns the events after lastID. resync is true when
// lastID cannot be replayed (older process, or too far behind) and the client must refetch.
func (h *Hub) Subscribe(owner, lastID string) (ch chan Event, missed []Event, resync bool, cancel func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	s := h.streamLocked(owner)
	ch = make(chan Event, subBuffer)
	s.subs[ch] = struct{}{}

	if lastID != "" {
		missed, resync = h.sinceLocked(s, lastID)
	}

	cancel = func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if _, ok := s.subs[ch]; ok {
			delete(s.subs, ch)
			close(ch)
		}
		s.lastSeen = time.Now()
	}
	return ch, missed, resync, cancel
}

func (h *Hub) sinceLocked(s *stream, lastID string) ([]Event, bool) {
	boot, seqStr, ok := strings.Cut(lastID, "-")
	if !ok || boot != h.boot {
		return nil, true
	}
	seq, err := strconv.ParseUint(seqStr, 10, 64)
	if err != nil {
		return nil, true
	}
	if seq >= h.seq {
		return nil, false
	}
	if len(s.ring) == 0 || s.ring[0].seq > seq+1 {
		return nil, true
	}
	var out []Event
	for _, ev := range s.ring {
		if ev.seq > seq {
			out = append(out, ev)
		}
	}
	return out, false
}

func (h *Hub) streamLocked(owner string) *stream {
	s, ok := h.users[owner]
	if !ok {
		h.pruneLocked()
		s = &stream{subs: make(map[chan Event]struct{})}
		h.users[owner] = s
	}
	s.lastSeen = time.Now()
	return s
}

func (h *Hub) pruneLocked() {
	cutoff := time.Now().Add(-maxIdleAge)
	for owner, s := range h.users {
		if len(s.subs) == 0 && s.lastSeen.Before(cutoff) {
			delete(h.users, owner)
		}
	}
}
