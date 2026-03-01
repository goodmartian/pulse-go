package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/goodmartian/pulse-go/internal/config"
	"github.com/goodmartian/pulse-go/internal/git"
	"github.com/goodmartian/pulse-go/internal/i18n"
)

const (
	actionCreate = "_create"
	actionBack   = "_back"

	statusActive = "active"
	statusDone   = "done"
	statusPaused = "paused"
	statusDead   = "dead"
	statusIdea   = "idea"
)

func Run(args []string) {
	if len(args) == 0 {
		cmdDashboard()
		return
	}

	switch args[0] {
	case "help", "--help", "-h":
		cmdHelp()
	case "map":
		cmdMap()
	case "focus":
		if len(args) < 2 {
			cmdFocusInteractive()
			return
		}
		cmdFocus(args[1])
	case "edit":
		if len(args) < 2 {
			cmdEditInteractive()
			return
		}
		cmdEdit(args[1])
	case "info":
		if len(args) < 2 {
			cmdInfoInteractive()
			return
		}
		cmdInfo(args[1])
	case "park":
		text := strings.Join(args[1:], " ")
		if text == "" {
			cmdParkInteractive()
			return
		}
		cmdPark(text)
	case "ideas":
		if len(args) > 1 && args[1] == "list" {
			cmdIdeas()
		} else {
			cmdIdeasInteractive()
		}
	case "unpark":
		if len(args) < 2 {
			cmdUnparkInteractive()
			return
		}
		cmdUnpark(args[1])
	case "log":
		period := "today"
		if len(args) > 1 {
			period = args[1]
		}
		cmdLog(period)
	case "status":
		if len(args) < 3 {
			cmdStatusInteractive()
			return
		}
		cmdStatus(args[1], args[2])
	case "switches":
		cmdSwitches()
	case "lang":
		cmdLang(args[1:])
	case "setup":
		CmdSetup()
	case "scan":
		projects := git.ScanProjects(config.ProjectsDir)
		fmt.Println()
		Step(i18n.Tr("scan.found", len(projects)))
		fmt.Println()
		cmdMap()
	default:
		fmt.Println()
		StepError(i18n.Tr("error.unknown_cmd") + args[0])
		cmdHelp()
	}
}

func CmdSetup() {
	StepHeader("◈ PULSE — setup")

	home, _ := os.UserHomeDir()
	defaultDir := filepath.Join(home, "Projects")

	cfg := config.LoadConfig()
	dir := cfg.ProjectsDir
	if dir == "" {
		dir = defaultDir
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(i18n.Tr("setup.projects_dir")).
				Description(i18n.Tr("setup.projects_dir_desc")).
				Value(&dir),
		),
	).WithTheme(PulseTheme()).WithKeyMap(PulseKeyMap()).Run()

	if err != nil || strings.TrimSpace(dir) == "" {
		return
	}
	dir = strings.TrimSpace(dir)

	if strings.HasPrefix(dir, "~/") {
		dir = filepath.Join(home, dir[2:])
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		StepError(err.Error())
		return
	}

	cfg.ProjectsDir = dir
	if err := config.SaveConfig(cfg); err != nil {
		StepError(i18n.Tr("error.save_failed") + err.Error())
		return
	}

	config.Init()

	fmt.Println()
	Step(i18n.Tr("setup.saved") + " " + GreenStyle.Render(dir))
	fmt.Println()
}

func cmdHelp() {
	StepHeader(i18n.Tr("help.title"))
	Step(i18n.Tr("help.commands"))
	cmds := [][2]string{
		{"pulse", i18n.Tr("help.cmd_dashboard")},
		{"pulse map", i18n.Tr("help.cmd_map")},
		{"pulse focus [project]", i18n.Tr("help.cmd_focus")},
		{"pulse edit [project]", i18n.Tr("help.cmd_edit")},
		{"pulse info [project]", i18n.Tr("help.cmd_info")},
		{i18n.Tr("help.park_syntax"), i18n.Tr("help.cmd_park")},
		{"pulse ideas", i18n.Tr("help.cmd_ideas")},
		{"pulse ideas list", i18n.Tr("help.cmd_ideas_list")},
		{"pulse unpark [N]", i18n.Tr("help.cmd_unpark")},
		{"pulse log", i18n.Tr("help.cmd_log")},
		{"pulse log week", i18n.Tr("help.cmd_log_week")},
		{"pulse status [p] [s]", i18n.Tr("help.cmd_status")},
		{"pulse switches", i18n.Tr("help.cmd_switches")},
		{"pulse lang [en|ru]", i18n.Tr("help.cmd_lang")},
		{"pulse setup", i18n.Tr("help.cmd_setup")},
	}
	for _, c := range cmds {
		StepResult(fmt.Sprintf("%s  %s", CyanStyle.Render(fmt.Sprintf("%-24s", c[0])), c[1]))
	}
	StepBlank()
	StepDim(i18n.Tr("help.philosophy"))
	for _, key := range []string{"help.p1", "help.p2", "help.p3", "help.p4"} {
		StepResult("• " + i18n.Tr(key))
	}
	fmt.Println()
}

func cmdLang(args []string) {
	if len(args) == 0 {
		fmt.Println()
		Step(i18n.Tr("lang.current", i18n.Lang))
		fmt.Println()
		return
	}
	lang := args[0]
	if lang != "en" && lang != "ru" {
		fmt.Println()
		StepError(i18n.Tr("lang.invalid", lang))
		fmt.Println()
		return
	}
	cfg := config.LoadConfig()
	cfg.Lang = lang
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Println()
		StepError(i18n.Tr("error.save_failed") + err.Error())
		fmt.Println()
		return
	}
	i18n.SetLang(lang)
	fmt.Println()
	Step(i18n.Tr("lang.switched", i18n.Lang))
	fmt.Println()
}
