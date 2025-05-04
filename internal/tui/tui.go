package tui

import (
	"errors"
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
	ErrView      AppState = iota
)

type AppModel struct {
	State         AppState
	tableModel    components.TableModel
	creationModel components.CreationModel

	repo tasks.TaskRepository
	Err  errMsg
}

type errMsg error

type TaskCompletedMsg struct {
	Task components.Task
}

type NoTasksMsg error

type TaskDeleteMsg struct {
	id int
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
			if m.State == MainView {
				return m, tea.Quit
			} else if m.State == TableView {
				m.State = MainView
				return m, nil
			}
		case "esc":
			m.State = MainView
			return m, nil
		case "t":
			if m.State == MainView {
				m.State = TableView
				return m, nil
			}
		case "c":
			if m.State == MainView {
				m.State = CreationView
				return m, nil
			}
		}
	case TaskDeleteMsg:
		err := m.repo.DeleteByID(msg.id)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		m.tableModel = components.InitialTableModel(m.repo)

		return m, nil

	case TaskCompletedMsg:
		err := m.repo.Create(msg.Task.Title, msg.Task.Description)
		if err != nil {
			return m, func() tea.Msg {
				return err
			}
		}

		m.tableModel = components.InitialTableModel(m.repo)
		m.creationModel = components.InitialCreationModel()

		m.State = MainView
		return m, nil

	case NoTasksMsg:
		m.State = CreationView
		return m, nil

	case errMsg:
		m.Err = msg
		m.State = ErrView
		return m, nil
	}

	switch m.State {
	case TableView:
		tableModel, tableCmd := m.tableModel.Update(msg)
		m.tableModel = tableModel.(components.TableModel)
		cmds = append(cmds, tableCmd)

		if m.tableModel.Err != nil {
			if m.tableModel.Err.Error() == "no tasks" {
				return m, func() tea.Msg {
					return NoTasksMsg(errors.New("no tasks"))
				}
			}
		}

		if m.tableModel.Method == "delete" {
			err := m.repo.DeleteByTitle(m.tableModel.Selected)
			if err != nil {
				return m, func() tea.Msg {
					return errMsg(m.Err)
				}
			}

			m.tableModel = components.InitialTableModel(m.repo)

			return m, nil

		}

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
	switch m.State {
	case TableView:
		return m.tableModel.View()
	case CreationView:
		return m.creationModel.View()
	case MainView:
		return "Press t - table view\nPress c - create task\nPress q to exit\n"
	case ErrView:
		return m.Err.Error()
	default:
		return "Press t - table view\nPress q to exit\n"
	}
}

func InitialAppModel(repo tasks.TaskRepository) AppModel {
	return AppModel{
		State:         MainView,
		tableModel:    components.InitialTableModel(repo),
		creationModel: components.InitialCreationModel(),

		repo: repo,
	}
}
