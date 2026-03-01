package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/goodmartian/pulse-go/internal/config"
	"github.com/goodmartian/pulse-go/internal/i18n"
)

const (
	fieldTagline  = 0
	fieldStack    = 1
	fieldDoneWhen = 2
	fieldDesc     = 3
	fieldNotes    = 4
	btnSave       = 5
	btnCancel     = 6
)

type editorModel struct {
	project string
	inputs  []textinput.Model
	areas   []textarea.Model
	focus   int
	saved   bool
	width   int
	height  int
}

var editorLabelKeys = []struct {
	key   string
	field int
}{
	{"editor.label_tagline", fieldTagline},
	{"editor.label_stack", fieldStack},
	{"editor.label_done_when", fieldDoneWhen},
	{"editor.label_description", fieldDesc},
	{"editor.label_notes", fieldNotes},
}

var barDimStyle = lipgloss.NewStyle().Foreground(DimC).Padding(0, 1)

func newEditor(name string) editorModel {
	pd := config.LoadProjectData(name)

	tagline := textinput.New()
	tagline.Placeholder = i18n.Tr("editor.placeholder_tagline")
	tagline.SetValue(pd.Tagline)
	tagline.CharLimit = 120
	tagline.Width = 60
	tagline.Focus()

	stack := textinput.New()
	stack.Placeholder = i18n.Tr("editor.placeholder_stack")
	stack.SetValue(pd.Stack)
	stack.CharLimit = 80
	stack.Width = 60

	doneWhen := textinput.New()
	doneWhen.Placeholder = i18n.Tr("editor.placeholder_done")
	doneWhen.SetValue(pd.DoneWhen)
	doneWhen.CharLimit = 200
	doneWhen.Width = 60

	desc := textarea.New()
	desc.Placeholder = i18n.Tr("editor.placeholder_desc")
	desc.SetValue(pd.Description)
	desc.SetWidth(64)
	desc.SetHeight(6)
	desc.CharLimit = 4000
	desc.ShowLineNumbers = false

	notes := textarea.New()
	notes.Placeholder = i18n.Tr("editor.placeholder_notes")
	notes.SetValue(pd.Notes)
	notes.SetWidth(64)
	notes.SetHeight(4)
	notes.CharLimit = 4000
	notes.ShowLineNumbers = false

	return editorModel{
		project: name,
		inputs:  []textinput.Model{tagline, stack, doneWhen},
		areas:   []textarea.Model{desc, notes},
		focus:   fieldTagline,
	}
}

