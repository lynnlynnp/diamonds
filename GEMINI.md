# Diamonds

**Diamonds** is a lightweight CLI utility for managing project-specific URLs and color codes. It features a TUI (Terminal User Interface) built with Go and the Charm ecosystem (Bubble Tea, Lip Gloss).

## Project Overview

- **Purpose:** Organize colors and URLs by project, preview colors in the terminal, and quickly copy data to the clipboard.
- **Key Features:** Vim-style navigation, JSON persistence, and clipboard integration.
- **Tech Stack:**
  - **Language:** Go
  - **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture)
  - **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
  - **Components:** [Bubbles](https://github.com/charmbracelet/bubbles) (List component)

## Building and Running

### Prerequisites
- Go 1.24+

### Commands
- **Run:** `go run .`
- **Build:** `go build -o diamonds .`
- **Install (Local):** `go install .`

### Reproduction Scripts
The `repro/` directory contains standalone Go scripts for reproducing specific issues or testing isolated behavior (e.g., `repro/check_setitems.go`).

## Architecture & Codebase

The application follows the **Elm Architecture** (Model-Update-View) standard in Bubble Tea applications.

### Key Files
- **`main.go`**: Contains the application entry point, the main `model` definition, and the `Update` loop handling all key events and state transitions.
- **`model.go`**: Defines data structures (`Project`, `namedURL`), list adapters (`FilterValue`, `Title`), and persistence logic (`loadProjects`, `saveProjects`).
- **`view.go`**: Handles all rendering logic. It contains the `View()` method which delegates to specific view functions (e.g., `viewProjectList`, `viewColorList`) based on the `currentView` state. It also defines all `lipgloss` styles.
- **`data.json`**: Data is stored in `~/.config/diamonds/data.json` (macOS/Linux) or `%APPDATA%\diamonds\data.json` (Windows).

### State Management
The `model` struct in `main.go` holds the entire application state:
- **`currentView`**: Enum tracking the active screen (ProjectList, ColorList, AddProject, etc.).
- **`projects`**: Slice of `Project` structs loaded from JSON.
- **`projectList`**: A `bubbles/list` component for the main project view.
- **`cursor` / `selectedProject`**: Integers tracking navigation state.

## Development Conventions

### Styling
- All styles are defined in `view.go` using `lipgloss`.
- Use `lipgloss.AdaptiveColor` to support both light and dark terminal themes.

### Navigation
- **Vim Bindings:** Ensure `j`/`k` are always available for vertical navigation alongside arrow keys.
- **Escape:** Used consistently to go back or cancel an action.
- **Enter:** Used to select or confirm.

### Persistence
- Data is saved to disk immediately after any modification (add/delete) via `saveProjects()`.
- Error handling for I/O should populate `m.message` to be displayed in the status area.
