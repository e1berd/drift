package gitrepo

import (
	"encoding/json"
	"path"
	"strings"
)

// ToolCall is one tool use of the agent, as stored in the session's transcript.
type ToolCall struct {
	MessageID string
	Name      string
	Input     string
}

// CreatedRef is a branch or a worktree that a tool call of the session created.
type CreatedRef struct {
	MessageID string
	Branch    string
	Worktree  string
	Dir       string
}

var commandSeparators = []string{"&&", "||", ";", "|", "\n"}

var branchKeepFlags = map[string]bool{
	"-f": true, "--force": true, "-t": true, "--track": true, "--no-track": true, "-q": true, "--quiet": true,
}

var worktreeValueFlags = map[string]bool{"--reason": true}

// CreatedRefs reads the agent's tool calls and returns, in order, every branch and worktree they created.
func CreatedRefs(calls []ToolCall) []CreatedRef {
	refs := make([]CreatedRef, 0, 4)
	for _, call := range calls {
		switch call.Name {
		case "Bash":
			refs = append(refs, refsOfShell(call.MessageID, call.Input)...)
		case "EnterWorktree":
			if name := enteredWorktreeName(call.Input); name != "" {
				refs = append(refs, CreatedRef{MessageID: call.MessageID, Worktree: name})
			}
		}
	}
	return refs
}

func enteredWorktreeName(input string) string {
	var fields struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if json.Unmarshal([]byte(input), &fields) != nil {
		return strings.TrimSpace(input)
	}
	if fields.Path != "" {
		return fields.Path
	}
	return fields.Name
}

func refsOfShell(messageID, script string) []CreatedRef {
	var refs []CreatedRef
	dir := ""
	for _, words := range simpleCommands(script) {
		words = withoutAssignments(words)
		if len(words) == 0 {
			continue
		}
		if words[0] == "cd" && len(words) > 1 {
			dir = joinDir(dir, words[1])
			continue
		}
		if path.Base(words[0]) != "git" {
			continue
		}
		gitDir, args := gitGlobalOptions(words[1:])
		for _, ref := range refsOfGit(args) {
			ref.MessageID = messageID
			ref.Dir = joinDir(dir, gitDir)
			refs = append(refs, ref)
		}
	}
	return refs
}

func joinDir(base, next string) string {
	switch {
	case next == "":
		return base
	case path.IsAbs(next) || strings.HasPrefix(next, "~"):
		return next
	case base == "":
		return path.Clean(next)
	}
	return path.Join(base, next)
}

func withoutAssignments(words []string) []string {
	for len(words) > 0 && isAssignment(words[0]) {
		words = words[1:]
	}
	return words
}

func isAssignment(word string) bool {
	name, _, found := strings.Cut(word, "=")
	return found && name != "" && !strings.ContainsAny(name, "-/.")
}

func gitGlobalOptions(args []string) (dir string, rest []string) {
	for len(args) > 0 {
		switch {
		case args[0] == "-C" && len(args) > 1:
			dir = joinDir(dir, args[1])
			args = args[2:]
		case args[0] == "-c" && len(args) > 1:
			args = args[2:]
		case strings.HasPrefix(args[0], "-"):
			args = args[1:]
		default:
			return dir, args
		}
	}
	return dir, nil
}

func refsOfGit(args []string) []CreatedRef {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "checkout":
		return branchAfterFlag(args[1:], "-b", "-B", "--orphan")
	case "switch":
		return branchAfterFlag(args[1:], "-c", "-C", "--create", "--force-create", "--orphan")
	case "branch":
		return createdBranch(args[1:])
	case "worktree":
		if len(args) > 1 && args[1] == "add" {
			return addedWorktree(args[2:])
		}
	}
	return nil
}

func branchAfterFlag(args []string, flags ...string) []CreatedRef {
	for i, arg := range args {
		for _, flag := range flags {
			if arg == flag && i+1 < len(args) {
				return []CreatedRef{{Branch: args[i+1]}}
			}
			if value, ok := strings.CutPrefix(arg, flag+"="); ok && strings.HasPrefix(flag, "--") {
				return []CreatedRef{{Branch: value}}
			}
		}
	}
	return nil
}

func createdBranch(args []string) []CreatedRef {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if !branchKeepFlags[arg] {
				return nil
			}
			continue
		}
		return []CreatedRef{{Branch: arg}}
	}
	return nil
}

func addedWorktree(args []string) []CreatedRef {
	var branch string
	var positional []string
	detached := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case (arg == "-b" || arg == "-B" || arg == "--orphan") && i+1 < len(args):
			branch = args[i+1]
			i++
		case arg == "--detach" || arg == "-d":
			detached = true
		case worktreeValueFlags[arg]:
			i++
		case strings.HasPrefix(arg, "-"):
		default:
			positional = append(positional, arg)
		}
	}
	if len(positional) == 0 {
		return nil
	}
	if branch == "" && !detached && len(positional) == 1 {
		branch = path.Base(positional[0])
	}
	return []CreatedRef{{Worktree: positional[0], Branch: branch}}
}

func simpleCommands(script string) [][]string {
	var commands [][]string
	var words []string
	var word strings.Builder
	inWord := false
	var quote rune
	flushWord := func() {
		if inWord {
			words = append(words, word.String())
			word.Reset()
			inWord = false
		}
	}
	flushCommand := func() {
		flushWord()
		if len(words) > 0 {
			commands = append(commands, words)
			words = nil
		}
	}
	runes := []rune(script)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				word.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			inWord = true
		case r == '\\' && i+1 < len(runes):
			i++
			if runes[i] != '\n' {
				word.WriteRune(runes[i])
				inWord = true
			}
		case r == ' ' || r == '\t' || r == '(' || r == ')':
			flushWord()
		default:
			if separator := separatorAt(runes, i); separator != "" {
				flushCommand()
				i += len([]rune(separator)) - 1
				continue
			}
			word.WriteRune(r)
			inWord = true
		}
	}
	flushCommand()
	return commands
}

func separatorAt(runes []rune, i int) string {
	rest := string(runes[i:min(len(runes), i+2)])
	for _, separator := range commandSeparators {
		if strings.HasPrefix(rest, separator) {
			return separator
		}
	}
	return ""
}
