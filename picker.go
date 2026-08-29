package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	currentMarkStyle  = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("34"))
)

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct {
	current string
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := string(i)
	isCurrent := str == d.current

	if index == m.Index() {
		prefix := "> "
		if isCurrent {
			prefix = "* "
		}
		fmt.Fprint(w, selectedItemStyle.Render(prefix+str))
	} else if isCurrent {
		fmt.Fprint(w, currentMarkStyle.Render("* "+str))
	} else {
		fmt.Fprint(w, itemStyle.Render(str))
	}
}

type model struct {
	list     list.Model
	selected string
	quitting bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.selected = string(i)
			}
			m.quitting = true
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

func termWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 80
}

func pickFromList(title string, options []string) (string, error) {
	items := make([]list.Item, len(options))
	for i, o := range options {
		items[i] = item(o)
	}

	delegate := itemDelegate{}
	l := list.New(items, delegate, termWidth(), min(len(options)+8, 20))
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	p := tea.NewProgram(model{list: l})
	result, err := p.Run()
	if err != nil {
		return "", err
	}

	return result.(model).selected, nil
}

func pickModel(models []string, current string) (string, error) {
	items := make([]list.Item, len(models))
	for i, m := range models {
		items[i] = item(m)
	}

	delegate := itemDelegate{current: current}
	l := list.New(items, delegate, termWidth(), min(len(models)+8, 20))
	l.Title = "Select Claude model"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	p := tea.NewProgram(model{list: l})
	result, err := p.Run()
	if err != nil {
		return "", err
	}

	return result.(model).selected, nil
}
