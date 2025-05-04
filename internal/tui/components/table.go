package components

import (
	"errors"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type TableModel struct {
	table table.Model

	Selected string
	Method   string
	Err      error
}

func (m TableModel) Init() tea.Cmd { return nil }

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.table.MoveUp(1)
		case "down", "j":
			m.table.MoveDown(1)
		case "x":
			task := m.table.SelectedRow()[0]
			m.Method = "delete"
			m.Selected = task
		}
	}

	return m, cmd
}

func (m TableModel) View() string {
	if m.Err != nil {
		return "No tasks!\nAny key - create task\nEsc - main menu\n"
	}
	return baseStyle.Render(m.table.View()) + "\n"
}

func InitialTableModel(repo tasks.TaskRepository) TableModel {
	var tableModel TableModel

	columns := []table.Column{
		{Title: "Task", Width: 25},
		{Title: "Description", Width: 15},
		{Title: "Status", Width: 15},
	}
	tableModel.Err = nil

	tasks, err := repo.GetAll()
	if len(tasks) < 1 {
		tableModel.Err = errors.New("no tasks")
	}
	if err != nil {
		tableModel.Err = err
	}

	var rows []table.Row
	for _, task := range tasks {
		status := "Incompleted"
		if task.Status {
			status = "Completed"
		}
		row := table.Row{task.Task, task.Description, status}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(10), // Set the height of the table
	)
	tableModel.table = t
	return tableModel
}
