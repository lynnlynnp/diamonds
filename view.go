package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// ViewState determines which view is currently active.
type ViewState int

const (
	ProjectListView ViewState = iota
	ColorListView
	AddProjectView
	AddColorView
	UrlListView
	AddUrlView
	ProjectMenuView
	ConfirmDeleteProjectView
)

// --- STYLING ---

var (
	// Palette
	appNameColor      = lipgloss.AdaptiveColor{Light: "#1E90FF", Dark: "#F6FFFE"}
	commentColor      = lipgloss.Color("#757575")
	selectionColor    = lipgloss.AdaptiveColor{Light: "#0000CD", Dark: "#BAF3EB"}
	itemDescColor     = lipgloss.AdaptiveColor{Light: "#5151D8", Dark: "#E9F8F5"}
	messageColor      = lipgloss.Color("#F1F1F1")
	messageBgColor    = lipgloss.Color("#FF5F87")
	inlineCodeColor   = lipgloss.Color("#FF5F87")
	inlineCodeBgColor = lipgloss.AdaptiveColor{Light: "#ADD8E6", Dark: "#3A3A3A"}
	quoteColor        = lipgloss.AdaptiveColor{Light: "#1E90FF", Dark: "#FF59C8"}
	normalTextColor   = lipgloss.AdaptiveColor{Light: "#1F2026", Dark: "#E5E5E5"}

	// Styles
	headerStyle = lipgloss.NewStyle().
			Foreground(appNameColor).
			Bold(true).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(commentColor)

	subtleStyle = lipgloss.NewStyle().
			Foreground(commentColor)

	messageStyle = lipgloss.NewStyle().
			Foreground(messageColor).
			Background(messageBgColor).
			Bold(true).
			Padding(0, 1)

	inlineCodeStyle = lipgloss.NewStyle().
			Foreground(inlineCodeColor).
			Background(inlineCodeBgColor).
			Padding(0, 1).
			Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(selectionColor)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(quoteColor).
			Padding(1, 2).
			Width(40)

	docStyle = lipgloss.NewStyle().Padding(2, 1).Foreground(normalTextColor)
)

func newCustomDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	c := selectionColor
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(c).BorderLeftForeground(c)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(itemDescColor).BorderLeftForeground(c)
	d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(normalTextColor)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(commentColor)
	return d
}

// --- VIEW METHODS ---

func (m *model) View() string {
	var view string
	switch m.currentView {
	case ProjectListView:
		view = m.viewProjectList()
	case ProjectMenuView:
		view = m.viewProjectMenu()
	case ColorListView:
		view = m.viewColorList()
	case UrlListView:
		view = m.viewUrlList()
	case AddProjectView:
		view = m.viewAddProject()
	case AddColorView:
		view = m.viewAddColor()
	case AddUrlView:
		view = m.viewAddUrl()
	case ConfirmDeleteProjectView:
		view = m.viewConfirmDeleteProject()
	}
	return docStyle.Render(view)
}

func (m *model) viewProjectList() string {
	var b strings.Builder
	b.WriteString(m.projectList.View())
	help := horizontalHelp("↑/↓ navigate", "n new", "d delete", "q quit")
	b.WriteString("\n" + help)

	if m.message != "" {
		b.WriteString("\n" + messageStyle.Render(m.message))
	}
	return b.String()
}

func (m *model) viewProjectMenu() string {
	project := m.projects[m.selectedProject]
	var b strings.Builder

	b.WriteString(headerStyle.Render("✨ " + project.Name) + "\n")

	options := []string{"Colors", "URLs"}
	for i, option := range options {
		if m.cursor == i {
			b.WriteString(selectedItemStyle.Render("> " + option) + "\n")
		} else {
			b.WriteString("  " + option + "\n")
		}
	}

	help := horizontalHelp("↑/↓ navigate", "enter select", "esc back", "q quit")
	b.WriteString("\n" + help)

	return b.String()
}

