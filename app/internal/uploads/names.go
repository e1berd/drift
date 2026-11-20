package uploads

import (
	"crypto/rand"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

const maxNameRunes = 255

// AttachmentName keeps only the base name of a client file name, at most 255 characters.
func AttachmentName(raw string) string {
	name := strings.TrimSpace(path.Base(strings.ReplaceAll(strings.TrimSpace(raw), `\`, "/")))
	if name == "" || name == "." || name == "/" || name == ".." {
		return fallbackName
	}
	if utf8.RuneCountInString(name) > maxNameRunes {
		name = string([]rune(name)[:maxNameRunes])
	}
	return name
}

func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !strings.ContainsRune("0123456789abcdef", r) {
				return false
			}
		}
	}
	return true
}
