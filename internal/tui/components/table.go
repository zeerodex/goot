package components

import (
	"fmt"
	"os"
	"strconv"

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

	selected int
	method   string
	err      error
}

func (m TableModel) Init() tea.Cmd {
	return nil
}

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
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[0]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TableModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func InitialTableModel(service tasks.TaskService) TableModel {
	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Task", Width: 25},
		{Title: "Description", Width: 15},
		{Title: "Status", Width: 15},
	}

	tasks, err := service.All()
	if len(tasks) < 1 {
		fmt.Println("No tasks")
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var rows []table.Row
	for _, task := range tasks {
		status := "Incompleted"
		if task.Status {
			status = "Completed"
		}
		row := table.Row{strconv.Itoa(task.ID), task.Task, task.Description, status}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(10), // Set the height of the table
	)

	return TableModel{table: t}
}
