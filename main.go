package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/atotto/clipboard"
	"github.com/pkg/browser"
)

const dataFile = "items.json"

type styles struct {
	app           lipgloss.Style
	title         lipgloss.Style
	statusMessage lipgloss.Style
	focusedPrompt lipgloss.Style
	blurredPrompt lipgloss.Style
}

func newStyles(darkBG bool) styles {
	lightDark := lipgloss.LightDark(darkBG)

	return styles{
		app: lipgloss.NewStyle().
			Padding(1, 2),
		title: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#1E90FF"), lipgloss.Color("#F6FFFE"))).
			Bold(true).
			Padding(0, 0),
		statusMessage: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#04B575"), lipgloss.Color("#04B575"))),
		focusedPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")),
		blurredPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

// --- Data Structure ---

type item struct {
	NameValue   string `json:"name"`
	TargetValue string `json:"target"`
	ItemType    string `json:"type"`
}

func (i item) Title() string       { return i.NameValue }
func (i item) Description() string { return fmt.Sprintf("[%s] %s", i.ItemType, i.TargetValue) }
func (i item) FilterValue() string { return i.NameValue }

// --- Persistence Helpers ---

func loadList() []list.Item {
	b, err := os.ReadFile(dataFile)
	if err != nil {
		return []list.Item{}
	}

	var savedItems []item
	if err := json.Unmarshal(b, &savedItems); err != nil {
		return []list.Item{}
	}

	items := make([]list.Item, len(savedItems))
	for i, itm := range savedItems {
		items[i] = itm
	}
	return items
}

func saveList(items []list.Item) {
	var toSave []item
	for _, i := range items {
		if itm, ok := i.(item); ok {
			toSave = append(toSave, itm)
		}
	}

	b, err := json.MarshalIndent(toSave, "", "  ")
	if err == nil {
		_ = os.WriteFile(dataFile, b, 0644)
	}
}

// --- Key Bindings ---

type listKeyMap struct {
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	insertItem       key.Binding
	copyValue        key.Binding // Added: Copy Binding
	openUrl          key.Binding // Added: Open Browser Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		insertItem:       key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add item")),
		togglePagination: key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "toggle pagination")),
		toggleHelpMenu:   key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "toggle help")),
		copyValue:        key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy value")),
		openUrl:          key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "open url")),
	}
}

// --- Main Model ---

type model struct {
	styles        styles
	darkBG        bool
	width, height int
	once          *sync.Once
	list          list.Model
	keys          *listKeyMap

	// Form State
	inputs     []textinput.Model
	focusIndex int
	addingItem bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		textinput.Blink,
	)
}

func (m *model) updateListProperties() {
	h, v := m.styles.app.GetFrameSize()
	m.list.SetSize(m.width-h, m.height-v)
	m.styles = newStyles(m.darkBG)
	m.list.Styles.Title = m.styles.title
}

func (m *model) resetForm() {
	m.addingItem = false
	m.focusIndex = 0
	for i := range m.inputs {
		m.inputs[i].SetValue("")
		if i == 0 {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.darkBG = msg.IsDark()
		m.updateListProperties()
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.updateListProperties()
		return m, nil
	}

	// --- 1. Form Mode (Adding a new item) ---
	if m.addingItem {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.resetForm()
				return m, nil

			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()

				// If Enter is pressed on the LAST input field, submit the form!
				if s == "enter" && m.focusIndex == len(m.inputs)-1 {
					name := strings.TrimSpace(m.inputs[0].Value())
					val := strings.TrimSpace(m.inputs[1].Value())

					if name != "" && val != "" {
						// Auto-detect type based on the value
						itemType := "TEXT"
						lowerVal := strings.ToLower(val)
						if strings.HasPrefix(lowerVal, "http://") || strings.HasPrefix(lowerVal, "https://") {
							itemType = "URL"
						} else if strings.HasPrefix(val, "#") {
							itemType = "HEX"
						}

						newItem := item{
							NameValue:   name,
							TargetValue: val,
							ItemType:    itemType,
						}

						insCmd := m.list.InsertItem(0, newItem)
						statusCmd := m.list.NewStatusMessage(m.styles.statusMessage.Render("Added " + newItem.Title()))
						cmds = append(cmds, insCmd, statusCmd)

						saveList(m.list.Items())
					}
					
					m.resetForm()
					return m, tea.Batch(cmds...)
				}

				// Cycle focus logic
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.inputs)-1 {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs) - 1
				}

				for i := 0; i <= len(m.inputs)-1; i++ {
					if i == m.focusIndex {
						cmds = append(cmds, m.inputs[i].Focus())
						continue
					}
					m.inputs[i].Blur()
				}
				return m, tea.Batch(cmds...)
			}
		}

		// Safely append commands to the slice instead of indexing directly
		var cmd tea.Cmd
		for i := range m.inputs {
			m.inputs[i], cmd = m.inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// --- 2. List Mode ---
	preLen := len(m.list.Items())

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil

		case key.Matches(msg, m.keys.insertItem):
			m.addingItem = true
			return m, textinput.Blink

		// --- NEW: Copy Value ---
		case key.Matches(msg, m.keys.copyValue):
		    if i, ok := m.list.SelectedItem().(item); ok {
			// We copy if it's URL or HEX as requested
			if i.ItemType == "URL" || i.ItemType == "HEX" {
			    _ = clipboard.WriteAll(i.TargetValue)
			    cmd := m.list.NewStatusMessage(m.styles.statusMessage.Render(fmt.Sprintf("Copied %s %s to clipboard", i.NameValue, i.ItemType)))
			    return m, cmd
			}
		    }

	 	// --- NEW: Open URL ---
		case key.Matches(msg, m.keys.openUrl):
			if i, ok := m.list.SelectedItem().(item); ok {
				// Only URLs trigger the browser
				if i.ItemType == "URL" {
					_ = browser.OpenURL(i.TargetValue)
					cmd := m.list.NewStatusMessage(m.styles.statusMessage.Render("Opening in browser: " + i.TargetValue))
					return m, cmd
				}
			}
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	if len(m.list.Items()) != preLen {
		saveList(m.list.Items())
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	var content string

	if m.addingItem {
		var b strings.Builder
		b.WriteString(m.styles.title.Render(" Add New Item ") + "\n\n")

		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		b.WriteString("\n\n(Tab to switch fields, Enter to save, Esc to cancel)")
		content = b.String()
	} else {
		content = m.list.View()
	}

	v := tea.NewView(m.styles.app.Render(content))
	v.AltScreen = true
	return v
}

func initialModel() model {
	m := model{}
	m.styles = newStyles(false)

	listKeys := newListKeyMap()

	// Initialize the Multi-Input Form
	m.inputs = make([]textinput.Model, 2)
	for i := range m.inputs {
		t := textinput.New()
		t.CharLimit = 156
		t.SetWidth(40)

		switch i {
		case 0:
			t.Placeholder = "Item Name (e.g. Dashboard)"
			t.Focus()
		case 1:
			t.Placeholder = "Value (e.g. https://... or #FF0000)"
		}
		m.inputs[i] = t
	}

	items := loadList()
	delegate := list.NewDefaultDelegate()
	
	mainList := list.New(items, delegate, 0, 0)
	mainList.Title = "🪩 DIAMONDS"
	mainList.SetShowStatusBar(false)
	mainList.Styles.Title = m.styles.title
	
	// Ensure the new keys show up in the help menu
	mainList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.insertItem,
			listKeys.copyValue,
			listKeys.openUrl,
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
		}
	}

	m.list = mainList
	m.keys = listKeys

	return m
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
