package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/goodmartian/pulse-go/internal/config"
	"github.com/goodmartian/pulse-go/internal/git"
	"github.com/goodmartian/pulse-go/internal/i18n"
)

func cmdDashboard() {
	state := config.LoadState()
	focus := state.Focus

	StepHeader("◈ PULSE")

	if focus == "" {
		StepError(i18n.Tr("dashboard.no_focus"))
		StepResult(i18n.Tr("dashboard.pick_one"))
		fmt.Println()
		cmdMapShort()
		return
	}

	projectPath := filepath.Join(config.ProjectsDir, focus)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		StepError(i18n.Tr("project.not_found", focus))
		return
	}

	daysFocused := 0
	if state.FocusSince != "" {
		if t, err := time.Parse(time.RFC3339, state.FocusSince); err == nil {
			daysFocused = int(time.Since(t).Hours() / 24)
		}
	}

	Step(i18n.Tr("dashboard.focus",
		GreenStyle.Bold(true).Render(focus),
		DimStyle.Render(i18n.Tr("dashboard.focus_days", daysFocused))))
	StepBlank()

	if git.IsGitRepo(projectPath) {
		streak := git.GetStreak(projectPath)
		Step(i18n.Tr("dashboard.streak", streakBar(streak)))
		StepBlank()

		heatmap := commitHeatmap(projectPath)
		fmt.Println(heatmap)
		StepBlank()

		today := git.GetTodayCommits(projectPath)
		if len(today) > 0 {
			Step(i18n.Tr("dashboard.today", len(today)))
			for j, c := range today {
				if j >= 5 {
					StepResult(DimStyle.Render(i18n.Tr("dashboard.and_more", len(today)-5)))
					break
				}
				StepResultStyled("✓ "+c, GreenStyle)
			}
		} else {
			StepError(i18n.Tr("dashboard.zero_commits"))
		}

		week := git.GetWeekCommits(projectPath)
		total := git.GetCommitCount(projectPath)
		StepBlank()
		StepResult(i18n.Tr("dashboard.week_total", len(week), total))
	}
	StepBlank()

	ideas := config.LoadIdeas()
	if len(ideas) > 0 {
		StepDim(i18n.Tr("dashboard.parked_ideas", len(ideas)))
	}

	projects := git.ScanProjects(config.ProjectsDir)
	var hot []string
	for _, p := range projects {
		if d := git.DaysAgo(p.LastDate); p.Name != focus && d >= 0 && d <= 3 {
			hot = append(hot, p.Name)
		}
	}
	if len(hot) > 0 {
		StepDim(i18n.Tr("dashboard.hot_not_focus") + strings.Join(hot, ", "))
	}
	StepBlank()
	StepDim(DimStyle.Render(config.ProjectsDir))
	fmt.Println()
}

func cmdFocus(name string) bool {
	if err := config.ValidateProjectName(name); err != nil {
		fmt.Println()
		StepError(err.Error())
		fmt.Println()
		return false
	}
	state := config.LoadState()
	current := state.Focus

	projectPath := filepath.Join(config.ProjectsDir, name)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		fmt.Println()
		StepError(i18n.Tr("focus.not_found", name, config.ProjectsDir))
		fmt.Println()
		return false
	}

	if current != "" && current != name {
		fmt.Println()
		StepError(i18n.Tr("focus.switch_title"))
		StepResult(i18n.Tr("focus.current", BoldStyle.Render(current)))

		currentPath := filepath.Join(config.ProjectsDir, current)
		if git.IsGitRepo(currentPath) {
			streak := git.GetStreak(currentPath)
			commits := git.GetCommitCount(currentPath)
			week := len(git.GetWeekCommits(currentPath))
			StepResult(i18n.Tr("focus.stats", streak, week, commits))
		}

		StepResult(i18n.Tr("focus.switch_to", CyanStyle.Render(name)))
		StepBlank()

		var reason string
		var confirm bool

		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(i18n.Tr("focus.why")).
					Value(&reason),
				huh.NewConfirm().
					Title(i18n.Tr("focus.confirm")).
					Affirmative(i18n.Tr("focus.yes")).
					Negative(i18n.Tr("focus.cancel")).
					Value(&confirm),
			),
		).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

		if err != nil || !confirm || reason == "" {
			fmt.Println()
			Step(i18n.Tr("focus.stay", current))
			fmt.Println()
			return false
		}

		state.Switches = append(state.Switches, config.Switch{
			From: current, To: name, Reason: reason,
			Date: time.Now().Format(time.RFC3339),
		})
	}

	state.Focus = name
	state.FocusSince = time.Now().Format(time.RFC3339)
	if err := config.SaveState(state); err != nil {
		fmt.Println()
		StepError(i18n.Tr("error.save_failed") + err.Error())
		fmt.Println()
		return false
	}
	fmt.Println()
	Step(i18n.Tr("focus.set") + name)
	StepResult(i18n.Tr("focus.rest_waits"))
	fmt.Println()
	return true
}