func (m *model) viewColorList() string {
	project := m.projects[m.selectedProject]
	var b strings.Builder

	b.WriteString(headerStyle.Render(project.Name) + "\n")

	if len(project.Colors) == 0 {
		b.WriteString(subtleStyle.Render("No colors yet. Press 'n' to add one.") + "\n")
	} else {
		for i, color := range project.Colors {
			colorBlock := lipgloss.NewStyle().Background(lipgloss.Color(color)).Render("  ")
			hexCodeStyled := inlineCodeStyle.Render(color)
			line := fmt.Sprintf("%s %s", colorBlock, hexCodeStyled)

			if m.cursor == i {
				cursorStyle := lipgloss.NewStyle().Foreground(selectionColor)
				styledCursor := cursorStyle.Render("> ")
				styledLine := selectedItemStyle.Render(line)
				b.WriteString(styledCursor + styledLine + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
	}

	help := horizontalHelp("↑/↓ navigate", "enter copy", "n new", "d delete", "esc back", "q quit")
	b.WriteString("\n" + help)

	if m.message != "" {
		b.WriteString("\n" + messageStyle.Render(m.message))
	}

	return b.String()
}

func (m *model) viewUrlList() string {
	project := m.projects[m.selectedProject]
	var b strings.Builder

	b.WriteString(headerStyle.Render(project.Name) + "\n")

	if len(project.Urls) == 0 {
		b.WriteString(subtleStyle.Render("No URLs yet. Press 'n' to add one.") + "\n")
	} else {
		for i, namedUrl := range project.Urls {
			if m.cursor == i {
				b.WriteString(selectedItemStyle.Render("> " + namedUrl.Name) + "\n")
			} else {
				b.WriteString("  " + namedUrl.Name + "\n")
			}
		}
	}

	help := horizontalHelp("↑/↓ navigate", "enter copy", "n new", "d delete", "esc back", "q quit")
	b.WriteString("\n" + help)

	if m.message != "" {
		b.WriteString("\n" + messageStyle.Render(m.message))
	}

	return b.String()
}

func (m *model) viewAddProject() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("Add New Project") + "\n")
	prompt := fmt.Sprintf("Project name: %s", m.inputBuffer)
	b.WriteString(inputStyle.Render(prompt) + "\n\n")
	b.WriteString(horizontalHelp("enter save", "esc cancel"))
	return b.String()
}

func (m *model) viewAddColor() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("Add New Color") + "\n")
	prompt := fmt.Sprintf("HEX color: %s", m.inputBuffer)
	b.WriteString(inputStyle.Render(prompt) + "\n\n")
	b.WriteString(helpStyle.Render("Enter HEX (e.g., #FF5F87)") + "\n")
	b.WriteString(horizontalHelp("enter save", "esc cancel"))
	return b.String()
}

func (m *model) viewAddUrl() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("Add New URL") + "\n")

	namePrompt := fmt.Sprintf("Name: %s", m.urlNameBuffer)
	urlPrompt := fmt.Sprintf("URL: %s", m.inputBuffer)

	if m.focusedField == 0 {
		b.WriteString(inputStyle.Render(namePrompt) + "\n")
		b.WriteString(subtleStyle.Render(urlPrompt) + "\n\n")
	} else {
		b.WriteString(subtleStyle.Render(namePrompt) + "\n")
		b.WriteString(inputStyle.Render(urlPrompt) + "\n\n")
	}

	b.WriteString(horizontalHelp("enter next/save", "tab switch fields", "esc cancel"))
	return b.String()
}

func (m *model) viewConfirmDeleteProject() string {
	projectName := ""
	if m.selectedProject >= 0 && m.selectedProject < len(m.projects) {
		projectName = m.projects[m.selectedProject].Name
	}
	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf("Delete '%s'?", projectName)) + "\n\n")
	b.WriteString("Are you sure? This action cannot be undone.\n\n")
	b.WriteString(horizontalHelp("y yes", "n no", "esc cancel"))
	return b.String()
}

func horizontalHelp(keys ...string) string {
	return helpStyle.Render(strings.Join(keys, " • "))
}
