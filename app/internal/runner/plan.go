package runner

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/e1berd/drift/app/internal/workspace"
)

var runnable = map[string]bool{"claude": true, "codex": true, "antigravity": true, "kimi": true, "mimo": true}

type runPlan struct {
	bin     string
	args    []string
	stdin   string
	stream  agentStream
	cleanup func()
}

func (r *Runner) plan(agent string, req workspace.RunRequest, agentSessionID, dir string, t transcriptSink, saveSessionID func(string)) (runPlan, error) {
	files, err := r.loadAttachments(req)
	if err != nil {
		return runPlan{}, fmt.Errorf("load attachments: %w", err)
	}
	switch agent {
	case "codex":
		images, others := splitInlineImages(files)
		return runPlan{
			bin:     r.cli.Codex,
			args:    codexArgs(req, agentSessionID, paths(images), r.askMCP),
			stdin:   withFilePaths(req.Text, others),
			stream:  &codexStream{t: t, projectDir: dir, threadID: agentSessionID, saveSessionID: saveSessionID},
			cleanup: func() {},
		}, nil
	case "antigravity":
		req.Text = withFilePaths(req.Text, files)
		return runPlan{
			bin:     r.cli.Antigravity,
			args:    antigravityArgs(req, agentSessionID),
			stream:  &antigravityStream{t: t, projectDir: dir, conversationID: agentSessionID, saveSessionID: saveSessionID},
			cleanup: func() {},
		}, nil
	case "kimi":
		req.Text = withFilePaths(req.Text, files)
		return runPlan{
			bin:     r.cli.KimiPath(),
			args:    kimiArgs(req, agentSessionID),
			stream:  &kimiStream{t: t, projectDir: dir, saveSessionID: saveSessionID},
			cleanup: func() {},
		}, nil
	case "mimo":
		images, others := splitInlineImages(files)
		return runPlan{
			bin:     r.cli.MimoPath(),
			args:    mimoArgs(req, agentSessionID, paths(images)),
			stdin:   withFilePaths(req.Text, others),
			stream:  &mimoStream{t: t, projectDir: dir, saveSessionID: saveSessionID},
			cleanup: func() {},
		}, nil
	}
	images, others := splitInlineImages(files)
	inline, err := readImages(images)
	if err != nil {
		return runPlan{}, fmt.Errorf("read attachments: %w", err)
	}
	return runPlan{
		bin:     r.cli.Claude,
		args:    claudeArgs(req, agentSessionID, len(inline) > 0, r.askMCP),
		stdin:   claudeInput(withFilePaths(req.Text, others), inline),
		stream:  &claudeStream{t: t, projectDir: dir, agentSessionID: agentSessionID, saveSessionID: saveSessionID},
		cleanup: func() {},
	}, nil
}

func claudeArgs(req workspace.RunRequest, agentSessionID string, streamInput bool, askMCP mcpCommand) []string {
	args := []string{"-p", "--output-format", "stream-json", "--verbose", "--include-partial-messages"}
	args = append(args, askMCP.claudeArgs()...)
	if streamInput {
		args = append(args, "--input-format", "stream-json")
	}
	if req.Model != "" && req.Model != "default" {
		args = append(args, "--model", req.Model)
	}
	if req.Mode != "" && req.Mode != "default" {
		args = append(args, "--permission-mode", req.Mode)
	}
	if agentSessionID != "" {
		args = append(args, "--resume", agentSessionID)
	}
	return args
}

type claudeInputBlock struct {
	Type   string             `json:"type"`
	Text   string             `json:"text,omitempty"`
	Source *claudeImageSource `json:"source,omitempty"`
}

type claudeImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type claudeUserMessage struct {
	Type            string  `json:"type"`
	SessionID       string  `json:"session_id"`
	ParentToolUseID *string `json:"parent_tool_use_id"`
	Message         struct {
		Role    string             `json:"role"`
		Content []claudeInputBlock `json:"content"`
	} `json:"message"`
}

func claudeInput(text string, images []attachedImage) string {
	if len(images) == 0 {
		return text
	}
	message := claudeUserMessage{Type: "user"}
	message.Message.Role = "user"
	if text != "" {
		message.Message.Content = append(message.Message.Content, claudeInputBlock{Type: "text", Text: text})
	}
	for _, image := range images {
		message.Message.Content = append(message.Message.Content, claudeInputBlock{
			Type:   "image",
			Source: &claudeImageSource{Type: "base64", MediaType: image.mediaType, Data: base64.StdEncoding.EncodeToString(image.data)},
		})
	}
	line, _ := json.Marshal(message)
	return string(line) + "\n"
}
