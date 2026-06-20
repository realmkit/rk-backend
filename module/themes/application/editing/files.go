package editing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// ETag returns an HTTP ETag for one file.
func ETag(file domain.ThemeFile) string {
	return `"` + string(file.ContentSHA256) + `"`
}

// newFile creates a domain file from a write command.
func newFile(command WriteFileCommand) (domain.ThemeFile, error) {
	filePath, err := safePath(command.Path)
	if err != nil {
		return domain.ThemeFile{}, err
	}
	return domain.ThemeFile{
		ID:            uuid.New(),
		VersionID:     command.VersionID,
		Kind:          command.Kind,
		Path:          filePath,
		ContentText:   command.ContentText,
		ContentSHA256: digest(command.ContentText),
		SizeBytes:     int64(len(command.ContentText)),
	}, nil
}

// updateFile updates one file inside a file slice.
func updateFile(files []domain.ThemeFile, command WriteFileCommand) (domain.ThemeFile, error) {
	for index, file := range files {
		if file.ID != command.FileID {
			continue
		}
		if command.ExpectedETag != "" && command.ExpectedETag != ETag(file) {
			return domain.ThemeFile{}, port.ErrPreconditionFailed
		}
		next, err := newFile(command)
		if err != nil {
			return domain.ThemeFile{}, err
		}
		next.ID = file.ID
		files[index] = next
		return next, nil
	}
	return domain.ThemeFile{}, port.ErrNotFound
}

// deleteFile removes one file from a file slice.
func deleteFile(files []domain.ThemeFile, command DeleteFileCommand) ([]domain.ThemeFile, error) {
	for index, file := range files {
		if file.ID != command.FileID {
			continue
		}
		if command.ExpectedETag != "" && command.ExpectedETag != ETag(file) {
			return nil, port.ErrPreconditionFailed
		}
		return append(files[:index], files[index+1:]...), nil
	}
	return nil, port.ErrNotFound
}

// safePath validates and normalizes an editor path.
func safePath(filePath domain.FilePath) (domain.FilePath, error) {
	raw := strings.ReplaceAll(string(filePath), "\\", "/")
	if raw == "" || strings.HasPrefix(raw, "/") || strings.Contains(raw, "../") || raw == ".." {
		return "", fmt.Errorf("unsafe theme file path")
	}
	return domain.NormalizeFilePath(filePath), nil
}

// digest returns a lowercase SHA-256 digest.
func digest(content string) domain.Digest {
	hash := sha256.Sum256([]byte(content))
	return domain.Digest(hex.EncodeToString(hash[:]))
}
