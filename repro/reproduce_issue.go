package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
        if msg.String() == "/" && m.list.FilterState() == list.Unfiltered {
            items := []list.Item{
                item{title: "Vital Few Slides", desc: "Description 1"},
                item{title: "Innovation Sync Up Slides", desc: "Description 2"},
            }
            // Ignore the command returned by SetItems
            m.list.SetItems(items)
        }
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.list.View()
}

func main() {
	items := []list.Item{
		item{title: "Item A", desc: "Desc A"},
	}
    
    // Note: Assuming Update returns (Model, Cmd) in the struct definition (which is true for bubbles/list).
    
    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.SetFilteringEnabled(true)
    l.Title = "Reproduction"
    
	m := model{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
