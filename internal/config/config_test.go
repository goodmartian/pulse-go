package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Errorf("Expected config dir to not be empty")
	}
	if !strings.Contains(dir, "pulse") {
		t.Errorf("Expected config dir to contain 'pulse', got: %s", dir)
	}
}

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"my-project", false},
		{"project_v2", false},
		{"simple", false},
		{"CamelCase", false},
		{"with.dot", false},
		{"../etc/passwd", true},
		{"/absolute/path", true},
		{"foo/bar", true},
		{"foo\\bar", true},
		{"..", true},
		{"", true},
		{"a/../b", true},
	}
	for _, tt := range tests {
		err := ValidateProjectName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestSaveLoadState(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	original := State{
		Focus:      "test-project",
		FocusSince: "2026-03-01T12:00:00Z",
		Statuses:   map[string]string{"proj1": "active", "proj2": "done"},
		Switches: []Switch{
			{From: "a", To: "b", Reason: "test", Date: "2026-03-01T12:00:00Z"},
		},
		Lang: "en",
	}

	if err := SaveState(original); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	loaded := LoadState()
	if loaded.Focus != original.Focus {
		t.Errorf("Focus: got %q, want %q", loaded.Focus, original.Focus)
	}
	if loaded.FocusSince != original.FocusSince {
		t.Errorf("FocusSince: got %q, want %q", loaded.FocusSince, original.FocusSince)
	}
	if len(loaded.Statuses) != len(original.Statuses) {
		t.Errorf("Statuses length: got %d, want %d", len(loaded.Statuses), len(original.Statuses))
	}
	if len(loaded.Switches) != len(original.Switches) {
		t.Errorf("Switches length: got %d, want %d", len(loaded.Switches), len(original.Switches))
	}
}

func TestSaveLoadConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	original := Config{
		ProjectsDir: "/home/user/projects",
		Lang:        "ru",
	}

	if err := SaveConfig(original); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded := LoadConfig()
	if loaded.ProjectsDir != original.ProjectsDir {
		t.Errorf("ProjectsDir: got %q, want %q", loaded.ProjectsDir, original.ProjectsDir)
	}
	if loaded.Lang != original.Lang {
		t.Errorf("Lang: got %q, want %q", loaded.Lang, original.Lang)
	}
}

func TestSaveLoadIdeas(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	original := []Idea{
		{Text: "idea 1", Date: "2026-03-01T12:00:00Z", Project: "proj1"},
		{Text: "idea 2", Date: "2026-03-01T13:00:00Z", Notes: "some notes"},
	}

	if err := SaveIdeas(original); err != nil {
		t.Fatalf("SaveIdeas failed: %v", err)
	}

	loaded := LoadIdeas()
	if len(loaded) != len(original) {
		t.Fatalf("ideas length: got %d, want %d", len(loaded), len(original))
	}
	if loaded[0].Text != original[0].Text {
		t.Errorf("ideas[0].Text: got %q, want %q", loaded[0].Text, original[0].Text)
	}
	if loaded[1].Notes != original[1].Notes {
		t.Errorf("ideas[1].Notes: got %q, want %q", loaded[1].Notes, original[1].Notes)
	}
}

func TestSaveLoadProjectData(t *testing.T) {
	tmp := t.TempDir()
	PulseDir = tmp

	original := ProjectData{
		Name:        "test-proj",
		Tagline:     "A test project",
		Description: "Longer description",
		DoneWhen:    "When tests pass",
		Stack:       "Go, TUI",
		Notes:       "Some notes",
	}

	if err := SaveProjectData("test-proj", original); err != nil {
		t.Fatalf("SaveProjectData failed: %v", err)
	}

	loaded := LoadProjectData("test-proj")
	if loaded.Tagline != original.Tagline {
		t.Errorf("Tagline: got %q, want %q", loaded.Tagline, original.Tagline)
	}
	if loaded.Stack != original.Stack {
		t.Errorf("Stack: got %q, want %q", loaded.Stack, original.Stack)
	}
	if loaded.Description != original.Description {
		t.Errorf("Description: got %q, want %q", loaded.Description, original.Description)
	}
}

func TestSaveProjectDataPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	PulseDir = tmp

	pd := ProjectData{Name: "evil"}
	err := SaveProjectData("../evil", pd)
	if err == nil {
		t.Error("expected error for path traversal name, got nil")
	}

	if _, statErr := os.Stat(filepath.Join(tmp, "evil.json")); statErr == nil {
		t.Error("file was created outside projects directory")
	}
}