func cmdPark(text string) {
	if text == "" {
		fmt.Println()
		StepError(i18n.Tr("park.usage"))
		fmt.Println()
		return
	}
	ideas := config.LoadIdeas()
	ideas = append(ideas, config.Idea{Text: text, Date: time.Now().Format(time.RFC3339)})
	if err := config.SaveIdeas(ideas); err != nil {
		fmt.Println()
		StepError(i18n.Tr("error.save_failed") + err.Error())
		fmt.Println()
		return
	}
	fmt.Println()
	Step(i18n.Tr("park.saved") + text)
	StepResult(i18n.Tr("park.dont_touch"))
	fmt.Println()
}

func cmdIdeas() {
	ideas := config.LoadIdeas()
	fmt.Println()
	if len(ideas) == 0 {
		StepDim(i18n.Tr("ideas.empty"))
	} else {
		Step(i18n.Tr("ideas.title", len(ideas)))
		for idx, idea := range ideas {
			StepResult(ideaLabel(idx, idea))
			if idea.Notes != "" {
				StepResult("   " + DimStyle.Render("\""+idea.Notes+"\""))
			}
		}
	}
	fmt.Println()
}

func cmdIdeasInteractive() {
	for {
		ideas := config.LoadIdeas()
		if len(ideas) == 0 {
			fmt.Println()
			StepDim(i18n.Tr("ideas.empty"))
			fmt.Println()
			return
		}

		options := make([]huh.Option[int], 0, len(ideas))
		for idx, idea := range ideas {
			options = append(options, huh.NewOption(ideaLabel(idx, idea), idx))
		}

		selected := -1
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title(i18n.Tr("ideas.select")).
					Options(options...).
					Value(&selected),
			),
		).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

		if err != nil || selected < 0 {
			return
		}

		var action string
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(ideas[selected].Text).
					Options(
						huh.NewOption(i18n.Tr("ideas.start"), "start"),
						huh.NewOption(i18n.Tr("ideas.edit"), "edit"),
						huh.NewOption(i18n.Tr("ideas.delete"), "delete"),
						huh.NewOption(i18n.Tr("back"), actionBack),
					).
					Value(&action),
			),
		).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMapBack()).Run()

		if err != nil || action == "" || action == actionBack {
			continue
		}

		switch action {
		case "start":
			cmdStartIdea(selected)
			return
		case "edit":
			cmdEditIdea(selected)
		case "delete":
			cmdDeleteIdea(selected)
		}
	}
}

