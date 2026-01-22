package components

import (
	"fmt"
	"strings"

	"github.com/paralerdev/paraler/internal/process"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar shows status and keybindings
type StatusBar struct {
	width  int
	styles StatusBarStyles
}

// StatusBarStyles contains status bar styles
type StatusBarStyles struct {
	Container    lipgloss.Style
	Key          lipgloss.Style
	Desc         lipgloss.Style
	Sep          lipgloss.Style
	RunningCount lipgloss.Style
	StoppedCount lipgloss.Style
	Info         lipgloss.Style
}

// DefaultStatusBarStyles returns default styles
func DefaultStatusBarStyles() StatusBarStyles {
	return StatusBarStyles{
		Container: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Padding(0, 1),
		Key: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true),
		Desc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		Sep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151")),
		RunningCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true),
		StoppedCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")),
	}
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	return &StatusBar{
		styles: DefaultStatusBarStyles(),
	}
}

// SetWidth sets the status bar width
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// View renders the status bar
func (s *StatusBar) View(manager *process.Manager, showHelp bool) string {
	if showHelp {
		return s.renderHelp()
	}

	return s.renderStatus(manager)
}

// renderStatus renders the normal status view
func (s *StatusBar) renderStatus(manager *process.Manager) string {
	running := manager.RunningCount()
	total := manager.TotalCount()

	// Left side: running status
	var statusStyle lipgloss.Style
	if running > 0 {
		statusStyle = s.styles.RunningCount
	} else {
		statusStyle = s.styles.StoppedCount
	}
	status := statusStyle.Render(fmt.Sprintf("Running: %d/%d", running, total))

	// Right side: key hints
	hints := []string{
		s.keyHint("s", "start"),
		s.keyHint("x", "stop"),
		s.keyHint("r", "restart"),
		s.keyHint("a", "add"),
		s.keyHint("d", "del"),
		s.keyHint("?", "help"),
		s.keyHint("q", "quit"),
	}
	keysHelp := strings.Join(hints, s.styles.Sep.Render(" │ "))

	// Calculate spacing
	statusWidth := lipgloss.Width(status)
	keysWidth := lipgloss.Width(keysHelp)
	padding := s.width - statusWidth - keysWidth - 4

	if padding < 1 {
		padding = 1
	}

	return s.styles.Container.
		Width(s.width).
		Render(status + strings.Repeat(" ", padding) + keysHelp)
}

// renderHelp renders the full help view
func (s *StatusBar) renderHelp() string {
	var b strings.Builder

	b.WriteString(s.styles.Info.Render("Keybindings:"))
	b.WriteString("\n\n")

	helpItems := [][]string{
		{"Navigation", "↑/k up", "↓/j down", "Tab switch panel", "pgup/pgdn scroll"},
		{"Services", "s start", "x stop", "r restart"},
		{"Bulk", "S start all", "X stop all"},
		{"Logs", "/ filter", "c clear", "g top", "G bottom"},
		{"Projects", "a add", "d delete service", "D delete project"},
		{"Other", "? help", "q quit"},
	}

	for _, group := range helpItems {
		category := group[0]
		items := group[1:]

		b.WriteString(s.styles.Key.Render(category))
		b.WriteString(s.styles.Sep.Render(": "))

		for i, item := range items {
			parts := strings.SplitN(item, " ", 2)
			if len(parts) == 2 {
				b.WriteString(s.styles.Key.Render(parts[0]))
				b.WriteString(" ")
				b.WriteString(s.styles.Desc.Render(parts[1]))
			}
			if i < len(items)-1 {
				b.WriteString(s.styles.Sep.Render(" │ "))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(s.styles.Info.Render("Press any key to close help"))

	return s.styles.Container.Render(b.String())
}

// keyHint formats a key hint
func (s *StatusBar) keyHint(key, desc string) string {
	return s.styles.Key.Render(key) + " " + s.styles.Desc.Render(desc)
}
