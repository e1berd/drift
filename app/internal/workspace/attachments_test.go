package workspace

import (
	"strings"
	"testing"
)

func TestParseAttachmentIDs(t *testing.T) {
	id := "6f1c1d1e-8a5b-4c1b-9d2e-0f1a2b3c4d5e"
	ids, err := parseAttachmentIDs([]string{id, id})
	if err != nil || len(ids) != 1 {
		t.Fatalf("ids %v, err %v", ids, err)
	}
	if _, err := parseAttachmentIDs([]string{"nope"}); err == nil {
		t.Error("invalid id accepted")
	}
	if _, err := parseAttachmentIDs(strings.Split(strings.Repeat(id+",", 11), ",")[:11]); err == nil {
		t.Error("more than 10 attachments accepted")
	}
}