func cmdEditIdea(idx int) {
	ideas := config.LoadIdeas()
	if idx < 0 || idx >= len(ideas) {
		return
	}
	idea := ideas[idx]

	text := idea.Text
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(i18n.Tr("park.describe")).
				Value(&text),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil {
		return
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	Step(i18n.Tr("park.describe") + " " + GreenStyle.Render(text))

	project, notes, ok := promptProjectAndNotes(idea.Project, idea.Notes)
	if !ok {
		return
	}

	ideas[idx].Text = text
	ideas[idx].Project = project
	ideas[idx].Notes = notes
	if err := config.SaveIdeas(ideas); err != nil {
		fmt.Println()
		StepError(i18n.Tr("error.save_failed") + err.Error())
		fmt.Println()
		return
	}
	fmt.Println()
	Step(i18n.Tr("ideas.saved"))
	fmt.Println()
}

func cmdDeleteIdea(idx int) {
	ideas := config.LoadIdeas()
	if idx < 0 || idx >= len(ideas) {
		return
	}

	var confirm bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(i18n.Tr("ideas.delete_confirm")).
				Description(ideas[idx].Text).
				Affirmative(i18n.Tr("focus.yes")).
				Negative(i18n.Tr("focus.cancel")).
				Value(&confirm),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil || !confirm {
		return
	}

	ideas = append(ideas[:idx], ideas[idx+1:]...)
	if err := config.SaveIdeas(ideas); err != nil {
		fmt.Println()
		StepError(i18n.Tr("error.save_failed") + err.Error())
		fmt.Println()
		return
	}
	fmt.Println()
	Step(i18n.Tr("ideas.deleted"))
	fmt.Println()
}

func cmdStartIdea(idx int) {
	ideas := config.LoadIdeas()
	if idx < 0 || idx >= len(ideas) {
		return
	}
	idea := ideas[idx]

	project := idea.Project
	if project == "" {
		var name string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(i18n.Tr("ideas.folder_name")).
					Value(&name),
			),
		).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

		if err != nil || strings.TrimSpace(name) == "" {
			return
		}
		project = strings.TrimSpace(name)
	}

	if err := config.ValidateProjectName(project); err != nil {
		StepError(err.Error())
		return
	}

	projectPath := filepath.Join(config.ProjectsDir, project)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		StepError(err.Error())
		return
	}

	ideaFile := filepath.Join(projectPath, "IDEA.md")
	content := "# " + idea.Text + "\n\n" + shortDate(idea.Date) + "\n"
	if idea.Notes != "" {
		content += "\n" + idea.Notes + "\n"
	}
	if err := os.WriteFile(ideaFile, []byte(content), 0644); err != nil {
		StepError(err.Error())
		return
	}
	Step(i18n.Tr("ideas.idea_saved_file") + " " + DimStyle.Render("IDEA.md"))

	state := config.LoadState()
	if state.Focus != "" && state.Focus != project {
		state.Switches = append(state.Switches, config.Switch{
			From: state.Focus, To: project, Reason: idea.Text,
			Date: time.Now().Format(time.RFC3339),
		})
	}
	state.Focus = project
	state.FocusSince = time.Now().Format(time.RFC3339)
	if err := config.SaveState(state); err != nil {
		StepError(i18n.Tr("error.save_failed") + err.Error())
		return
	}

	ideas = append(ideas[:idx], ideas[idx+1:]...)
	if err := config.SaveIdeas(ideas); err != nil {
		StepError(i18n.Tr("error.save_failed") + err.Error())
		return
	}

	fmt.Println()
	Step(i18n.Tr("ideas.started") + " " + GreenStyle.Render(project))
	StepResult(i18n.Tr("focus.rest_waits"))
	fmt.Println()
}

func cmdMap() {
	state := config.LoadState()
	focus := state.Focus
	projects := git.ScanProjects(config.ProjectsDir)

	StepHeader(i18n.Tr("map.title", len(projects)))

	sortProjectsByRecency(projects, focus)

	for _, p := range projects {
		isFocus := p.Name == focus
		override := state.Statuses[p.Name]
		if isFocus {
			override = statusActive
		}
		badge := statusBadge(p.LastDate, override)

		commits := ""
		if p.IsGit {
			commits = fmt.Sprintf("%dc", p.Commits)
		}
		ago := timeAgo(p.LastDate)
		msg := p.LastMsg
		if len(msg) > 40 {
			msg = msg[:40]
		}

		pad := 22 - len(p.Name)
		if pad < 1 {
			pad = 1
		}
		detail := DimStyle.Render(fmt.Sprintf("%5s  %14s  %s", commits, ago, msg))

		if isFocus {
			name := GreenStyle.Bold(true).Render("▶ "+p.Name) + strings.Repeat(" ", pad)
			fmt.Printf("  %s %s  %s\n", name, badge, detail)
		} else {
			name := "  " + p.Name + strings.Repeat(" ", pad)
			fmt.Printf("  %s %s  %s\n", name, badge, detail)
		}
	}

	fmt.Println()
	wip := "0/1"
	if focus != "" {
		wip = "1/1 ✓"
	}
	StepDim(i18n.Tr("map.wip", wip))
	fmt.Println()
}

func cmdMapShort() {
	projects := git.ScanProjects(config.ProjectsDir)
	var hot []git.Project
	for _, p := range projects {
		if d := git.DaysAgo(p.LastDate); d >= 0 && d <= 7 {
			hot = append(hot, p)
		}
	}
	sort.Slice(hot, func(i, j int) bool {
		return git.DaysAgo(hot[i].LastDate) < git.DaysAgo(hot[j].LastDate)
	})

	if len(hot) > 0 {
		StepDim(i18n.Tr("map.active_week"))
		for idx, p := range hot {
			if idx >= 6 {
				break
			}
			StepResult(fmt.Sprintf("• %s  %s", p.Name, DimStyle.Render("("+timeAgo(p.LastDate)+")")))
		}
	}
	fmt.Println()
}

