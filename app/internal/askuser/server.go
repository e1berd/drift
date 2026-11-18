package askuser

import (
	"bufio"
	"encoding/json"
	"io"
)

// Command is the argument that makes the drift binary serve the ask_user tool over stdio MCP.
const Command = "mcp-ask"

const (
	protocolVersion = "2025-06-18"
	shownToUser     = "The questions are shown to the user. End your turn now without doing more work: the answers arrive as the user's next message."

	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
)

const toolDescription = `Ask the user one to four multiple-choice questions when a decision is theirs to make: an ambiguous requirement, a choice between approaches, a missing preference. Each question offers short options; the user can always type their own answer instead. After calling this tool, end your turn: the answers come back as the user's next message.`

var inputSchema = json.RawMessage(`{
  "type": "object",
  "required": ["questions"],
  "additionalProperties": false,
  "properties": {
    "questions": {
      "type": "array", "minItems": 1, "maxItems": 4,
      "items": {
        "type": "object",
        "required": ["question", "options"],
        "additionalProperties": false,
        "properties": {
          "question": {"type": "string", "description": "The full question, ending with a question mark."},
          "header": {"type": "string", "description": "A label of at most 12 characters, e.g. \"Database\"."},
          "options": {
            "type": "array", "minItems": 2, "maxItems": 6,
            "items": {
              "type": "object",
              "required": ["label"],
              "additionalProperties": false,
              "properties": {
                "label": {"type": "string", "description": "One to five words."},
                "description": {"type": "string", "description": "What choosing this means."}
              }
            }
          },
          "multiSelect": {"type": "boolean", "description": "Allow choosing several options."}
        }
      }
    }
  }
}`)

type rpcRequest struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content []textContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// Serve answers MCP requests read from in until it is closed.
func Serve(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		var req rpcRequest
		if json.Unmarshal(scanner.Bytes(), &req) != nil || len(req.ID) == 0 {
			continue
		}
		resp := rpcResponse{JSONRPC: "2.0", ID: req.ID}
		resp.Result, resp.Error = handle(req)
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func handle(req rpcRequest) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		return initializeResult(req.Params), nil
	case "ping":
		return struct{}{}, nil
	case "tools/list":
		return map[string]any{"tools": []any{toolDefinition()}}, nil
	case "tools/call":
		return callTool(req.Params)
	}
	return nil, &rpcError{Code: codeMethodNotFound, Message: "method not found: " + req.Method}
}

func initializeResult(params json.RawMessage) map[string]any {
	var client struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	version := protocolVersion
	if json.Unmarshal(params, &client) == nil && client.ProtocolVersion != "" {
		version = client.ProtocolVersion
	}
	return map[string]any{
		"protocolVersion": version,
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]string{"name": ServerName, "version": "1"},
	}
}

func toolDefinition() map[string]any {
	return map[string]any{
		"name":        ToolName,
		"description": toolDescription,
		"inputSchema": inputSchema,
		"annotations": map[string]bool{"readOnlyHint": true, "destructiveHint": false, "openWorldHint": false},
	}
}

func callTool(params json.RawMessage) (any, *rpcError) {
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &call); err != nil || call.Name != ToolName {
		return nil, &rpcError{Code: codeInvalidParams, Message: "unknown tool"}
	}
	if _, err := ParseQuestions(call.Arguments); err != nil {
		return toolResult{Content: []textContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	return toolResult{Content: []textContent{{Type: "text", Text: shownToUser}}}, nil
}
