// Package askuser lets an agent ask the user multiple-choice questions. Drift serves the
// ask_user tool to the agent CLIs over MCP; the tool call shows up in the agent's event stream,
// the chat renders it as a question card, and the answers go back as the next prompt.
package askuser

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	ServerName = "drift"
	ToolName   = "ask_user"

	maxQuestions = 4
	maxOptions   = 6
	maxTextRunes = 2000
)

// Option is one answer the agent offers.
type Option struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// Question is one question of an ask_user call. The user may always type an answer of their own.
type Question struct {
	Question    string   `json:"question"`
	Header      string   `json:"header,omitempty"`
	Options     []Option `json:"options,omitempty"`
	MultiSelect bool     `json:"multiSelect,omitempty"`
}

// Answer is the user's reply to the question with the same index.
type Answer struct {
	Selected []string `json:"selected,omitempty"`
	Custom   string   `json:"custom,omitempty"`
}

// Card states: pending until the user answers, dismissed when the next run starts without answers.
const (
	CardPending   = "pending"
	CardAnswered  = "answered"
	CardDismissed = "dismissed"
)

// Card is the body of a question message in the transcript.
type Card struct {
	Questions []Question `json:"questions"`
	Answers   []Answer   `json:"answers,omitempty"`
	Status    string     `json:"status"`
}

type toolInput struct {
	Questions []Question `json:"questions"`
}

// ParseQuestions reads and checks the arguments of an ask_user call.
func ParseQuestions(raw json.RawMessage) ([]Question, error) {
	var input toolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("arguments are not a JSON object with questions: %w", err)
	}
	if len(input.Questions) == 0 || len(input.Questions) > maxQuestions {
		return nil, fmt.Errorf("ask 1 to %d questions", maxQuestions)
	}
	for i := range input.Questions {
		if err := input.Questions[i].normalize(); err != nil {
			return nil, fmt.Errorf("question %d: %w", i+1, err)
		}
	}
	return input.Questions, nil
}

func (q *Question) normalize() error {
	q.Question = strings.TrimSpace(q.Question)
	q.Header = strings.TrimSpace(q.Header)
	if q.Question == "" || len([]rune(q.Question)) > maxTextRunes {
		return errors.New("question text is empty or too long")
	}
	if len(q.Options) > maxOptions {
		return fmt.Errorf("offer at most %d options", maxOptions)
	}
	seen := make(map[string]bool, len(q.Options))
	for i := range q.Options {
		option := &q.Options[i]
		option.Label = strings.TrimSpace(option.Label)
		option.Description = strings.TrimSpace(option.Description)
		if option.Label == "" || seen[option.Label] {
			return errors.New("option labels must be unique and not empty")
		}
		seen[option.Label] = true
	}
	return nil
}

// CheckAnswers verifies there is one answer per question, built from offered options or own text.
func CheckAnswers(questions []Question, answers []Answer) error {
	if len(answers) != len(questions) {
		return errors.New("answer every question")
	}
	for i, answer := range answers {
		question := questions[i]
		custom := strings.TrimSpace(answer.Custom)
		if len(answer.Selected) == 0 && custom == "" {
			return fmt.Errorf("question %d has no answer", i+1)
		}
		if len([]rune(custom)) > maxTextRunes {
			return fmt.Errorf("answer %d is too long", i+1)
		}
		if !question.MultiSelect && len(answer.Selected) > 1 {
			return fmt.Errorf("question %d takes one option", i+1)
		}
		for _, label := range answer.Selected {
			if !question.offers(label) {
				return fmt.Errorf("question %d has no option %q", i+1, label)
			}
		}
	}
	return nil
}

func (q Question) offers(label string) bool {
	return slices.ContainsFunc(q.Options, func(option Option) bool { return option.Label == label })
}

// Prompt words the answers for the agent's next turn.
func Prompt(questions []Question, answers []Answer) string {
	var b strings.Builder
	b.WriteString("Answers to your questions:\n")
	for i, question := range questions {
		parts := slices.Clone(answers[i].Selected)
		if custom := strings.TrimSpace(answers[i].Custom); custom != "" {
			parts = append(parts, custom)
		}
		fmt.Fprintf(&b, "\n%s\n→ %s\n", question.Question, strings.Join(parts, "; "))
	}
	return strings.TrimRight(b.String(), "\n")
}
