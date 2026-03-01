package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ProjectsDir string
	PulseDir    string
)

type Config struct {
	ProjectsDir string `json:"projects_dir"`
	Lang        string `json:"lang"`
}

type State struct {
	Focus      string            `json:"focus"`
	FocusSince string            `json:"focus_since"`
	Statuses   map[string]string `json:"statuses"`
	Switches   []Switch          `json:"switches"`
	Lang       string            `json:"lang"`
}

type Switch struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
	Date   string `json:"date"`
}

type Idea struct {
	Text    string `json:"text"`
	Date    string `json:"date"`
	Project string `json:"project,omitempty"`
	Notes   string `json:"notes,omitempty"`
}

type ProjectData struct {
	Name        string `json:"name"`
	Tagline     string `json:"tagline"`
	Description string `json:"description"`
	DoneWhen    string `json:"done_when"`
	Stack       string `json:"stack"`
	Notes       string `json:"notes"`
}

func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("project name cannot be an absolute path: %s", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("project name cannot contain '..': %s", name)
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("project name cannot contain path separators: %s", name)
	}
	return nil
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func ConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "pulse")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func LoadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return Config{}
	}
	var c Config
	json.Unmarshal(data, &c)
	return c
}

func SaveConfig(c Config) error {
	return writeJSON(configPath(), c)
}

func LoadState() State {
	dir := ConfigDir()
	data, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		data, err = os.ReadFile(filepath.Join(PulseDir, "state.json"))
		if err != nil {
			return State{Statuses: map[string]string{}}
		}
	}
	var s State
	if json.Unmarshal(data, &s) != nil {
		return State{Statuses: map[string]string{}}
	}
	if s.Statuses == nil {
		s.Statuses = map[string]string{}
	}
	return s
}

func SaveState(s State) error {
	return writeJSON(filepath.Join(ConfigDir(), "state.json"), s)
}

func LoadIdeas() []Idea {
	dir := ConfigDir()
	data, err := os.ReadFile(filepath.Join(dir, "ideas.json"))
	if err != nil {
		data, err = os.ReadFile(filepath.Join(PulseDir, "ideas.json"))
		if err != nil {
			return nil
		}
	}
	var ideas []Idea
	if err := json.Unmarshal(data, &ideas); err != nil {
		fmt.Fprintf(os.Stderr, "pulse: error reading ideas.json: %v\n", err)
		return nil
	}
	return ideas
}

func SaveIdeas(ideas []Idea) error {
	return writeJSON(filepath.Join(ConfigDir(), "ideas.json"), ideas)
}

func LoadProjectData(name string) ProjectData {
	if err := ValidateProjectName(name); err != nil {
		return ProjectData{Name: name}
	}
	dir := filepath.Join(PulseDir, "projects")
	data, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if err != nil {
		return ProjectData{Name: name}
	}
	var pd ProjectData
	if err := json.Unmarshal(data, &pd); err != nil {
		fmt.Fprintf(os.Stderr, "pulse: error reading %s.json: %v\n", name, err)
		return ProjectData{Name: name}
	}
	if pd.Name == "" {
		pd.Name = name
	}
	return pd
}

func SaveProjectData(name string, pd ProjectData) error {
	if err := ValidateProjectName(name); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}
	return writeJSON(filepath.Join(PulseDir, "projects", name+".json"), pd)
}

func Init() {
	if env := os.Getenv("PULSE_DIR"); env != "" {
		ProjectsDir = env
		PulseDir = filepath.Join(env, ".pulse")
		return
	}

	cfg := LoadConfig()
	if cfg.ProjectsDir != "" {
		ProjectsDir = cfg.ProjectsDir
		PulseDir = filepath.Join(cfg.ProjectsDir, ".pulse")
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	ProjectsDir = dir
	PulseDir = filepath.Join(dir, ".pulse")
}

func NeedsSetup() bool {
	return os.Getenv("PULSE_DIR") == "" && LoadConfig().ProjectsDir == ""
}
