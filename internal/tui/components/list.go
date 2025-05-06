package components

import (
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

type ListModel struct {
	list list.Model

	Err error
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := msg.Width, msg.Height
		m.list.SetSize(h, v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	return m.list.View()
}

func InitialListModel(tasks tasks.Tasks) ListModel {
	var m ListModel

	var items []list.Item

	for _, task := range tasks {
		item := item{title: task.Task, desc: task.Description}
		items = append(items, item)
	}

	m = ListModel{list: list.New(items, list.NewDefaultDelegate(), 75, 30)}
	m.list.Title = "Tasks"

	return m
}
