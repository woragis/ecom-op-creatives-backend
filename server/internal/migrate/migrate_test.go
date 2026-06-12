package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDirFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MIGRATIONS_DIR", dir)
	got := ResolveDir()
	if got != dir {
		t.Fatalf("ResolveDir() = %q, want %q", got, dir)
	}
}

func TestResolveDirFindsRelativeMigrations(t *testing.T) {
	root := t.TempDir()
	migrations := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrations, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	t.Setenv("MIGRATIONS_DIR", "")
	got := ResolveDir()
	if got != "migrations" {
		t.Fatalf("ResolveDir() = %q, want migrations", got)
	}
}
