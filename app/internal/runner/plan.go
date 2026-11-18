package runner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/workspace"
)

var runnable = map[string]bool{"claude": true, "codex": true, "antigravity": true}

var imageExtensions = map[string]string{
	"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp",
}

type runPlan struct {
	bin     string
	args    []string
	stdin   string
	stream  agentStream
	cleanup func()
}

type attachedImage struct {
	name      string
	mediaType string
	data      []byte
}

func (r *Runner) plan(agent string, req workspace.RunRequest, agentSessionID, dir string, t transcriptSink, saveSessionID func(string)) (runPlan, error) {
	images, err := r.loadImages(req)
	if err != nil {
		return runPlan{}, fmt.Errorf("load attachments: %w", err)
	}
	switch agent {
	case "codex":
		folder, paths, err := writeImages(images)
		if err != nil {
			return runPlan{}, err
		}
		return runPlan{
			bin:     r.cli.Codex,
			args:    codexArgs(req, agentSessionID, paths),
			stdin:   req.Text,
			stream:  &codexStream{t: t, projectDir: dir, threadID: agentSessionID, saveSessionID: saveSessionID},
			cleanup: removeFolder(folder),
		}, nil
	case "antigravity":
		folder, paths, err := writeImages(images)
		if err != nil {
			return runPlan{}, err
		}
		return runPlan{
			bin:     r.cli.Antigravity,
			args:    antigravityArgs(withImagePaths(req, paths), agentSessionID),
			stream:  &antigravityStream{t: t, projectDir: dir, conversationID: agentSessionID, saveSessionID: saveSessionID},
			cleanup: removeFolder(folder),
		}, nil
	}
	return runPlan{
		bin:     r.cli.Claude,
		args:    claudeArgs(req, agentSessionID, len(images) > 0),
		stdin:   claudeInput(req.Text, images),
		stream:  &claudeStream{t: t, projectDir: dir, agentSessionID: agentSessionID, saveSessionID: saveSessionID},
		cleanup: func() {},
	}, nil
}

func (r *Runner) loadImages(req workspace.RunRequest) ([]attachedImage, error) {
	if len(req.Attachments) == 0 {
		return nil, nil
	}
	rows, err := r.q.ListAttachmentData(context.WithoutCancel(r.ctx), db.ListAttachmentDataParams{OwnerID: req.Owner, Ids: req.Attachments})
	if err != nil {
		return nil, err
	}
	images := make([]attachedImage, 0, len(rows))
	for _, row := range rows {
		images = append(images, attachedImage{name: row.Name, mediaType: row.MediaType, data: row.Data})
	}
	return images, nil
}

func writeImages(images []attachedImage) (string, []string, error) {
	if len(images) == 0 {
		return "", nil, nil
	}
	folder, err := os.MkdirTemp("", "drift-attachments-")
	if err != nil {
		return "", nil, fmt.Errorf("save attachments: %w", err)
	}
	paths := make([]string, 0, len(images))
	for i, image := range images {
		base := strings.TrimSuffix(image.name, filepath.Ext(image.name))
		path := filepath.Join(folder, fmt.Sprintf("%02d-%s%s", i+1, safeFileName(base), imageExtensions[image.mediaType]))
		if err := os.WriteFile(path, image.data, 0o600); err != nil {
			_ = os.RemoveAll(folder)
			return "", nil, fmt.Errorf("save attachments: %w", err)
		}
		paths = append(paths, path)
	}
	return folder, paths, nil
}

func safeFileName(name string) string {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			return r
		}
		return '_'
	}, name)
	if clean = strings.Trim(clean, "._"); clean == "" {
		return "image"
	}
	return clip(clean, 60)
}

func removeFolder(folder string) func() {
	return func() {
		if folder != "" {
			_ = os.RemoveAll(folder)
		}
	}
}

func withImagePaths(req workspace.RunRequest, paths []string) workspace.RunRequest {
	if len(paths) == 0 {
		return req
	}
	lines := append([]string{"Attached images:"}, paths...)
	req.Text = strings.TrimSpace(req.Text + "\n\n" + strings.Join(lines, "\n"))
	return req
}

func claudeArgs(req workspace.RunRequest, agentSessionID string, streamInput bool) []string {
	args := []string{"-p", "--output-format", "stream-json", "--verbose", "--include-partial-messages"}
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
