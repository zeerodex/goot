package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
	"github.com/zeerodex/go-todo-tui/internal/tui/components"
)

type AppState int

const (
	TableView AppState = iota
)

type AppModel struct {
	state      AppState
	tableModel components.TableModel
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			m.state = TableView
			return m, nil
		}

		switch m.state {
		case TableView:
			newState, cmd := m.tableModel.Update(msg)
			m.tableModel = newState.(components.TableModel)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m AppModel) View() string {
	switch m.state {
	case TableView:
		return m.tableModel.View()
	default:
		return `"t" - table view\nPress "q" to exit\n`
	}
}

func InitAppModel(s *tasks.TaskService) *AppModel {
	tableModel := components.InitTableModel(s)
	return &AppModel{tableModel: *tableModel}
}
