package agents

import (
	"errors"
	"sync"
)

// Login is a sign-in running in the background until the user finishes it in the browser.
// URL is the page to open; UserCode is the code to enter there (device flow); NeedsCode means
// the browser shows a code the user pastes back into Drift (Claude flow).
type Login struct {
	URL       string
	UserCode  string
	NeedsCode bool

	submit   func(code string) error
	stop     func()
	once     sync.Once
	finished chan struct{}
	result   error
}

func newLogin(stop func()) *Login {
	return &Login{stop: stop, finished: make(chan struct{})}
}

func (l *Login) finish(err error) {
	l.once.Do(func() {
		l.result = err
		close(l.finished)
	})
}

// Finished is closed when the login ends; Result is valid after that.
func (l *Login) Finished() <-chan struct{} { return l.finished }

// Result is the outcome after Finished is closed: nil when the subscription is signed in.
func (l *Login) Result() error { return l.result }

// Submit passes the code from the browser to the CLI.
func (l *Login) Submit(code string) error {
	if l.submit == nil {
		return errors.New("this sign-in does not take a code")
	}
	return l.submit(code)
}

// Cancel stops the sign-in and its process.
func (l *Login) Cancel() {
	l.finish(errors.New("sign-in cancelled"))
	l.stop()
}