func (m editorModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m editorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w := msg.Width - 8
		if w < 30 {
			w = 30
		}
		if w > 80 {
			w = 80
		}
		for i := range m.inputs {
			m.inputs[i].Width = w
		}
		for i := range m.areas {
			m.areas[i].SetWidth(w + 4)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			m.saved = true
			return m, tea.Quit

		case "esc":
			m.saved = false
			return m, tea.Quit

		case "tab", "shift+tab":
			if msg.String() == "tab" {
				m.focus++
				if m.focus > btnCancel {
					m.focus = fieldTagline
				}
			} else {
				m.focus--
				if m.focus < fieldTagline {
					m.focus = btnCancel
				}
			}
			m.updateFocus()
			return m, nil

		case "enter":
			if m.focus == fieldDesc || m.focus == fieldNotes {
				break
			}
			if m.focus == btnSave {
				m.saved = true
				return m, tea.Quit
			}
			if m.focus == btnCancel {
				m.saved = false
				return m, tea.Quit
			}
			m.focus++
			m.updateFocus()
			return m, nil

		case "up":
			if m.focus < fieldDesc {
				m.focus--
				if m.focus < fieldTagline {
					m.focus = fieldTagline
				}
				m.updateFocus()
				return m, nil
			}

		case "down":
			if m.focus < fieldDesc {
				m.focus++
				m.updateFocus()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case fieldTagline, fieldStack, fieldDoneWhen:
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	case fieldDesc:
		m.areas[0], cmd = m.areas[0].Update(msg)
	case fieldNotes:
		m.areas[1], cmd = m.areas[1].Update(msg)
	}
	return m, cmd
}

func (m *editorModel) updateFocus() {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	for i := range m.areas {
		m.areas[i].Blur()
	}

	switch m.focus {
	case fieldTagline, fieldStack, fieldDoneWhen:
		m.inputs[m.focus].Focus()
	case fieldDesc:
		m.areas[0].Focus()
	case fieldNotes:
		m.areas[1].Focus()
	}
}

func (m editorModel) View() string {
	var b strings.Builder

	header := TitleStyle.Render(i18n.Tr("editor.header", m.project))
	b.WriteString("  " + header + "\n\n")

	for _, l := range editorLabelKeys {
		label := DimStyle.Render(i18n.Tr(l.key))
		if m.focus == l.field {
			label = BoldStyle.Render(i18n.Tr(l.key))
		}
		b.WriteString("    " + label + "\n")

		switch l.field {
		case fieldTagline, fieldStack, fieldDoneWhen:
			b.WriteString("    " + m.inputs[l.field].View() + "\n")
		case fieldDesc:
			b.WriteString(indentBlock(m.areas[0].View(), "    ") + "\n")
		case fieldNotes:
			b.WriteString(indentBlock(m.areas[1].View(), "    ") + "\n")
		}
		b.WriteString("\n")
	}

	save := DimStyle.Render(i18n.Tr("editor.save"))
	cancel := DimStyle.Render(i18n.Tr("editor.cancel"))
	if m.focus == btnSave {
		save = GreenStyle.Bold(true).Render(i18n.Tr("editor.save"))
	}
	if m.focus == btnCancel {
		cancel = RedStyle.Bold(true).Render(i18n.Tr("editor.cancel"))
	}
	b.WriteString(fmt.Sprintf("    %s      %s\n", save, cancel))

	bar := i18n.Tr("editor.bar")
	if m.focus == fieldDesc || m.focus == fieldNotes {
		bar = i18n.Tr("editor.bar_textarea")
	}
	b.WriteString("\n" + barDimStyle.Render(bar))

	return b.String()
}

func indentBlock(s string, pad string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}

func (m editorModel) projectData() config.ProjectData {
	return config.ProjectData{
		Name:        m.project,
		Tagline:     m.inputs[fieldTagline].Value(),
		Stack:       m.inputs[fieldStack].Value(),
		DoneWhen:    m.inputs[fieldDoneWhen].Value(),
		Description: m.areas[0].Value(),
		Notes:       m.areas[1].Value(),
	}
}

func cmdEdit(name string) {
	if err := config.ValidateProjectName(name); err != nil {
		fmt.Printf("\n  %s\n\n", RedStyle.Render(err.Error()))
		return
	}
	projectPath := filepath.Join(config.ProjectsDir, name)
	if _, err := os.Stat(projectPath); err != nil {
		fmt.Printf("\n  %s\n\n", RedStyle.Render(i18n.Tr("editor.not_found", name)))
		return
	}

	m := newEditor(name)
	p := tea.NewProgram(&m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("\n  %s\n\n", RedStyle.Render(i18n.Tr("editor.tui_error", err.Error())))
		return
	}

	if fm, ok := finalModel.(*editorModel); ok && fm.saved {
		if err := config.SaveProjectData(name, fm.projectData()); err != nil {
			fmt.Printf("\n  %s\n\n", RedStyle.Render(i18n.Tr("error.save_failed")+err.Error()))
			return
		}
		fmt.Printf("\n  %s\n\n", GreenStyle.Bold(true).Render(i18n.Tr("editor.saved", name)))
	} else {
		fmt.Println(DimStyle.Render("\n  " + i18n.Tr("editor.cancelled") + "\n"))
	}
}
