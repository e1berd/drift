package runner

import (
	"context"
	"os"
	"strings"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/workspace"
)

const maxInlineImageBytes = 5 << 20

var inlineImageTypes = map[string]bool{
	"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true,
}

type attachedFile struct {
	name      string
	mediaType string
	size      int64
	path      string
}

type attachedImage struct {
	mediaType string
	data      []byte
}

func (r *Runner) loadAttachments(req workspace.RunRequest) ([]attachedFile, error) {
	if len(req.Attachments) == 0 {
		return nil, nil
	}
	rows, err := r.q.ListAttachmentData(context.WithoutCancel(r.ctx), db.ListAttachmentDataParams{OwnerID: req.Owner, Ids: req.Attachments})
	if err != nil {
		return nil, err
	}
	files := make([]attachedFile, 0, len(rows))
	for _, row := range rows {
		if row.Path.Valid {
			files = append(files, attachedFile{name: row.Name, mediaType: row.MediaType, size: row.Size, path: r.folders.Attachments.Path(row.Path.String)})
		}
	}
	return files, nil
}

func splitInlineImages(files []attachedFile) (images, others []attachedFile) {
	for _, file := range files {
		if inlineImageTypes[file.mediaType] && file.size <= maxInlineImageBytes {
			images = append(images, file)
		} else {
			others = append(others, file)
		}
	}
	return images, others
}

func readImages(files []attachedFile) ([]attachedImage, error) {
	images := make([]attachedImage, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file.path)
		if err != nil {
			return nil, err
		}
		images = append(images, attachedImage{mediaType: file.mediaType, data: data})
	}
	return images, nil
}

func paths(files []attachedFile) []string {
	result := make([]string, 0, len(files))
	for _, file := range files {
		result = append(result, file.path)
	}
	return result
}

func withFilePaths(text string, files []attachedFile) string {
	if len(files) == 0 {
		return text
	}
	lines := append([]string{"Attached files:"}, paths(files)...)
	return strings.TrimSpace(text + "\n\n" + strings.Join(lines, "\n"))
}
