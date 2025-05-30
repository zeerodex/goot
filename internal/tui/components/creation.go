package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/pkg/timeutil"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type CreationModel struct {
	focusIndex int
	inputs     []textinput.Model

	Done   bool
	Task   *tasks.Task
	Method string
}

func InitialCreationModel() CreationModel {
	m := CreationModel{
		inputs: make([]textinput.Model, 3),
	}
	task := &tasks.Task{}
	m.Task = task
	m.Method = "create"
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		switch i {
		case 0:
			t.Focus()
			t.Placeholder = "Title"
			t.CharLimit = 1024
			t.Width = 50
		case 1:
			t.Placeholder = "Description (Not required)"
			t.CharLimit = 8192
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

func InitialUpdateModel(task *tasks.Task) CreationModel {
	m := CreationModel{
		inputs: make([]textinput.Model, 3),
	}
	m.Method = "update"
	m.Task = task
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		switch i {
		case 0:
			t.Focus()
			t.Placeholder = "Title"
			t.CharLimit = 1024
			t.Width = 50
			t.SetValue(m.Task.Title)
		case 1:
			t.Placeholder = "Description (Not required)"
			t.CharLimit = 8192
			t.Width = 50
			t.SetValue(m.Task.Description)
		case 2:
			t.Placeholder = "Due [YYYY-MM-DD (HH:MM)] (Today by default)"
			t.CharLimit = 156
			t.Width = 50
			t.SetValue(m.Task.DueStr())

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
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				due, err := timeutil.ParseAndValidateTimestamp(m.inputs[2].Value())
				if m.inputs[0].Value() == "" {
					m.inputs[0].Placeholder = "Task title cannot be empty"
				} else if len(m.inputs[0].Value()) > 1024 {
					m.inputs[0].SetValue("")
					m.inputs[0].Placeholder = "Length of title is up to 1024 characters"
				} else if len(m.inputs[1].Value()) > 8192 {
					m.inputs[0].SetValue("")
					m.inputs[0].Placeholder = "Length of description is up to 1024 characters"
				} else if err != nil {
					m.inputs[2].SetValue("")
					m.inputs[2].Placeholder = strings.ToUpper(err.Error()[:1]) + err.Error()[1:]
				} else {
					m.Done = true

					m.Task.Title = m.inputs[0].Value()
					m.Task.Description = m.inputs[1].Value()
					m.Task.Due = due

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
