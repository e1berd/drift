package runner

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/e1berd/drift/app/internal/askuser"
)

type mcpCommand struct {
	path string
	args []string
}

func askMCPCommand() (mcpCommand, error) {
	path, err := os.Executable()
	return mcpCommand{path: path, args: []string{askuser.Command}}, err
}

func (c mcpCommand) claudeArgs() []string {
	config, _ := json.Marshal(map[string]any{
		"mcpServers": map[string]any{askuser.ServerName: map[string]any{"command": c.path, "args": c.args}},
	})
	return []string{
		"--mcp-config", string(config),
		"--allowedTools", claudeAskTool,
		"--disallowedTools", "AskUserQuestion",
	}
}

func (c mcpCommand) codexArgs() []string {
	quoted := make([]string, len(c.args))
	for i, arg := range c.args {
		quoted[i] = strconv.Quote(arg)
	}
	prefix := "mcp_servers." + askuser.ServerName + "."
	return []string{
		"-c", prefix + "command=" + strconv.Quote(c.path),
		"-c", prefix + "args=[" + strings.Join(quoted, ",") + "]",
		"-c", prefix + `default_tools_approval_mode="approve"`,
	}
}
