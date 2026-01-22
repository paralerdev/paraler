package components

import (
	"fmt"
	"strings"

	"github.com/paralerdev/paraler/internal/config"
	"github.com/paralerdev/paraler/internal/log"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// LogPanel displays logs for a selected service
type LogPanel struct {
	viewport      viewport.Model
	filterInput   textinput.Model
	serviceID     config.ServiceID
	serviceConfig *config.Service
	filter        string
	filtering     bool
	autoScroll    bool
	width         int
	height        int
	focused       bool
	styles        LogPanelStyles
}

// LogPanelStyles contains log panel styles
type LogPanelStyles struct {
	Container      lipgloss.Style
	Title          lipgloss.Style
	TitleFocused   lipgloss.Style
	Line           lipgloss.Style
	LineStderr     lipgloss.Style
	Timestamp      lipgloss.Style
	FilterPrompt   lipgloss.Style
	FilterInput    lipgloss.Style
	NoLogs         lipgloss.Style
	ServiceColor   lipgloss.Style
	Footer         lipgloss.Style
	FooterLabel    lipgloss.Style
	FooterValue    lipgloss.Style
}

// DefaultLogPanelStyles returns default styles
func DefaultLogPanelStyles() LogPanelStyles {
	return LogPanelStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1),
		TitleFocused: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8B5CF6")).
			Padding(0, 1),
		Line: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")),
		LineStderr: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")),
		Timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		FilterPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true),
		FilterInput: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")),
		NoLogs: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true),
		ServiceColor: lipgloss.NewStyle().
			Bold(true),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1),
		FooterLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")),
		FooterValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")),
	}
}

// NewLogPanel creates a new log panel
func NewLogPanel() *LogPanel {
	ti := textinput.New()
	ti.Placeholder = "Filter logs..."
	ti.CharLimit = 100

	vp := viewport.New(0, 0)

	return &LogPanel{
		viewport:    vp,
		filterInput: ti,
		autoScroll:  true,
		styles:      DefaultLogPanelStyles(),
	}
}

// SetSize sets the panel dimensions
func (l *LogPanel) SetSize(width, height int) {
	l.width = width
	l.height = height

	// Adjust viewport size for borders and title
	vpHeight := height - 4
	if l.filtering {
		vpHeight -= 1
	}
	if vpHeight < 1 {
		vpHeight = 1
	}

	l.viewport.Width = width - 4
	l.viewport.Height = vpHeight
}

// SetFocused sets the focus state
func (l *LogPanel) SetFocused(focused bool) {
	l.focused = focused
}

// SetService sets the current service to display
func (l *LogPanel) SetService(id config.ServiceID) {
	if l.serviceID != id {
		l.serviceID = id
		l.autoScroll = true
	}
}

// SetServiceConfig sets the current service configuration for footer display
func (l *LogPanel) SetServiceConfig(cfg *config.Service) {
	l.serviceConfig = cfg
}

// StartFilter starts filtering mode
func (l *LogPanel) StartFilter() {
	l.filtering = true
	l.filterInput.Focus()
	l.SetSize(l.width, l.height) // Recalculate sizes
}

// StopFilter stops filtering mode
func (l *LogPanel) StopFilter() {
	l.filtering = false
	l.filterInput.Blur()
	l.SetSize(l.width, l.height)
}

// ApplyFilter applies the current filter
func (l *LogPanel) ApplyFilter() {
	l.filter = l.filterInput.Value()
	l.StopFilter()
}

// ClearFilter clears the filter
func (l *LogPanel) ClearFilter() {
	l.filter = ""
	l.filterInput.SetValue("")
	l.StopFilter()
}

// IsFiltering returns true if in filter mode
func (l *LogPanel) IsFiltering() bool {
	return l.filtering
}

// Filter returns the current filter string
func (l *LogPanel) Filter() string {
	return l.filter
}

// FilterInput returns the filter input model
func (l *LogPanel) FilterInput() *textinput.Model {
	return &l.filterInput
}

// Viewport returns the viewport model
func (l *LogPanel) Viewport() *viewport.Model {
	return &l.viewport
}

// Update updates the log panel with new entries
func (l *LogPanel) Update(buffer *log.Buffer) {
	entries := buffer.GetFiltered(l.serviceID, l.filter)

	var lines []string
	for _, entry := range entries {
		timestamp := l.styles.Timestamp.Render(entry.Timestamp.Format("15:04:05"))
		var line string
		if entry.IsStderr {
			line = l.styles.LineStderr.Render(entry.Line)
		} else {
			line = l.styles.Line.Render(entry.Line)
		}
		lines = append(lines, fmt.Sprintf("%s %s", timestamp, line))
	}

	content := strings.Join(lines, "\n")
	l.viewport.SetContent(content)

	if l.autoScroll {
		l.viewport.GotoBottom()
	}
}

