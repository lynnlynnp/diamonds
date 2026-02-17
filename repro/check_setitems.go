package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	cmd := l.SetItems([]list.Item{})
    _ = tea.Cmd(cmd)
	if cmd == nil {
		fmt.Println("SetItems returns nil cmd (or void if this doesn't compile)")
	} else {
		fmt.Println("SetItems returns a cmd")
	}
}
