package kratos

import (
	"encoding/json"
	"testing"
)

func TestUIAccessors(t *testing.T) {
	var ui UI
	raw := `{"action":"/x","method":"POST","messages":[{"id":1,"text":"flow","type":"info"}],"nodes":[
		{"type":"input","group":"default","attributes":{"name":"csrf_token","type":"hidden","value":"tok"},"messages":[]},
		{"type":"input","group":"password","attributes":{"name":"password","type":"password"},"messages":[{"id":2,"text":"bad","type":"error"}]},
		{"type":"input","group":"passkey","attributes":{"name":"passkey_challenge","type":"hidden","value":"{}"},"messages":[]}
	]}`
	if err := json.Unmarshal([]byte(raw), &ui); err != nil {
		t.Fatal(err)
	}
	if got := ui.Value("csrf_token"); got != "tok" {
		t.Errorf("csrf_token = %q", got)
	}
	if got := ui.Value("missing"); got != "" {
		t.Errorf("missing = %q", got)
	}
	if _, ok := ui.Node("password"); !ok {
		t.Error("password node not found")
	}
	if got := len(ui.AllMessages()); got != 2 {
		t.Errorf("AllMessages() has %d entries, want 2", got)
	}
}