func cmdLog(period string) {
	projects := git.ScanProjects(config.ProjectsDir)

	isWeek := period == "week"
	if isWeek {
		StepHeader(i18n.Tr("log.week_title"))
	} else {
		StepHeader(i18n.Tr("log.today_title"))
	}

	total := 0
	for _, p := range projects {
		if !p.IsGit {
			continue
		}
		var commits []string
		if isWeek {
			commits = git.GetWeekCommits(p.Path)
		} else {
			commits = git.GetTodayCommits(p.Path)
		}
		if len(commits) == 0 {
			continue
		}
		total += len(commits)
		Step(i18n.Tr("log.project_commits", p.Name, len(commits)))
		limit := 8
		for j, c := range commits {
			if j >= limit {
				StepResult(DimStyle.Render(i18n.Tr("dashboard.and_more", len(commits)-limit)))
				break
			}
			StepResultStyled("✓ "+c, GreenStyle)
		}
		StepBlank()
	}

	if total == 0 {
		if isWeek {
			StepError(i18n.Tr("log.zero_week"))
		} else {
			StepError(i18n.Tr("log.zero_today"))
		}
	} else {
		label := i18n.Tr("log.label_today")
		if isWeek {
			label = i18n.Tr("log.label_week")
		}
		Step(i18n.Tr("log.total", total, label))
	}
	fmt.Println()
}

func cmdStatus(name, status string) {
	valid := map[string]bool{statusDone: true, statusPaused: true, statusDead: true, statusActive: true, statusIdea: true}
	if !valid[status] {
		fmt.Println()
		StepError(i18n.Tr("status.invalid"))
		fmt.Println()
		return
	}
	state := config.LoadState()
	state.Statuses[name] = status
	if err := config.SaveState(state); err != nil {
		fmt.Println()
		StepError(i18n.Tr("error.save_failed") + err.Error())
		fmt.Println()
		return
	}
	fmt.Println()
	Step(name + " → " + status)
	fmt.Println()
}

func cmdSwitches() {
	state := config.LoadState()
	fmt.Println()
	if len(state.Switches) == 0 {
		Step(i18n.Tr("switches.empty"))
	} else {
		Step(i18n.Tr("switches.title", len(state.Switches)))
		start := 0
		if len(state.Switches) > 10 {
			start = len(state.Switches) - 10
		}
		for _, s := range state.Switches[start:] {
			StepResult(fmt.Sprintf("%s  %s → %s", DimStyle.Render(shortDate(s.Date)), s.From, CyanStyle.Render(s.To)))
			StepResult(i18n.Tr("switches.reason", s.Reason))
		}
		if len(state.Switches) > 3 {
			StepBlank()
			StepError(i18n.Tr("switches.warning", len(state.Switches)))
		}
	}
	fmt.Println()
}

func cmdInfo(name string) {
	if err := config.ValidateProjectName(name); err != nil {
		fmt.Println()
		StepError(err.Error())
		fmt.Println()
		return
	}
	pd := config.LoadProjectData(name)
	StepHeader("◈ " + name)

	projectPath := filepath.Join(config.ProjectsDir, name)
	if git.IsGitRepo(projectPath) {
		commits := git.GetCommitCount(projectPath)
		branch := git.GetBranch(projectPath)
		last := timeAgo(git.GetLastCommitDate(projectPath))
		StepResult(fmt.Sprintf("%s │ %d commits │ %s", branch, commits, last))
		StepBlank()
	}

	hasData := false
	for _, pair := range [][2]string{
		{"Tagline", pd.Tagline}, {"Stack", pd.Stack}, {"Done when", pd.DoneWhen},
		{"Description", pd.Description}, {"Notes", pd.Notes},
	} {
		if pair[1] != "" {
			hasData = true
			Step(pair[0] + ":")
			for _, line := range strings.Split(pair[1], "\n") {
				StepResult(line)
			}
		}
	}

	if !hasData {
		StepDim(i18n.Tr("info.no_data", name))
	}
	fmt.Println()
}

func cmdUnpark(arg string) {
	idx, err := strconv.Atoi(arg)
	if err != nil || idx < 1 {
		fmt.Println()
		StepError(i18n.Tr("unpark.usage"))
		fmt.Println()
		return
	}
	ideas := config.LoadIdeas()
	if idx > len(ideas) {
		fmt.Println()
		StepError(i18n.Tr("unpark.not_found", idx, len(ideas)))
		fmt.Println()
		return
	}
	removed := ideas[idx-1]
	ideas = append(ideas[:idx-1], ideas[idx:]...)
	if err := config.SaveIdeas(ideas); err != nil {
		fmt.Println()
		StepError(i18n.Tr("error.save_failed") + err.Error())
		fmt.Println()
		return
	}
	fmt.Println()
	Step(i18n.Tr("unpark.removed") + removed.Text)
	StepResult(i18n.Tr("unpark.remaining", len(ideas)))
	fmt.Println()
}
