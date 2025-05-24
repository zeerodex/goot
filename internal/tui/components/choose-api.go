package components

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ChooseAPIsModel struct {
	cursor  int
	options []string
	choices map[string]bool

	quitting bool
}

func (m ChooseAPIsModel) Init() tea.Cmd {
	return nil
}

func (m ChooseAPIsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.cursor == len(m.options) {
				return m, tea.Quit
			}
			selected := m.options[m.cursor]
			m.choices[selected] = !m.choices[selected]
			return m, nil

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.options)+1 {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.options)
			}
		}
	}

	return m, nil
}

func (m ChooseAPIsModel) View() string {
	s := strings.Builder{}
	for i, option := range m.options {
		if m.cursor == i {
			if m.choices[option] {
				s.WriteString(focusedStyle.Render("[x] "))
			} else {
				s.WriteString(focusedStyle.Render("[ ] "))
			}
			s.WriteString(focusedStyle.Render(m.options[i]))
		} else if m.choices[option] {
			s.WriteString("[x] ")
			s.WriteString(m.options[i])
		} else {
			s.WriteString("[ ] ")
			s.WriteString(blurredStyle.Render(m.options[i]))
		}
		s.WriteString("\n")
	}
	button := &blurredButton
	if m.cursor == len(m.options) {
		button = &focusedButton
	}
	s.WriteString("\n")
	s.WriteString(*button)

	return s.String()
}

func ChooseAPI(choices map[string]bool) (map[string]bool, bool) {
	var options []string
	for option := range choices {
		options = append(options, option)
	}

	model := ChooseAPIsModel{
		cursor:  0,
		options: options,
		choices: choices,
	}

	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	if final, ok := finalModel.(ChooseAPIsModel); ok {
		if final.quitting {
			return nil, false
		}
		return final.choices, true
	}
	return nil, false
}
