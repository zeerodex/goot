package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
	"github.com/zeerodex/go-todo-tui/internal/tui/components"
)

type AppState int

const (
	MainView     AppState = iota
	CreationView AppState = iota
	TableView    AppState = iota
)

type AppModel struct {
	state         AppState
	tableModel    components.TableModel
	creationModel components.CreationModel

	service tasks.TaskService
}

type TaskCompletedMsg struct {
	Task components.Task
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
			if m.state == MainView {
				return m, tea.Quit
			} else if m.state == TableView {
				m.state = MainView
				return m, nil
			}
		case "esc":
			m.state = MainView
			return m, nil
		case "t":
			if m.state == MainView {
				m.state = TableView
				return m, nil
			}
		case "c":
			if m.state == MainView {
				m.state = CreationView
				return m, nil
			}
		}

	case TaskCompletedMsg:
		err := m.service.Create(msg.Task.Title, msg.Task.Description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		m.tableModel = components.InitialTableModel(m.service)
		m.creationModel = components.InitialCreationModel()

		m.state = MainView
		return m, nil
	}

	switch m.state {
	case TableView:
		tableModel, tableCmd := m.tableModel.Update(msg)
		m.tableModel = tableModel.(components.TableModel)
		cmds = append(cmds, tableCmd)
	case CreationView:
		creationModel, creationCmd := m.creationModel.Update(msg)
		m.creationModel = creationModel.(components.CreationModel)
		cmds = append(cmds, creationCmd)

		if m.creationModel.Done {
			return m, func() tea.Msg {
				return TaskCompletedMsg{Task: m.creationModel.Result}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	switch m.state {
	case TableView:
		return m.tableModel.View()
	case CreationView:
		return m.creationModel.View()
	case MainView:
		return "Press t - table view\nPress c - create task\nPress q to exit\n"
	default:
		return "Press t - table view\nPress q to exit\n"
	}
}

func InitialAppModel(service tasks.TaskService) AppModel {
	return AppModel{
		state:         MainView,
		tableModel:    components.InitialTableModel(service),
		creationModel: components.InitialCreationModel(),

		service: service,
	}
}
