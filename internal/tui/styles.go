package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/goodmartian/pulse-go/internal/git"
)

var (
	Cyan    = lipgloss.Color("6")
	Green   = lipgloss.Color("2")
	Yellow  = lipgloss.Color("3")
	Red     = lipgloss.Color("1")
	Magenta = lipgloss.Color("5")
	Blue    = lipgloss.Color("4")
	DimC    = lipgloss.Color("8")
	White   = lipgloss.Color("15")

	TitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(Cyan)
	BoldStyle    = lipgloss.NewStyle().Bold(true)
	DimStyle     = lipgloss.NewStyle().Foreground(DimC)
	GreenStyle   = lipgloss.NewStyle().Foreground(Green)
	YellowStyle  = lipgloss.NewStyle().Foreground(Yellow)
	RedStyle     = lipgloss.NewStyle().Foreground(Red)
	MagentaStyle = lipgloss.NewStyle().Foreground(Magenta)
	CyanStyle    = lipgloss.NewStyle().Foreground(Cyan)

	badgeActive = lipgloss.NewStyle().Bold(true).Foreground(Green)
	badgeDone   = lipgloss.NewStyle().Bold(true).Foreground(Blue)
	badgePaused = lipgloss.NewStyle().Foreground(Yellow)
	badgeDead   = lipgloss.NewStyle().Foreground(Red)
	badgeIdea   = lipgloss.NewStyle().Foreground(Magenta)
)

func statusBadge(lastDate string, override string) string {
	if override != "" {
		switch override {
		case "active":
			return badgeActive.Render("● active")
		case "done":
			return badgeDone.Render("● done")
		case "paused":
			return badgePaused.Render("○ paused")
		case "dead":
			return badgeDead.Render("✕ dead")
		case "idea":
			return badgeIdea.Render("◇ idea")
		}
	}

	days := git.DaysAgo(lastDate)
	if days < 0 {
		return DimStyle.Render("no git")
	}
	if days <= 3 {
		return GreenStyle.Render("● hot")
	}
	if days <= 14 {
		return YellowStyle.Render("○ warm")
	}
	if days <= 60 {
		return RedStyle.Render("◌ cold")
	}
	return DimStyle.Render("✕ frozen")
}
