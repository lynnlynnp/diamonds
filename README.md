# 🪩 DIAMONDS

Diamonds is a lightweight CLI utility that allows you to **quickly manage URLs and color codes** for your projects from anywhere in your terminal.

I've always found it exhausting to clutter my browser with bookmarks for URLs I need to access frequently, and for a long time I'd been using a bash script to allow me to easily copy URLs from a markdown file. 

However, after seeing my partner scrambling back and forth between apps and even physical notes to copy the color codes for her projects, I decided it was time to create a little CLI tool that could make our day-to-day work friendlier (and prettier!)

And so was Diamonds born! A little gift to my girlfriend and, hopefully, a successful hint 😉💍

> [!NOTE]
> This project is my first try at learning _Go_; improvements and new features will come, though they may indeed take a long time.

## FEATURES

- **Project Management**: Organize your colors and URLs by project.
- **Color Palette**: Store HEX color codes and visually preview them directly in the terminal.
- **Bookmark Manager**: Keep frequently accessed URLs handy.
- **Clipboard Integration**: Copy colors or URLs to your clipboard with a single keystroke.
- **Vim Keybindings**: Navigate quickly using familiar vim motions.

## INSTALLATION

### Homebrew (macOS/Linux)

```bash
brew tap lynnlynnp/diamonds
brew install diamonds
```

### From Source

```bash
git clone https://github.com/lynnlynnp/diamonds.git
cd diamonds
go build -o diamonds .
```

## USAGE

Start the application by running:

```bash
diamonds
```

> [!NOTE]
> For MacOS, you can access the TUI directly from Spotlight:
> `⌘ + SPACE` > `diamonds`

### Navigation & Controls

| Key | Action |
| :--- | :--- |
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Select project / Copy item to clipboard |
| `n` | Create new Project / Color / URL |
| `d` | Delete selected item |
| `Esc` | Go back / Cancel |
| `q` / `Ctrl+c` | Quit application |

## CONFIGURATION

Diamonds stores your data in a simple JSON file located at:

- **macOS/Linux**: `~/.config/diamonds/data.json`
- **Windows**: `%APPDATA%\diamonds\data.json`

You can manually back up or edit this file if needed.

## ACKNOWLEDGMENTS

A massive shoutout to [Charmbracelet](https://charm.sh/) whose work not only supports a huge chunk of this project but also inspired it to begin with 💖