// ScrollUp scrolls up
func (l *LogPanel) ScrollUp() {
	l.autoScroll = false
	l.viewport.LineUp(1)
}

// ScrollDown scrolls down
func (l *LogPanel) ScrollDown() {
	l.viewport.LineDown(1)
	if l.viewport.AtBottom() {
		l.autoScroll = true
	}
}

// PageUp scrolls up a page
func (l *LogPanel) PageUp() {
	l.autoScroll = false
	l.viewport.HalfViewUp()
}

// PageDown scrolls down a page
func (l *LogPanel) PageDown() {
	l.viewport.HalfViewDown()
	if l.viewport.AtBottom() {
		l.autoScroll = true
	}
}

// GoToTop scrolls to top
func (l *LogPanel) GoToTop() {
	l.autoScroll = false
	l.viewport.GotoTop()
}

// GoToBottom scrolls to bottom
func (l *LogPanel) GoToBottom() {
	l.autoScroll = true
	l.viewport.GotoBottom()
}

// View renders the log panel
func (l *LogPanel) View(buffer *log.Buffer) string {
	var b strings.Builder

	// Title
	title := "Logs"
	if l.serviceID.Service != "" {
		title = fmt.Sprintf("Logs: %s/%s", l.serviceID.Project, l.serviceID.Service)
	}
	if l.filter != "" {
		title += fmt.Sprintf(" (filter: %s)", l.filter)
	}

	if l.focused {
		b.WriteString(l.styles.TitleFocused.Render(title))
	} else {
		b.WriteString(l.styles.Title.Render(title))
	}
	b.WriteString("\n")

	// Update content
	l.Update(buffer)

	// Viewport
	if l.viewport.TotalLineCount() == 0 {
		noLogsMsg := "No logs yet. Start a service to see output."
		if l.filter != "" {
			noLogsMsg = "No logs match the filter."
		}
		b.WriteString(l.styles.NoLogs.Render(noLogsMsg))
	} else {
		b.WriteString(l.viewport.View())
	}

	// Filter input
	if l.filtering {
		b.WriteString("\n")
		b.WriteString(l.styles.FilterPrompt.Render("/"))
		b.WriteString(l.filterInput.View())
	}

	// Footer with env/port info
	if l.serviceConfig != nil && !l.filtering {
		footer := l.renderFooter()
		if footer != "" {
			b.WriteString("\n")
			b.WriteString(l.styles.Footer.Render(footer))
		}
	}

	// Container style
	if l.focused {
		l.styles.Container = l.styles.Container.BorderForeground(lipgloss.Color("#8B5CF6"))
	} else {
		l.styles.Container = l.styles.Container.BorderForeground(lipgloss.Color("#374151"))
	}

	return l.styles.Container.
		Width(l.width).
		Height(l.height).
		Render(b.String())
}

// ServiceID returns the current service ID
func (l *LogPanel) ServiceID() config.ServiceID {
	return l.serviceID
}

// renderFooter renders the footer with service info
func (l *LogPanel) renderFooter() string {
	if l.serviceConfig == nil {
		return ""
	}

	var parts []string

	// Port info
	if l.serviceConfig.Port > 0 {
		portInfo := fmt.Sprintf("%s %s",
			l.styles.FooterLabel.Render("Port:"),
			l.styles.FooterValue.Render(fmt.Sprintf("%d", l.serviceConfig.Port)))
		parts = append(parts, portInfo)
	}

	// Env info (show first 3 vars)
	if len(l.serviceConfig.Env) > 0 {
		envVars := l.serviceConfig.Env
		if len(envVars) > 3 {
			envVars = envVars[:3]
		}
		envStr := strings.Join(envVars, ", ")
		if len(l.serviceConfig.Env) > 3 {
			envStr += fmt.Sprintf(" (+%d more)", len(l.serviceConfig.Env)-3)
		}
		envInfo := fmt.Sprintf("%s %s",
			l.styles.FooterLabel.Render("Env:"),
			l.styles.FooterValue.Render(envStr))
		parts = append(parts, envInfo)
	}

	// Dependencies
	if len(l.serviceConfig.DependsOn) > 0 {
		depsInfo := fmt.Sprintf("%s %s",
			l.styles.FooterLabel.Render("Deps:"),
			l.styles.FooterValue.Render(strings.Join(l.serviceConfig.DependsOn, ", ")))
		parts = append(parts, depsInfo)
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " │ ")
}
