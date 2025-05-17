package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zeerodex/goot/internal/tasks"
)

const listHeight = 14

var (
	// TODO: move main color to cfg
	titleStyle        = list.DefaultStyles().Title.Background(lipgloss.Color("12"))
	titleBarStyle     = list.DefaultStyles().TitleBar.Padding(0, 0, 0, 3)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(3)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("12"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	quitTextStyle     = lipgloss.NewStyle().Margin(0)
)

type item struct {
	id    string
	title string
}

func (i item) Title() string       { return i.title }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(i.title))
}

type ChoiceListModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m ChoiceListModel) Init() tea.Cmd {
	return nil
}

func (m ChoiceListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.id
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ChoiceListModel) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("Not hungry? That’s cool.")
	}
	return m.list.View()
}

func ChooseList(tasks tasks.Tasks) string {
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		item := item{id: task.ID, title: task.FullTitle()}
		items[i] = item
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Choose task:"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.Styles.Title = titleStyle
	l.Styles.TitleBar = titleBarStyle
	l.Styles.PaginationStyle = paginationStyle

	m := ChoiceListModel{list: l}

	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if final, ok := finalModel.(ChoiceListModel); ok {
		if final.quitting {
			return ""
		}
		return final.choice
	}
	return ""
}
