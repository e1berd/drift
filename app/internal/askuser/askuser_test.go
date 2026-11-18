package askuser

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const sampleArguments = `{"questions":[
  {"question":" Which database? ","header":"DB","options":[{"label":"Postgres"},{"label":"SQLite","description":"single file"}]},
  {"question":"Which features?","options":[{"label":"Auth"},{"label":"Search"}],"multiSelect":true}
]}`

func TestParseQuestionsNormalizesText(t *testing.T) {
	questions, err := ParseQuestions(json.RawMessage(sampleArguments))
	if err != nil {
		t.Fatal(err)
	}
	if len(questions) != 2 || questions[0].Question != "Which database?" || !questions[1].MultiSelect {
		t.Fatalf("questions %+v", questions)
	}
}

func TestParseQuestionsRejectsBadInput(t *testing.T) {
	for _, raw := range []string{
		`{}`,
		`{"questions":[{"question":""}]}`,
		`{"questions":[{"question":"a?","options":[{"label":"x"},{"label":"x"}]}]}`,
		`not json`,
	} {
		if _, err := ParseQuestions(json.RawMessage(raw)); err == nil {
			t.Errorf("%s accepted", raw)
		}
	}
}

func TestCheckAnswers(t *testing.T) {
	questions, _ := ParseQuestions(json.RawMessage(sampleArguments))
	valid := []Answer{{Selected: []string{"Postgres"}}, {Selected: []string{"Auth", "Search"}, Custom: "and RSS"}}
	if err := CheckAnswers(questions, valid); err != nil {
		t.Errorf("valid answers refused: %v", err)
	}
	for _, answers := range [][]Answer{
		{{Selected: []string{"Postgres"}}},
		{{Selected: []string{"Postgres", "SQLite"}}, {Custom: "x"}},
		{{Selected: []string{"MySQL"}}, {Custom: "x"}},
		{{Custom: "  "}, {Custom: "x"}},
	} {
		if err := CheckAnswers(questions, answers); err == nil {
			t.Errorf("%+v accepted", answers)
		}
	}
}

func TestPromptListsEveryAnswer(t *testing.T) {
	questions, _ := ParseQuestions(json.RawMessage(sampleArguments))
	prompt := Prompt(questions, []Answer{{Custom: "MariaDB"}, {Selected: []string{"Auth", "Search"}}})
	want := "Answers to your questions:\n\nWhich database?\n→ MariaDB\n\nWhich features?\n→ Auth; Search"
	if prompt != want {
		t.Errorf("prompt %q", prompt)
	}
}

func compact(raw string) string {
	var out bytes.Buffer
	_ = json.Compact(&out, []byte(raw))
	return out.String()
}

func TestServeSpeaksMCP(t *testing.T) {
	in := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"ask_user","arguments":` + compact(sampleArguments) + `}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"ask_user","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"resources/list"}`,
	}, "\n")
	var out bytes.Buffer
	if err := Serve(strings.NewReader(in), &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d responses: %s", len(lines), out.String())
	}
	checks := []string{`"protocolVersion":"2025-03-26"`, `"name":"ask_user"`, shownToUser, `"isError":true`, `"code":-32601`}
	for i, want := range checks {
		if !strings.Contains(lines[i], want) {
			t.Errorf("response %d %s lacks %s", i, lines[i], want)
		}
	}
}
