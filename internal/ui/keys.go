package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Tab        key.Binding
	Start      key.Binding
	Stop       key.Binding
	Restart    key.Binding
	StartAll   key.Binding
	StopAll    key.Binding
	Filter     key.Binding
	ClearLogs  key.Binding
	Help       key.Binding
	Quit       key.Binding
	Enter      key.Binding
	Escape     key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Home       key.Binding
	End        key.Binding
	AddProject    key.Binding
	DeleteService key.Binding
	DeleteProject key.Binding
	Space         key.Binding
	Confirm       key.Binding
	ReloadConfig    key.Binding
	ExportLogs      key.Binding
	ToggleSelect    key.Binding
	ClearSelect     key.Binding
	MoveService     key.Binding
	Rename          key.Binding
	CopyMode        key.Binding
	CopyModeSelect  key.Binding
	CopyModeCopy    key.Binding
	Fullscreen      key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start"),
		),
		Stop: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "stop"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		StartAll: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "start all"),
		),
		StopAll: key.NewBinding(
			key.WithKeys("X"),
			key.WithHelp("X", "stop all"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearLogs: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear logs"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdown", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end", "bottom"),
		),
		AddProject: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add project"),
		),
		DeleteService: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete service"),
		),
		DeleteProject: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete project"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		ReloadConfig: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "reload config"),
		),
		ExportLogs: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export logs"),
		),
		ToggleSelect: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "toggle select"),
		),
		ClearSelect: key.NewBinding(
			key.WithKeys("V"),
			key.WithHelp("V", "clear selection"),
		),
		MoveService: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "move service"),
		),
		Rename: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "rename"),
		),
		CopyMode: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy mode"),
		),
		CopyModeSelect: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "select"),
		),
		CopyModeCopy: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y", "copy"),
		),
		Fullscreen: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "fullscreen"),
		),
	}
}

// ShortHelp returns a short help string
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Start, k.Stop, k.Restart, k.Filter, k.Help, k.Quit}
}

// FullHelp returns the full help
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab},
		{k.Start, k.Stop, k.Restart},
		{k.StartAll, k.StopAll},
		{k.Filter, k.ClearLogs},
		{k.DeleteService, k.DeleteProject},
		{k.MoveService, k.Rename, k.ReloadConfig},
		{k.Help, k.Quit},
	}
}
