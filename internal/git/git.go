package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Project struct {
	Name     string
	Path     string
	IsGit    bool
	Commits  int
	LastDate string
	LastMsg  string
	Branch   string
}

var skipDirs = map[string]bool{
	".pulse": true, "dead-projects": true, "sandbox": true,
	"__pycache__": true, "node_modules": true, ".git": true,
}

func gitCmd(projectPath string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", projectPath}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func IsGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func GetLastCommitDate(path string) string {
	return gitCmd(path, "log", "-1", "--format=%aI")
}

func getLastCommitMsg(path string) string {
	return gitCmd(path, "log", "-1", "--format=%s")
}

func GetCommitCount(path string) int {
	s := gitCmd(path, "rev-list", "--count", "HEAD")
	n, _ := strconv.Atoi(s)
	return n
}

func GetBranch(path string) string {
	b := gitCmd(path, "branch", "--show-current")
	if b == "" {
		return "?"
	}
	return b
}

func GetTodayCommits(path string) []string {
	today := time.Now().Format("2006-01-02")
	log := gitCmd(path, "log", "--since="+today+" 00:00", "--format=%s")
	if log == "" {
		return nil
	}
	var result []string
	for _, l := range strings.Split(log, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			result = append(result, l)
		}
	}
	return result
}

func GetWeekCommits(path string) []string {
	weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	log := gitCmd(path, "log", "--since="+weekAgo, "--format=%h %s")
	if log == "" {
		return nil
	}
	var result []string
	for _, l := range strings.Split(log, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			result = append(result, l)
		}
	}
	return result
}

func GetStreak(path string) int {
	log := gitCmd(path, "log", "--since=365 days ago", "--format=%aI")
	if log == "" {
		return 0
	}
	dates := map[string]bool{}
	for _, line := range strings.Split(log, "\n") {
		if d := strings.TrimSpace(line); len(d) >= 10 {
			dates[d[:10]] = true
		}
	}
	streak := 0
	day := time.Now()
	if !dates[day.Format("2006-01-02")] {
		day = day.AddDate(0, 0, -1)
	}
	for dates[day.Format("2006-01-02")] {
		streak++
		day = day.AddDate(0, 0, -1)
	}
	return streak
}

func ScanProjects(projectsDir string) []Project {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if skipDirs[name] || strings.HasPrefix(name, "__") || strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(projectsDir, name)

		if !IsGitRepo(fullPath) {
			subEntries, _ := os.ReadDir(fullPath)
			if len(subEntries) > 0 {
				projects = append(projects, Project{
					Name: name, Path: fullPath, IsGit: false,
				})
			}
			continue
		}

		projects = append(projects, Project{
			Name:     name,
			Path:     fullPath,
			IsGit:    true,
			Commits:  GetCommitCount(fullPath),
			LastDate: GetLastCommitDate(fullPath),
			LastMsg:  getLastCommitMsg(fullPath),
			Branch:   GetBranch(fullPath),
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})
	return projects
}

func GetCommitCounts(path string, weeks int) map[string]int {
	since := time.Now().AddDate(0, 0, -weeks*7).Format("2006-01-02")
	log := gitCmd(path, "log", "--since="+since, "--format=%aI")
	if log == "" {
		return nil
	}
	counts := map[string]int{}
	for _, line := range strings.Split(log, "\n") {
		if d := strings.TrimSpace(line); len(d) >= 10 {
			counts[d[:10]]++
		}
	}
	return counts
}

func DaysAgo(isoStr string) int {
	if isoStr == "" {
		return -1
	}
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return -1
	}
	return int(time.Since(t).Hours() / 24)
}
