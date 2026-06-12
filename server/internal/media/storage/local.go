package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Local struct {
	root string
}

func NewLocal(root string) (*Local, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &Local{root: root}, nil
}

func (l *Local) RunDir(runID string) string {
	return filepath.Join(l.root, "runs", runID)
}

func (l *Local) EnsureRunDir(runID string) (string, error) {
	dir := l.RunDir(runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func (l *Local) WriteFile(runID, name string, data []byte) (string, error) {
	dir, err := l.EnsureRunDir(runID)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (l *Local) PublicPath(runID, name string) string {
	return fmt.Sprintf("/media/runs/%s/%s", runID, name)
}

func (l *Local) ResolvePublicURL(runID, name string) string {
	return l.PublicPath(runID, name)
}

func (l *Local) FilePath(runID, name string) string {
	return filepath.Join(l.RunDir(runID), name)
}

func (l *Local) InputDir(runID string) string {
	return filepath.Join(l.RunDir(runID), "input")
}

func (l *Local) WriteInputFile(runID, assetType, ext string, data []byte) (string, string, error) {
	dir := l.InputDir(runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	if ext == "" {
		ext = ".bin"
	}
	if ext[0] != '.' {
		ext = "." + ext
	}
	name := assetType + ext
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", "", err
	}
	public := fmt.Sprintf("/media/runs/%s/input/%s", runID, name)
	return path, public, nil
}

func PublicToRelPath(publicURL string) string {
	const prefix = "/media/runs/"
	if !strings.HasPrefix(publicURL, prefix) {
		return publicURL
	}
	rest := publicURL[len(prefix):]
	if i := strings.Index(rest, "/"); i >= 0 {
		return rest[i+1:]
	}
	return rest
}
