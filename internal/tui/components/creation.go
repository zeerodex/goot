package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeerodex/goot/pkg/timeutil"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type CreationModel struct {
	focusIndex int
	inputs     []textinput.Model

	Done   bool
	Result Task
}

type Task struct {
	Title       string
	Description string
	Due         time.Time
}

func InitialCreationModel() CreationModel {
	m := CreationModel{
		inputs: make([]textinput.Model, 3),
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		switch i {
		case 0:
			t.Focus()
			t.Placeholder = "Title (Required)"
			t.CharLimit = 156
			t.Width = 50
		case 1:
			t.Placeholder = "Description (Not required)"
			t.CharLimit = 156
			t.Width = 50
		case 2:
			t.Placeholder = "Due [YYYY-MM-DD (HH:MM)] (Today by default)"
			t.CharLimit = 156
			t.Width = 50

		}
		m.inputs[i] = t
	}

	return m
}

func (m CreationModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				due, err := timeutil.ParseAndValidateTimestamp(m.inputs[2].Value())
				if m.inputs[0].Value() == "" {
					m.inputs[0].Placeholder = "Task title cannot be empty"
				} else if err != nil {
					m.inputs[2].SetValue("")
					m.inputs[2].Placeholder = err.Error()
				} else {
					m.Done = true

					m.Result = Task{
						Title:       m.inputs[0].Value(),
						Description: m.inputs[1].Value(),
						Due:         due,
					}
					return m, nil
				}
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}
	cmd = m.updateInputs(msg)

	return m, cmd
}

func (m *CreationModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m CreationModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}
