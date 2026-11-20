package events

import (
	"testing"
)

func drain(ch <-chan Event) []Event {
	var out []Event
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, ev)
		default:
			return out
		}
	}
}

func TestPublishReachesOnlyOwner(t *testing.T) {
	h := NewHub()
	alice := h.Subscribe("alice", "")
	defer alice.Cancel()
	bob := h.Subscribe("bob", "")
	defer bob.Cancel()

	h.Publish("alice", "draft.updated", map[string]string{"key": "k"}, "phone")

	got := drain(alice.Events)
	if len(got) != 1 || got[0].Type != "draft.updated" || got[0].Origin != "phone" {
		t.Fatalf("alice got %+v", got)
	}
	if other := drain(bob.Events); len(other) != 0 {
		t.Fatalf("bob got %+v", other)
	}
}

func TestReplayAfterLastEventID(t *testing.T) {
	h := NewHub()
	sub := h.Subscribe("alice", "")
	for i := 0; i < 3; i++ {
		h.Publish("alice", "session.created", i, "")
	}
	first := drain(sub.Events)[0]
	sub.Cancel()

	resumed := h.Subscribe("alice", first.ID)
	defer resumed.Cancel()
	if resumed.Resync {
		t.Fatal("unexpected resync")
	}
	if len(resumed.Missed) != 2 {
		t.Fatalf("want 2 missed events, got %d", len(resumed.Missed))
	}
}

func TestResyncForForeignBootOrGap(t *testing.T) {
	h := NewHub()
	foreign := h.Subscribe("alice", "oldboot-5")
	foreign.Cancel()
	if !foreign.Resync {
		t.Fatal("foreign boot id must resync")
	}

	for i := 0; i < ringSize+10; i++ {
		h.Publish("alice", "x", i, "")
	}
	behind := h.Subscribe("alice", h.boot+"-1")
	defer behind.Cancel()
	if !behind.Resync {
		t.Fatal("event outside the ring must resync")
	}
}

func TestSlowSubscriberIsDropped(t *testing.T) {
	h := NewHub()
	sub := h.Subscribe("alice", "")
	defer sub.Cancel()
	for i := 0; i < subBuffer+1; i++ {
		h.Publish("alice", "x", i, "")
	}
	n := 0
	for range sub.Events {
		n++
	}
	if n != subBuffer {
		t.Fatalf("want %d buffered events before close, got %d", subBuffer, n)
	}
}

func TestCursorResumesTabThatSawNoEvents(t *testing.T) {
	h := NewHub()
	h.Publish("bob", "x", 0, "")
	idle := h.Subscribe("alice", "")
	idle.Cancel()

	h.Publish("alice", "session.created", 1, "")
	resumed := h.Subscribe("alice", idle.Cursor)
	defer resumed.Cancel()
	if resumed.Resync || len(resumed.Missed) != 1 || resumed.Missed[0].Type != "session.created" {
		t.Fatalf("want the one event published while away, got resync=%v missed=%+v", resumed.Resync, resumed.Missed)
	}
}
