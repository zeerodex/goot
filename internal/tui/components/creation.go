package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type CreationModel struct {
	TaskInput textinput.Model
	DescInput textinput.Model
	DueInput  textinput.Model

	Step   int
	Done   bool
	Result Task
}

type Task struct {
	Title       string
	Description string
	Due         string
}

func InitialCreationModel() CreationModel {
	taskInput := textinput.New()
	taskInput.Focus()
	taskInput.Placeholder = "Title (Required)"
	taskInput.CharLimit = 156
	taskInput.Width = 20

	descInput := textinput.New()
	descInput.Placeholder = "Description (Not required)"
	descInput.CharLimit = 156
	descInput.Width = 40

	dueInput := textinput.New()
	dueInput.Placeholder = "Due [YYYY-MM-DD (HH:MM)] (Today by default)"
	dueInput.CharLimit = 16
	dueInput.Width = 40

	return CreationModel{
		TaskInput: taskInput,
		DescInput: descInput,
		DueInput:  dueInput,
		Step:      1,
	}
}

func (m CreationModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.Step == 1 {
				if m.TaskInput.Value() == "" {
					m.TaskInput.Placeholder = "Task title cannot be empty"
				} else {
					m.TaskInput.Blur()
					m.DescInput.Focus()

					m.Step = 2
					return m, textinput.Blink
				}
			} else if m.Step == 2 {
				m.DescInput.Blur()
				m.DueInput.Focus()

				m.Step = 3
				return m, textinput.Blink
			} else if m.Step == 3 {
				m.Done = true

				m.Result = Task{
					Title:       m.TaskInput.Value(),
					Description: m.DescInput.Value(),
					Due:         m.DueInput.Value(),
				}
			}
		}
	}

	switch m.Step {
	case 1:
		m.TaskInput, cmd = m.TaskInput.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		m.DescInput, cmd = m.DescInput.Update(msg)
		cmds = append(cmds, cmd)
	case 3:
		m.DescInput, cmd = m.DueInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m CreationModel) View() string {
	var s string = fmt.Sprintf(
		"What's your task?\n\n%s",
		m.TaskInput.View(),
	) + "\n"

	if m.Step == 2 {
		s += fmt.Sprintf(
			"Enter a description:\n\n%s",
			m.DescInput.View()) + "\n\n"
		s += "Press Enter to save"
	}

	return s
}
