package kratos

import "encoding/json"

// Identity is the subset of a Kratos identity the application reads.
type Identity struct {
	ID             string          `json:"id"`
	Traits         Traits          `json:"traits"`
	MetadataPublic json.RawMessage `json:"metadata_public,omitempty"`
}

// Traits mirrors deploy/kratos/identity.schema.json.
type Traits struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Session is the result of /sessions/whoami.
type Session struct {
	ID       string   `json:"id"`
	Active   bool     `json:"active"`
	Identity Identity `json:"identity"`
}

// Flow is a self-service browser flow (login, settings, …) as returned by Kratos.
type Flow struct {
	ID       string `json:"id"`
	Refresh  bool   `json:"refresh"`
	ReturnTo string `json:"return_to"`
	UI       UI     `json:"ui"`
}

// UI is the form description Kratos attaches to every flow.
type UI struct {
	Action   string    `json:"action"`
	Method   string    `json:"method"`
	Nodes    []Node    `json:"nodes"`
	Messages []Message `json:"messages"`
}

// Node is one form element of a flow.
type Node struct {
	Type       string         `json:"type"`
	Group      string         `json:"group"`
	Attributes NodeAttributes `json:"attributes"`
	Messages   []Message      `json:"messages"`
}

// NodeAttributes covers the attributes of input nodes.
type NodeAttributes struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    any    `json:"value"`
	Required bool   `json:"required"`
}

// Message is a user-facing message attached to a flow or node.
type Message struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Type string `json:"type"`
}

// Node returns the first node with the given attribute name.
func (u UI) Node(name string) (Node, bool) {
	for _, n := range u.Nodes {
		if n.Attributes.Name == name {
			return n, true
		}
	}
	return Node{}, false
}

// Value returns the string value of the named node, or "" when it is absent.
func (u UI) Value(name string) string {
	n, ok := u.Node(name)
	if !ok {
		return ""
	}
	s, _ := n.Attributes.Value.(string)
	return s
}

// AllMessages returns flow-level messages followed by every node message.
func (u UI) AllMessages() []Message {
	out := append([]Message(nil), u.Messages...)
	for _, n := range u.Nodes {
		out = append(out, n.Messages...)
	}
	return out
}

// FlowError is a self-service error fetched from /self-service/errors.
type FlowError struct {
	ID    string `json:"id"`
	Error struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
	} `json:"error"`
}
