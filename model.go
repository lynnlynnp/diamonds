package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	name       string
	colorCount int
	urlCount   int
}

func (p projectItem) FilterValue() string { return p.name }
func (p projectItem) Title() string       { return p.name }
func (p projectItem) Description() string {
	colorStr := "colors"
	if p.colorCount == 1 {
		colorStr = "color"
	}
	urlStr := "URLs"
	if p.urlCount == 1 {
		urlStr = "URL"
	}
	return fmt.Sprintf("%d %s, %d %s", p.colorCount, colorStr, p.urlCount, urlStr)
}

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
		items[i] = projectItem{name: project.Name, colorCount: len(project.Colors), urlCount: len(project.Urls)}
	}
	m.projectList.SetItems(items)
}
