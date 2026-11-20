package uploads

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const sniffBytes = 512

func detectMediaType(path, name string) string {
	sniffed := sniffMediaType(path)
	if !isGenericMediaType(sniffed) {
		return sniffed
	}
	if byExtension := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); byExtension != "" {
		return byExtension
	}
	return sniffed
}

func sniffMediaType(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return "application/octet-stream"
	}
	defer file.Close()
	head := make([]byte, sniffBytes)
	n, _ := io.ReadFull(file, head)
	return http.DetectContentType(head[:n])
}

func isGenericMediaType(mediaType string) bool {
	return strings.HasPrefix(mediaType, "application/octet-stream") || strings.HasPrefix(mediaType, "text/plain")
}
