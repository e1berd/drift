package events

import (
	"testing"
)

func drain(ch chan Event) []Event {
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
	a, _, _, cancelA := h.Subscribe("alice", "")
	defer cancelA()
	b, _, _, cancelB := h.Subscribe("bob", "")
	defer cancelB()

	h.Publish("alice", "draft.updated", map[string]string{"key": "k"}, "phone")

	got := drain(a)
	if len(got) != 1 || got[0].Type != "draft.updated" || got[0].Origin != "phone" {
		t.Fatalf("alice got %+v", got)
	}
	if other := drain(b); len(other) != 0 {
		t.Fatalf("bob got %+v", other)
	}
}

func TestReplayAfterLastEventID(t *testing.T) {
	h := NewHub()
	ch, _, _, cancel := h.Subscribe("alice", "")
	for i := 0; i < 3; i++ {
		h.Publish("alice", "session.created", i, "")
	}
	first := drain(ch)[0]
	cancel()

	_, missed, resync, cancel2 := h.Subscribe("alice", first.ID)
	defer cancel2()
	if resync {
		t.Fatal("unexpected resync")
	}
	if len(missed) != 2 {
		t.Fatalf("want 2 missed events, got %d", len(missed))
	}
}

func TestResyncForForeignBootOrGap(t *testing.T) {
	h := NewHub()
	if _, _, resync, cancel := h.Subscribe("alice", "oldboot-5"); !resync {
		t.Fatal("foreign boot id must resync")
	} else {
		cancel()
	}

	for i := 0; i < ringSize+10; i++ {
		h.Publish("alice", "x", i, "")
	}
	_, _, resync, cancel := h.Subscribe("alice", h.boot+"-1")
	defer cancel()
	if !resync {
		t.Fatal("event outside the ring must resync")
	}
}

func TestSlowSubscriberIsDropped(t *testing.T) {
	h := NewHub()
	ch, _, _, cancel := h.Subscribe("alice", "")
	defer cancel()
	for i := 0; i < subBuffer+1; i++ {
		h.Publish("alice", "x", i, "")
	}
	n := 0
	for range ch {
		n++
	}
	if n != subBuffer {
		t.Fatalf("want %d buffered events before close, got %d", subBuffer, n)
	}
}
