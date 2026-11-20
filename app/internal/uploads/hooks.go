package uploads

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tus/tusd/v2/pkg/filestore"
	tus "github.com/tus/tusd/v2/pkg/handler"
)

var (
	errLengthRequired = tus.NewError("ERR_LENGTH_REQUIRED", "the file size must be known when the upload starts", http.StatusLengthRequired)
	errEmptyFile      = tus.NewError("ERR_EMPTY_FILE", "the file is empty", http.StatusBadRequest)
	errNoDiskSpace    = tus.NewError("ERR_NO_DISK_SPACE", "not enough free disk space on the server", http.StatusInsufficientStorage)
	errSignedOut      = tus.NewError("ERR_SIGNED_OUT", "not signed in", http.StatusUnauthorized)
)

func tooLarge(limit int64) tus.Error {
	return tus.NewError("ERR_TOO_LARGE", fmt.Sprintf("a file must be at most %d bytes", limit), http.StatusRequestEntityTooLarge)
}

func (s *Server) beforeCreate(hook tus.HookEvent) (tus.HTTPResponse, tus.FileInfoChanges, error) {
	owner := ownerFrom(hook.Context)
	upload := hook.Upload
	switch {
	case owner == "":
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, errSignedOut
	case upload.SizeIsDeferred:
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, errLengthRequired
	case upload.Size <= 0:
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, errEmptyFile
	}
	limit, err := s.limits.MaxUploadBytes(hook.Context)
	if err != nil {
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, err
	}
	if upload.Size > limit {
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, tooLarge(limit)
	}
	if !fitsOnDisk(s.store.root, upload.Size) {
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, errNoDiskSpace
	}
	id, err := newID()
	if err != nil {
		return tus.HTTPResponse{}, tus.FileInfoChanges{}, err
	}
	return tus.HTTPResponse{}, tus.FileInfoChanges{
		ID: id,
		MetaData: tus.MetaData{
			ownerMetadata: owner,
			nameMetadata:  AttachmentName(upload.MetaData[nameMetadata]),
			typeMetadata:  upload.MetaData[typeMetadata],
		},
	}, nil
}

func (s *Server) finish(hook tus.HookEvent) (tus.HTTPResponse, error) {
	upload := hook.Upload
	name := upload.MetaData[nameMetadata]
	rel := relativePath(upload.ID, name)
	if err := s.moveFinished(upload, rel); err != nil {
		return tus.HTTPResponse{}, fmt.Errorf("store upload %s: %w", upload.ID, err)
	}
	var id pgtype.UUID
	if err := id.Scan(upload.ID); err != nil {
		return tus.HTTPResponse{}, err
	}
	_, err := s.q.CreateAttachment(hook.Context, db.CreateAttachmentParams{
		ID: id, OwnerID: upload.MetaData[ownerMetadata], Name: name,
		MediaType: detectMediaType(s.store.Path(rel), name), Size: upload.Size,
		Path: pgtype.Text{String: rel, Valid: true},
	})
	if err != nil {
		_ = os.RemoveAll(filepath.Dir(s.store.Path(rel)))
		return tus.HTTPResponse{}, fmt.Errorf("save attachment %s: %w", upload.ID, err)
	}
	return tus.HTTPResponse{Header: tus.HTTPHeader{"Attachment-Id": upload.ID}}, nil
}

func (s *Server) moveFinished(upload tus.FileInfo, rel string) error {
	partial, info := upload.Storage[filestore.StorageKeyPath], upload.Storage[filestore.StorageKeyInfoPath]
	if partial == "" {
		return errors.New("upload has no file")
	}
	target := s.store.Path(rel)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	if err := os.Rename(partial, target); err != nil {
		return err
	}
	if info != "" {
		_ = os.Remove(info)
	}
	return nil
}
