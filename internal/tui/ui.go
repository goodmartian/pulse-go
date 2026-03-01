package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/goodmartian/pulse-go/internal/i18n"
)

const (
	markerFilled = "◆"
	markerEmpty  = "◇"
	markerDot    = "●"
	connector    = "│"
)

func Step(text string) {
	fmt.Printf("  %s  %s\n", GreenStyle.Render(markerFilled), text)
}

func StepDim(text string) {
	fmt.Printf("  %s  %s\n", DimStyle.Render(markerEmpty), DimStyle.Render(text))
}

func StepResult(text string) {
	fmt.Printf("  %s  %s\n", DimStyle.Render(connector), DimStyle.Render(text))
}

func StepResultStyled(text string, style lipgloss.Style) {
	fmt.Printf("  %s  %s\n", DimStyle.Render(connector), style.Render(text))
}

func StepBlank() {
	fmt.Printf("  %s\n", DimStyle.Render(connector))
}

func StepError(text string) {
	fmt.Printf("  %s  %s\n", RedStyle.Render(markerFilled), RedStyle.Render(text))
}

func StepHeader(text string) {
	fmt.Printf("\n  %s\n\n", TitleStyle.Render(text))
}

func PulseKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("esc", "ctrl+c"))
	return km
}

func PulseKeyMapBack() *huh.KeyMap {
	km := PulseKeyMap()
	km.Select.ClearFilter = key.NewBinding(
		key.WithKeys("ctrl+esc"),
		key.WithHelp("esc", i18n.Tr("back.hint")),
	)
	return km
}

func PulseTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Title = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
	t.Focused.Description = lipgloss.NewStyle().Foreground(DimC)
	t.Focused.SelectSelector = lipgloss.NewStyle().Foreground(Green).SetString(markerDot + " ")
	t.Focused.UnselectedOption = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	t.Focused.SelectedOption = lipgloss.NewStyle().Foreground(Green).Bold(true)
	t.Focused.FocusedButton = lipgloss.NewStyle().Foreground(White).Background(Green).Bold(true).Padding(0, 1)
	t.Focused.BlurredButton = lipgloss.NewStyle().Foreground(DimC).Padding(0, 1)
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(Green)
	t.Focused.TextInput.Prompt = lipgloss.NewStyle().Foreground(Green).SetString(markerFilled + " ")

	t.Blurred.Title = lipgloss.NewStyle().Foreground(DimC)
	t.Blurred.SelectSelector = lipgloss.NewStyle().Foreground(DimC).SetString("  ")
	t.Blurred.TextInput.Prompt = lipgloss.NewStyle().Foreground(DimC).SetString(markerEmpty + " ")

	return t
}
