package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/internal/tui/components"
)

type AppState int

const (
	MainView AppState = iota
	ListView
	CreationView
	ErrView
)

type MainModel struct {
	currentState  AppState
	previuosState AppState

	listModel     components.ListModel
	creationModel components.CreationModel

	tasks tasks.Tasks
	repo  tasks.TaskRepository
	err   error
}

func fetchTasksCmd(repo tasks.TaskRepository) tea.Cmd {
	return func() tea.Msg {
		tasks, err := repo.GetAll()
		if err != nil {
			return errMsg{err: err}
		}
		return fetchedTasksMsg{Tasks: tasks}
	}
}

func createTaskCmd(repo tasks.TaskRepository, task components.Task) tea.Cmd {
	return func() tea.Msg {
		err := repo.Create(task.Title, task.Description, task.Due)
		if err != nil {
			return errMsg{err: err}
		}
		return fetchTasksCmd(repo)()
	}
}

func deleteTaskCmd(repo tasks.TaskRepository, id int) tea.Cmd {
	return func() tea.Msg {
		err := repo.DeleteByID(id)
		if err != nil {
			return errMsg{err: err}
		}
		return fetchTasksCmd(repo)()
	}
}

func toogleTaskCmd(repo tasks.TaskRepository, id int, completed bool) tea.Cmd {
	return func() tea.Msg {
		err := repo.Toogle(id, completed)
		if err != nil {
			return errMsg{err: err}
		}
		return fetchTasksCmd(repo)()
	}
}

type fetchedTasksMsg struct {
	Tasks tasks.Tasks
}

type deleteTaskMsg struct {
	id int
}

type toogleTaskMsg struct {
	id        int
	completed bool
}

type createTaskMsg struct {
	Task components.Task
}

type errMsg struct {
	err error
}

func (m MainModel) Init() tea.Cmd {
	return fetchTasksCmd(m.repo)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.currentState == MainView {
				return m, tea.Quit
			} else if m.currentState == ListView {
				m.currentState = MainView
				return m, nil
			} else if m.currentState == ErrView {
				m.currentState = m.previuosState
				return m, nil
			}
		case "l":
			if m.currentState == MainView {
				m.currentState = ListView
				return m, nil
			}
		case "c":
			if m.currentState == MainView {
				m.currentState = CreationView
				return m, nil
			}
		}

	case deleteTaskMsg:
		cmds = append(cmds, deleteTaskCmd(m.repo, msg.id))

	case toogleTaskMsg:
		cmds = append(cmds, toogleTaskCmd(m.repo, msg.id, msg.completed))

	case fetchedTasksMsg:
		m.tasks = msg.Tasks
		cmds = append(cmds, m.listModel.SetTasks(m.tasks))

	case createTaskMsg:
		cmds = append(cmds, createTaskCmd(m.repo, msg.Task))
		m.creationModel = components.InitialCreationModel()

		m.currentState = m.previuosState

	case errMsg:
		m.err = msg.err
		m.previuosState = m.currentState
		m.currentState = ErrView
	}

	switch m.currentState {
	case ListView:
		listModel, listCmd := m.listModel.Update(msg)
		m.listModel = listModel.(components.ListModel)
		cmds = append(cmds, listCmd)

		if m.listModel.Method == "delete" {
			m.listModel.Method = ""
			cmds = append(cmds, func() tea.Msg {
				return deleteTaskMsg{id: m.listModel.Selected.ID()}
			})
		}
		if m.listModel.Method == "create" {
			m.listModel.Method = ""
			m.previuosState = m.currentState
			m.currentState = CreationView
		}
		if m.listModel.Method == "toogle" {
			m.listModel.Method = ""
			cmds = append(cmds, func() tea.Msg {
				return toogleTaskMsg{id: m.listModel.Selected.ID(), completed: m.listModel.Selected.Completed()}
			})

		}

	case CreationView:
		creationModel, creationCmd := m.creationModel.Update(msg)
		m.creationModel = creationModel.(components.CreationModel)
		cmds = append(cmds, creationCmd)

		if m.creationModel.Done {
			cmds = append(cmds, func() tea.Msg {
				return createTaskMsg{Task: m.creationModel.Result}
			})
		}
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	switch m.currentState {
	case ListView:
		return m.listModel.View()
	case CreationView:
		return m.creationModel.View()
	case MainView:
		return "Press l - list view\nPress c - create task\nPress q to exit\n"
	case ErrView:
		return m.err.Error()
	default:
		return "Press l - list view\nPress q to exit\n"
	}
}

func InitialMainModel(repo tasks.TaskRepository) MainModel {
	listModel := components.InitialListModel()
	creationModel := components.InitialCreationModel()

	m := MainModel{
		currentState:  MainView,
		listModel:     listModel,
		creationModel: creationModel,

		repo: repo,
	}

	return m
}
