package auth

import "testing"

func TestSafeReturnTo(t *testing.T) {
	cases := map[string]string{
		"":                     "/",
		"/settings":            "/settings",
		"/settings?tab=keys":   "/settings?tab=keys",
		"https://evil.example": "/",
		"//evil.example/x":     "/",
		"/\\evil.example":      "/",
		"settings":             "/",
		"javascript:alert(1)":  "/",
	}
	for in, want := range cases {
		if got := safeReturnTo(in); got != want {
			t.Errorf("safeReturnTo(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateSetup(t *testing.T) {
	const pw = "correct-horse"
	cases := []struct {
		name, email, password, confirm string
		ok                             bool
	}{
		{"Root", "root@example.com", pw, pw, true},
		{"", "root@example.com", pw, pw, true},
		{"", "", pw, pw, false},
		{"", "Root <root@example.com>", pw, pw, false},
		{"", "not-an-email", pw, pw, false},
		{"", "root@example.com", "short", "short", false},
		{"", "root@example.com", pw, pw + "x", false},
	}
	for _, c := range cases {
		msg := validateSetup(c.name, c.email, c.password, c.confirm)
		if (msg == "") != c.ok {
			t.Errorf("validateSetup(%q, %q, …) = %q, want ok=%v", c.name, c.email, msg, c.ok)
		}
	}
}
