package components

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
)

type ListErrorMsg struct {
	Err error
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type listKeyMap struct {
	deleteTask key.Binding
	createTask key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		createTask: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "create task"),
		),
		deleteTask: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete task"),
		),
	}
}

type ListModel struct {
	list list.Model
	keys *listKeyMap

	Method   string
	Selected string
}

func (m *ListModel) SetTasks(tasks tasks.Tasks) tea.Cmd {
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = item{title: task.Task, desc: task.Description}
	}
	return m.list.SetItems(items)
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := msg.Width, msg.Height
		m.list.SetSize(h, v)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.deleteTask):
			m.Method = "delete"
			m.Selected = m.list.SelectedItem().(item).Title()
			statusCmd := m.list.NewStatusMessage("Deleted " + m.Selected)
			return m, statusCmd
		case key.Matches(msg, m.keys.createTask):
			m.Method = "create"
			return m, nil
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ListModel) View() string {
	return m.list.View()
}

func InitialListModel() ListModel {
	var m ListModel

	var items []list.Item

	listKeys := newListKeyMap()

	list := list.New(items, list.NewDefaultDelegate(), 75, 30)

	// HACK:
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.createTask,
			listKeys.deleteTask,
		}
	}
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.createTask,
			listKeys.deleteTask,
		}
	}

	m = ListModel{list: list, keys: listKeys}

	return m
}
