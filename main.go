package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// --- MAIN STATE MODEL ---

type model struct {
	projectList     list.Model
	projects        []Project
	currentView     ViewState
	cursor          int
	selectedProject int
	inputBuffer     string // Used for single-line inputs
	urlNameBuffer   string // Used for the URL name in AddUrlView
	focusedField    int    // Used in AddUrlView to track focus
	message         string
}

// --- HELPER FUNCTIONS ---

// deleteLastRune removes the last character from a string, handling unicode characters correctly.
func deleteLastRune(s string) string {
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

// --- INITIALIZATION ---

func initialModel() model {
	loadedProjects, err := loadProjects()
	if err != nil {
		fmt.Printf("Error loading projects: %v\n", err)
		os.Exit(1)
	}

	items := make([]list.Item, len(loadedProjects))
	for i, project := range loadedProjects {
		items[i] = &projectItem{project: project}
	}

	delegate := newCustomDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "🪩 DIAMONDS "
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = headerStyle.MarginTop(0).PaddingTop(1)
	l.Styles.HelpStyle = helpStyle
	l.SetShowHelp(false)

	return model{
		projectList: l,
		projects:    loadedProjects,
		currentView: ProjectListView,
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

// --- UPDATE LOOP ---

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Clear the message on any key press
	if _, ok := msg.(tea.KeyMsg); ok {
		m.message = ""
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		h, v := docStyle.GetHorizontalPadding(), docStyle.GetVerticalPadding()
		m.projectList.SetSize(msg.Width-h, msg.Height-v)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.currentView {
		case ProjectListView:
			return m.updateProjectList(msg)
		case ProjectMenuView:
			return m.updateProjectMenu(msg)
		case ColorListView:
			return m.updateColorList(msg)
		case UrlListView:
			return m.updateUrlList(msg)
		case AddProjectView:
			return m.updateAddProject(msg)
		case AddColorView:
			return m.updateAddColor(msg)
		case AddUrlView:
			return m.updateAddUrl(msg)
		case ConfirmDeleteProjectView:
			return m.updateConfirmDeleteProject(msg)
		}
	}
	return m, nil
}

// --- UPDATE LOGIC HANDLERS ---

func (m *model) updateProjectList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	if msg.String() == "ctrl+c" || msg.String() == "q" {
		return m, tea.Quit
	}

	// Trigger search
	if msg.String() == "/" && m.projectList.FilterState() == list.Unfiltered {
		m.switchToSearchItems()
	}

	// Application keys (only when not filtering)
	if m.projectList.FilterState() == list.Unfiltered {
		switch msg.String() {
		case "esc":
			// No-op here because if we are Unfiltered, we don't need to reset filter.
		case "enter":
			selectedItem := m.projectList.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}

			switch item := selectedItem.(type) {
			case *projectItem:
				for i, p := range m.projects {
					if p.Name == item.project.Name {
						m.selectedProject = i
						m.currentView = ProjectMenuView
						m.cursor = 0
						break
					}
				}
			case *colorItem:
				clipboard.WriteAll(item.color)
				m.message = fmt.Sprintf(" Copied %s to clipboard! ", item.color)
			case *urlItem:
				clipboard.WriteAll(item.url.URL)
				m.message = fmt.Sprintf(" Copied %s to clipboard! ", item.url.URL)
			}
			return m, nil
		case "n":
			m.currentView = AddProjectView
			m.inputBuffer = ""
			return m, nil
		case "d":
			selectedItem, ok := m.projectList.SelectedItem().(*projectItem)
			if ok {
				for i, p := range m.projects {
					if p.Name == selectedItem.project.Name {
						m.selectedProject = i
						m.currentView = ConfirmDeleteProjectView
						break
					}
				}
			}
			return m, nil
		}
	}

	wasFiltering := m.projectList.FilterState() == list.Filtering

	var cmd tea.Cmd
	m.projectList, cmd = m.projectList.Update(msg)

	// If we just stopped filtering (e.g. user pressed Esc), restore project items
	if wasFiltering && m.projectList.FilterState() == list.Unfiltered {
		m.updateProjectListItems()
	}

	return m, cmd
}

func (m *model) updateProjectMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.currentView = ProjectListView
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 1 {
			m.cursor++
		}
	case "enter":
		if m.cursor == 0 {
			m.currentView = ColorListView
		} else {
			m.currentView = UrlListView
		}
		m.cursor = 0
	}
	return m, nil
}

func (m *model) updateColorList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.currentView = ProjectMenuView
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.projects[m.selectedProject].Colors)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.projects[m.selectedProject].Colors) > 0 {
			color := m.projects[m.selectedProject].Colors[m.cursor]
			err := clipboard.WriteAll(color)
			if err != nil {
				m.message = fmt.Sprintf("Error copying to clipboard: %v", err)
			} else {
				m.message = fmt.Sprintf(" Copied %s to clipboard! ", color)
			}
		}
	case "d":
		if len(m.projects[m.selectedProject].Colors) > 0 {
			deletedColor := m.projects[m.selectedProject].Colors[m.cursor]
			m.projects[m.selectedProject].Colors = append(m.projects[m.selectedProject].Colors[:m.cursor], m.projects[m.selectedProject].Colors[m.cursor+1:]...)
			m.updateProjectListItems()
			m.saveProjects()
			m.message = fmt.Sprintf("Deleted color %s", deletedColor)

			if m.cursor > 0 && m.cursor >= len(m.projects[m.selectedProject].Colors) {
				m.cursor--
			}
		}
	case "n":
		m.currentView = AddColorView
		m.inputBuffer = ""
	}
	return m, nil
}

