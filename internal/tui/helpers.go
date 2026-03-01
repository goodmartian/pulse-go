package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/goodmartian/pulse-go/internal/config"
	"github.com/goodmartian/pulse-go/internal/git"
	"github.com/goodmartian/pulse-go/internal/i18n"
)

func timeAgo(isoStr string) string {
	if isoStr == "" {
		return i18n.Tr("time.never")
	}
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return "?"
	}
	d := time.Since(t)
	days := int(d.Hours() / 24)
	if days > 30 {
		return i18n.Tr("time.months_ago", days/30)
	}
	if days > 0 {
		return i18n.Tr("time.days_ago", days)
	}
	hours := int(d.Hours())
	if hours > 0 {
		return i18n.Tr("time.hours_ago", hours)
	}
	mins := int(d.Minutes())
	if mins > 0 {
		return i18n.Tr("time.minutes_ago", mins)
	}
	return i18n.Tr("time.just_now")
}

func commitHeatmap(path string) string {
	const weeks = 26
	const cellW = 2
	counts := git.GetCommitCounts(path, weeks)
	if counts == nil {
		counts = map[string]int{}
	}

	ansiCell := func(color int) string {
		return fmt.Sprintf("\033[38;5;%dm■\033[0m ", color)
	}
	colorFor := func(count int) string {
		switch {
		case count == 0:
			return ansiCell(238)
		case count == 1:
			return ansiCell(22)
		case count <= 3:
			return ansiCell(28)
		default:
			return ansiCell(34)
		}
	}

	now := time.Now()
	start := now.AddDate(0, 0, -weeks*7)
	for start.Weekday() != time.Monday {
		start = start.AddDate(0, 0, -1)
	}

	type cell struct {
		date  string
		count int
	}
	grid := make([][]cell, 7)
	for i := range grid {
		grid[i] = make([]cell, 0, weeks+1)
	}
	d := start
	for !d.After(now) {
		ds := d.Format("2006-01-02")
		row := int(d.Weekday())
		if row == 0 {
			row = 6
		} else {
			row--
		}
		grid[row] = append(grid[row], cell{date: ds, count: counts[ds]})
		d = d.AddDate(0, 0, 1)
	}

	colCount := len(grid[0])
	indent := "        "

	monthBuf := make([]byte, colCount*cellW)
	for i := range monthBuf {
		monthBuf[i] = ' '
	}
	prevMonth := -1
	for col := 0; col < colCount; col++ {
		t, _ := time.Parse("2006-01-02", grid[0][col].date)
		m := int(t.Month())
		if m != prevMonth {
			label := t.Format("Jan")
			pos := col * cellW
			canPlace := true
			if pos > 0 && pos < len(monthBuf) && monthBuf[pos-1] != ' ' {
				canPlace = false
			}
			if canPlace && pos+len(label) <= len(monthBuf) {
				copy(monthBuf[pos:], label)
				prevMonth = m
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(indent)
	sb.Write(monthBuf)
	sb.WriteString("\n")

	rowLabels := map[int]string{
		0: i18n.Tr("heatmap.mon"),
		2: i18n.Tr("heatmap.wed"),
		4: i18n.Tr("heatmap.fri"),
	}

	for row := 0; row < 7; row++ {
		if label, ok := rowLabels[row]; ok {
			sb.WriteString(fmt.Sprintf("  %-5s ", label))
		} else {
			sb.WriteString(indent)
		}
		for _, c := range grid[row] {
			sb.WriteString(colorFor(c.count))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("%s%s %s %s %s %s %s",
		indent,
		i18n.Tr("heatmap.less"),
		ansiCell(238), ansiCell(22), ansiCell(28), ansiCell(34),
		i18n.Tr("heatmap.more"),
	))

	return sb.String()
}

func streakBar(streak int) string {
	if streak == 0 {
		return DimStyle.Render(i18n.Tr("streak.none"))
	}
	n := streak
	if n > 30 {
		n = 30
	}
	blocks := strings.Repeat("█", n)
	var s string
	if streak >= 7 {
		s = GreenStyle.Render(blocks)
	} else if streak >= 3 {
		s = YellowStyle.Render(blocks)
	} else {
		s = RedStyle.Render(blocks)
	}
	return fmt.Sprintf("%s %s", s, i18n.Tr("streak.days", streak))
}

func shortDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func ideaLabel(idx int, idea config.Idea) string {
	label := fmt.Sprintf("%d. %s  %s", idx+1, idea.Text, DimStyle.Render("("+shortDate(idea.Date)+")"))
	if idea.Project != "" {
		label += DimStyle.Render("  → " + idea.Project)
	}
	return label
}

func buildProjectOptions() []huh.Option[string] {
	projects := git.ScanProjects(config.ProjectsDir)
	opts := []huh.Option[string]{
		huh.NewOption(i18n.Tr("ideas.no_project"), ""),
	}
	for _, p := range projects {
		opts = append(opts, huh.NewOption(p.Name, p.Name))
	}
	opts = append(opts, huh.NewOption(i18n.Tr("ideas.create_folder"), actionCreate))
	return opts
}

func promptCreateFolder() (string, bool) {
	var name string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(i18n.Tr("ideas.folder_name")).
				Value(&name),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil || strings.TrimSpace(name) == "" {
		return "", false
	}
	name = strings.TrimSpace(name)
	if err := config.ValidateProjectName(name); err != nil {
		StepError(err.Error())
		return "", false
	}
	path := filepath.Join(config.ProjectsDir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		StepError(err.Error())
		return "", false
	}
	Step(i18n.Tr("ideas.folder_created") + " " + name)
	return name, true
}

func pickProject(titleKey string) (string, bool) {
	projects := git.ScanProjects(config.ProjectsDir)
	if len(projects) == 0 {
		StepError(i18n.Tr("project.none_found"))
		return "", false
	}
	options := make([]huh.Option[string], 0, len(projects))
	for _, p := range projects {
		options = append(options, huh.NewOption(p.Name, p.Name))
	}
	var selected string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(i18n.Tr(titleKey)).
				Options(options...).
				Value(&selected),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil || selected == "" {
		return "", false
	}
	return selected, true
}

func sortProjectsByRecency(projects []git.Project, pinFocus string) {
	sort.Slice(projects, func(i, j int) bool {
		if pinFocus != "" {
			if projects[i].Name == pinFocus {
				return true
			}
			if projects[j].Name == pinFocus {
				return false
			}
		}
		di := git.DaysAgo(projects[i].LastDate)
		dj := git.DaysAgo(projects[j].LastDate)
		if di < 0 && dj < 0 {
			return projects[i].Name < projects[j].Name
		}
		if di < 0 {
			return false
		}
		if dj < 0 {
			return true
		}
		return di < dj
	})
}

func promptProjectAndNotes(initialProject, initialNotes string) (string, string, bool) {
	project := initialProject
	projectOptions := buildProjectOptions()
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(i18n.Tr("ideas.project")).
				Options(projectOptions...).
				Height(30).
				Value(&project),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil {
		return "", "", false
	}

	if project == actionCreate {
		name, ok := promptCreateFolder()
		if !ok {
			return "", "", false
		}
		project = name
	}

	if project != "" {
		Step(i18n.Tr("ideas.project") + ": " + GreenStyle.Render(project))
	} else {
		StepDim(i18n.Tr("ideas.project") + ": " + i18n.Tr("ideas.no_project"))
	}

	notes := initialNotes
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(i18n.Tr("ideas.notes")).
				Value(&notes),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil {
		return "", "", false
	}

	notes = strings.TrimSpace(notes)
	if notes != "" {
		Step(i18n.Tr("ideas.notes") + ": " + GreenStyle.Render(notes))
	}

	return project, notes, true
}
