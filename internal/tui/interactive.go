package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/goodmartian/pulse-go/internal/config"
	"github.com/goodmartian/pulse-go/internal/git"
	"github.com/goodmartian/pulse-go/internal/i18n"
)

func cmdFocusInteractive() {
	projects := git.ScanProjects(config.ProjectsDir)
	if len(projects) == 0 {
		StepError(i18n.Tr("project.none_found"))
		return
	}

	sortProjectsByRecency(projects, "")

	options := make([]huh.Option[string], 0, len(projects))
	for _, p := range projects {
		label := p.Name
		if p.IsGit {
			label += DimStyle.Render(fmt.Sprintf("  %s  %dc", timeAgo(p.LastDate), p.Commits))
		}
		options = append(options, huh.NewOption(label, p.Name))
	}

	for {
		var selected string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(i18n.Tr("focus.pick")).
					Options(options...).
					Value(&selected),
			),
		).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

		if err != nil || selected == "" {
			return
		}
		if cmdFocus(selected) {
			return
		}
	}
}

func cmdParkInteractive() {
	var text string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(i18n.Tr("park.describe")).
				Placeholder(i18n.Tr("park.placeholder")).
				Value(&text),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil || strings.TrimSpace(text) == "" {
		return
	}
	text = strings.TrimSpace(text)
	Step(i18n.Tr("park.describe") + " " + GreenStyle.Render(text))

	project, notes, ok := promptProjectAndNotes("", "")
	if !ok {
		return
	}

	idea := config.Idea{
		Text:    text,
		Date:    time.Now().Format(time.RFC3339),
		Project: project,
		Notes:   notes,
	}
	ideas := config.LoadIdeas()
	ideas = append(ideas, idea)
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

func cmdUnparkInteractive() {
	ideas := config.LoadIdeas()
	if len(ideas) == 0 {
		fmt.Println()
		StepDim(i18n.Tr("ideas.empty_short"))
		fmt.Println()
		return
	}

	options := make([]huh.Option[string], 0, len(ideas))
	for idx, idea := range ideas {
		options = append(options, huh.NewOption(ideaLabel(idx, idea), strconv.Itoa(idx+1)))
	}

	var selected string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(i18n.Tr("unpark.pick")).
				Options(options...).
				Value(&selected),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil || selected == "" {
		return
	}
	cmdUnpark(selected)
}

func cmdStatusInteractive() {
	for {
		project, ok := pickProject("status.pick_project")
		if !ok {
			return
		}

		var status string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(i18n.Tr("status.pick_status", project)).
					Options(
						huh.NewOption(i18n.Tr("status.opt_active"), statusActive),
						huh.NewOption(i18n.Tr("status.opt_done"), statusDone),
						huh.NewOption(i18n.Tr("status.opt_paused"), statusPaused),
						huh.NewOption(i18n.Tr("status.opt_dead"), statusDead),
						huh.NewOption(i18n.Tr("status.opt_idea"), statusIdea),
						huh.NewOption(i18n.Tr("back"), actionBack),
					).
					Value(&status),
			),
		).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMapBack()).Run()

		if err != nil || status == "" || status == actionBack {
			continue
		}
		cmdStatus(project, status)
		return
	}
}

func cmdEditInteractive() {
	selected, ok := pickProject("edit.pick")
	if !ok {
		return
	}
	cmdEdit(selected)
}

func cmdInfoInteractive() {
	selected, ok := pickProject("info.pick")
	if !ok {
		return
	}
	cmdInfo(selected)
}
