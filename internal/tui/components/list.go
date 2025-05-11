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
	id          int
	title, desc string
	completed   bool
}

func (i item) ID() int             { return i.id }
func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) Completed() bool     { return i.completed }
func (i item) FilterValue() string { return i.title }

type listKeyMap struct {
	deleteTask     key.Binding
	createTask     key.Binding
	toogleComplete key.Binding
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
		toogleComplete: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toogle task completion"),
		),
	}
}

type ListModel struct {
	list list.Model
	keys *listKeyMap

	Method   string
	Selected item
}

func (m *ListModel) SetTasks(tasks tasks.Tasks) tea.Cmd {
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		title := task.Title
		if task.Completed {
			title += " | Completed"
			title += " | " + task.DueStr()
		} else {
			title += " | Uncompleted"
			title += " | " + task.DueStr()
		}
		items[i] = item{id: task.ID, title: title, desc: task.Description, completed: task.Completed}
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
			selected := m.list.SelectedItem().(item)
			statusCmd := m.list.NewStatusMessage("Deleted " + selected.Title())
			m.Selected = selected
			return m, statusCmd
		case key.Matches(msg, m.keys.createTask):
			m.Method = "create"
			return m, nil
		case key.Matches(msg, m.keys.toogleComplete):
			m.Method = "toogle"
			selected := m.list.SelectedItem().(item)
			statusCmd := m.list.NewStatusMessage("Toogle completed for" + selected.Title())
			m.Selected = selected
			return m, statusCmd
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
			listKeys.toogleComplete,
		}
	}
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.createTask,
			listKeys.deleteTask,
			listKeys.toogleComplete,
		}
	}

	m = ListModel{list: list, keys: listKeys}

	return m
}
