package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
	"github.com/zeerodex/go-todo-tui/internal/tui/components"
)

type AppState int

const (
	MainView AppState = iota
	ListView
	CreationView
)

type MainModel struct {
	State         AppState
	listModel     components.ListModel
	creationModel components.CreationModel

	tasks tasks.Tasks
	repo  tasks.TaskRepository
	err   error
}

func fetchTasks(repo tasks.TaskRepository) tea.Cmd {
	return func() tea.Msg {
		tasks, err := repo.GetAll()
		if err != nil {
			return errMsg{err: err}
		}
		return FetchedTasksMsg{Tasks: tasks}
	}
}

func createTaskCmd(repo tasks.TaskRepository, task components.Task) tea.Cmd {
	return func() tea.Msg {
		err := repo.Create(task.Title, task.Description)
		if err != nil {
			return errMsg{err: err}
		}
		return fetchTasks(repo)()
	}
}

type FetchedTasksMsg struct {
	Tasks tasks.Tasks
}

type errMsg struct {
	err error
}

type TaskCompletedMsg struct {
	Task components.Task
}

func (m MainModel) Init() tea.Cmd {
	return fetchTasks(m.repo)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.State == MainView {
				return m, tea.Quit
			} else if m.State == ListView {
				m.State = MainView
				return m, nil
			}
		case "l":
			if m.State == MainView {
				m.State = ListView
				return m, nil
			}
		case "c":
			if m.State == MainView {
				m.State = CreationView
				return m, nil
			}
		}

	case FetchedTasksMsg:
		m.tasks = msg.Tasks
		cmds = append(cmds, m.listModel.SetTasks(m.tasks))

	case TaskCompletedMsg:
		cmds = append(cmds, createTaskCmd(m.repo, msg.Task))
		m.creationModel = components.InitialCreationModel()

		m.State = MainView

	case errMsg:
		m.err = msg.err
	}

	switch m.State {
	case ListView:
		listModel, listCmd := m.listModel.Update(msg)
		m.listModel = listModel.(components.ListModel)
		cmds = append(cmds, listCmd)

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

func (m MainModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	switch m.State {
	case ListView:
		return m.listModel.View()
	case CreationView:
		return m.creationModel.View()
	case MainView:
		return "Press l - list view\nPress c - create task\nPress q to exit\n"
	default:
		return "Press l - list view\nPress q to exit\n"
	}
}

func InitialMainModel(repo tasks.TaskRepository) MainModel {
	listModel := components.InitialListModel(tasks.Tasks{})
	creationModel := components.InitialCreationModel()

	m := MainModel{
		State:         MainView,
		listModel:     listModel,
		creationModel: creationModel,

		repo: repo,
	}

	return m
}
