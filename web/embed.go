package web

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:static
var staticFS embed.FS

// MaterializeStatic writes embedded static assets to a temp directory for
// hosts that ship only the binary (Vercel).
func MaterializeStatic() (string, error) {
	root, err := os.MkdirTemp("", "authfly-static-*")
	if err != nil {
		return "", fmt.Errorf("static temp: %w", err)
	}
	err = fs.WalkDir(staticFS, "static", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel("static", path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		raw, err := staticFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
	if err != nil {
		_ = os.RemoveAll(root)
		return "", fmt.Errorf("static materialize: %w", err)
	}
	return root, nil
}
