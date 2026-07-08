package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Enter       key.Binding
	SSH         key.Binding
	Search      key.Binding
	ClearSearch key.Binding
	SwitchSub   key.Binding
	SetJumpHost key.Binding
	Refresh     key.Binding
	Help        key.Binding
	Quit        key.Binding
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "ssh"),
	),
	SSH: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "ssh"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	ClearSearch: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear search/close"),
	),
	SwitchSub: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch subscription"),
	),
	SetJumpHost: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "set jump host"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh VMs"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
