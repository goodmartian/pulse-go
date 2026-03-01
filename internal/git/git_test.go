package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	tmp := t.TempDir()

	if IsGitRepo(tmp) {
		t.Error("expected false for non-git directory")
	}

	gitDir := filepath.Join(tmp, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if !IsGitRepo(tmp) {
		t.Error("expected true for directory with .git dir")
	}

	tmp2 := t.TempDir()
	gitFile := filepath.Join(tmp2, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: /some/path"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsGitRepo(tmp2) {
		t.Error("expected true for directory with .git file (worktree)")
	}
}

func TestDaysAgo(t *testing.T) {
	if DaysAgo("") != -1 {
		t.Error("expected -1 for empty string")
	}
	if DaysAgo("invalid") != -1 {
		t.Error("expected -1 for invalid date")
	}
}
