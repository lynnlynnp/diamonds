package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

const dataFileName = "data.json"
const configDirName = "diamonds"

// --- DATA STRUCTURES ---

type namedURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Project struct {
	Name   string     `json:"name"`
	Colors []string   `json:"colors"`
	Urls   []namedURL `json:"urls"`
}

// --- LIST ADAPTER (Project) ---

// projectItem adapts Project to the list.Item interface required by bubbles/list
type projectItem struct {
	project Project
}

func (p *projectItem) FilterValue() string {
	var b strings.Builder
	b.WriteString(p.project.Name)
	for _, c := range p.project.Colors {
		b.WriteString(" " + c)
	}
	for _, u := range p.project.Urls {
		b.WriteString(" " + u.Name)
	}
	return b.String()
}

func (p *projectItem) Title() string { return p.project.Name }
func (p *projectItem) Description() string {
	colorCount := len(p.project.Colors)
	urlCount := len(p.project.Urls)
	colorStr := "colors"
	if colorCount == 1 {
		colorStr = "color"
	}
	urlStr := "URLs"
	if urlCount == 1 {
		urlStr = "URL"
	}
	return fmt.Sprintf("%d %s, %d %s", colorCount, colorStr, urlCount, urlStr)
}

// --- LIST ADAPTER (Color & URL) ---

type colorItem struct {
	color   string
	project string
}

func (c *colorItem) FilterValue() string { return c.color + " " + c.project }
func (c *colorItem) Title() string       { return c.color }
func (c *colorItem) Description() string { return fmt.Sprintf("Color in %s", c.project) }

type urlItem struct {
	url     namedURL
	project string
}

func (u *urlItem) FilterValue() string { return u.url.Name + " " + u.url.URL + " " + u.project }
func (u *urlItem) Title() string       { return u.url.Name }
func (u *urlItem) Description() string { return fmt.Sprintf("%s • %s", u.url.URL, u.project) }

// --- FILE I/O ---

func getDataFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get user config dir: %w", err)
	}

	appConfigDir := filepath.Join(configDir, configDirName)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("could not create app config dir: %w", err)
	}

	return filepath.Join(appConfigDir, dataFileName), nil
}

func loadProjects() ([]Project, error) {
	path, err := getDataFilePath()
	if err != nil {
		return nil, fmt.Errorf("could not get data file path: %w", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Project{}, nil // No file, start fresh
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read data file: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("could not parse data file: %w", err)
	}

	return projects, nil
}

// --- MODEL METHODS (Data) ---

func (m *model) saveProjects() {
	path, err := getDataFilePath()
	if err != nil {
		m.message = fmt.Sprintf("Error getting data path: %v", err)
		return
	}

	data, err := json.MarshalIndent(m.projects, "", "  ")
	if err != nil {
		m.message = fmt.Sprintf("Error saving data: %v", err)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		m.message = fmt.Sprintf("Error writing data: %v", err)
	}
}

func (m *model) updateProjectListItems() {
	items := make([]list.Item, len(m.projects))
	for i, project := range m.projects {
		items[i] = &projectItem{project: project}
	}
	m.projectList.SetItems(items)
}

func (m *model) switchToSearchItems() {
	var items []list.Item
	for _, p := range m.projects {
		for _, c := range p.Colors {
			items = append(items, &colorItem{color: c, project: p.Name})
		}
		for _, u := range p.Urls {
			items = append(items, &urlItem{url: u, project: p.Name})
		}
	}
	m.projectList.SetItems(items)
}