func (m *model) updateUrlList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.currentView = ProjectMenuView
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.projects[m.selectedProject].Urls)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.projects[m.selectedProject].Urls) > 0 {
			url := m.projects[m.selectedProject].Urls[m.cursor].URL
			err := clipboard.WriteAll(url)
			if err != nil {
				m.message = fmt.Sprintf("Error copying to clipboard: %v", err)
			} else {
				m.message = fmt.Sprintf(" Copied %s to clipboard! ", url)
			}
		}
	case "d":
		if len(m.projects[m.selectedProject].Urls) > 0 {
			deletedUrl := m.projects[m.selectedProject].Urls[m.cursor].Name
			m.projects[m.selectedProject].Urls = append(m.projects[m.selectedProject].Urls[:m.cursor], m.projects[m.selectedProject].Urls[m.cursor+1:]...)
			m.updateProjectListItems()
			m.saveProjects()
			m.message = fmt.Sprintf("Deleted URL '%s'", deletedUrl)

			if m.cursor > 0 && m.cursor >= len(m.projects[m.selectedProject].Urls) {
				m.cursor--
			}
		}
	case "n":
		m.currentView = AddUrlView
		m.inputBuffer = ""
		m.urlNameBuffer = ""
		m.focusedField = 0
	}
	return m, nil
}

func (m *model) updateAddProject(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.currentView = ProjectListView
		m.inputBuffer = ""
	case "enter":
		if m.inputBuffer != "" {
			m.projects = append(m.projects, Project{Name: m.inputBuffer, Colors: []string{}, Urls: []namedURL{}})
			m.updateProjectListItems()
			m.saveProjects()
			m.currentView = ProjectListView
			m.inputBuffer = ""
		}
	case "backspace":
		m.inputBuffer = deleteLastRune(m.inputBuffer)
	case " ":
		m.inputBuffer += " "
	default:
		if msg.Type == tea.KeyRunes {
			m.inputBuffer += string(msg.Runes)
		}
	}
	return m, nil
}

func (m *model) updateAddColor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.currentView = ColorListView
		m.inputBuffer = ""
	case "enter":
		if m.inputBuffer != "" && strings.HasPrefix(m.inputBuffer, "#") && (len(m.inputBuffer) == 7 || len(m.inputBuffer) == 4) {
			m.projects[m.selectedProject].Colors = append(m.projects[m.selectedProject].Colors, m.inputBuffer)
			m.updateProjectListItems()
			m.saveProjects()
			m.currentView = ColorListView
			m.cursor = len(m.projects[m.selectedProject].Colors) - 1
			m.inputBuffer = ""
		}
	case "backspace":
		m.inputBuffer = deleteLastRune(m.inputBuffer)
	default:
		if msg.Type == tea.KeyRunes && len(m.inputBuffer) < 7 {
			m.inputBuffer += string(msg.Runes)
		}
	}
	return m, nil
}

func (m *model) updateAddUrl(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.currentView = UrlListView
		m.urlNameBuffer = ""
		m.inputBuffer = ""
		m.focusedField = 0
	case "enter":
		if m.focusedField == 0 {
			m.focusedField = 1
		} else {
			if m.urlNameBuffer != "" && m.inputBuffer != "" {
				m.projects[m.selectedProject].Urls = append(m.projects[m.selectedProject].Urls, namedURL{Name: m.urlNameBuffer, URL: m.inputBuffer})
				m.updateProjectListItems()
				m.saveProjects()
				m.currentView = UrlListView
				m.cursor = len(m.projects[m.selectedProject].Urls) - 1
				m.urlNameBuffer = ""
				m.inputBuffer = ""
				m.focusedField = 0
			}
		}
	case "backspace":
		if m.focusedField == 0 {
			m.urlNameBuffer = deleteLastRune(m.urlNameBuffer)
		} else {
			m.inputBuffer = deleteLastRune(m.inputBuffer)
		}
	case "tab":
		m.focusedField = (m.focusedField + 1) % 2
	case " ":
		if m.focusedField == 0 {
			m.urlNameBuffer += " "
		} else {
			m.inputBuffer += " "
		}
	default:
		if msg.Type == tea.KeyRunes {
			if m.focusedField == 0 {
				m.urlNameBuffer += string(msg.Runes)
			} else {
				m.inputBuffer += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func (m *model) updateConfirmDeleteProject(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		if m.selectedProject >= 0 && m.selectedProject < len(m.projects) {
			deletedProjectName := m.projects[m.selectedProject].Name
			m.projects = append(m.projects[:m.selectedProject], m.projects[m.selectedProject+1:]...)
			m.updateProjectListItems()
			m.saveProjects()
			m.message = fmt.Sprintf("Deleted project '%s'", deletedProjectName)
		}
		m.currentView = ProjectListView
	case "n", "esc":
		m.currentView = ProjectListView
	}
	return m, nil
}

// --- ENTRY POINT ---

func main() {
	m := initialModel()
	p := tea.NewProgram(&m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
