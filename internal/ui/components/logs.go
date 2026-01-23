package components

import (
	"fmt"
	"strings"

	"github.com/paralerdev/paraler/internal/config"
	"github.com/paralerdev/paraler/internal/log"
	"github.com/paralerdev/paraler/internal/process"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// LogPanel displays logs for a selected service
type LogPanel struct {
	filterInput   textinput.Model
	serviceID     config.ServiceID
	serviceConfig *config.Service
	serviceStatus process.Status
	filter        string
	filtering     bool
	autoScroll    bool
	scrollOffset  int
	width         int
	height        int
	focused       bool
	styles        LogPanelStyles
	lines         []string
	rawLines      []string // Lines without styling for copying
	viewHeight    int

	// Copy mode state
	copyMode        bool
	copyCursor      int  // Current cursor position in copy mode
	copySelecting   bool // Whether we're selecting (after pressing v)
	copySelectStart int  // Start of selection
}

// LogPanelStyles contains log panel styles
type LogPanelStyles struct {
	Container       lipgloss.Style
	Title           lipgloss.Style
	TitleFocused    lipgloss.Style
	Line            lipgloss.Style
	LineStderr      lipgloss.Style
	Timestamp       lipgloss.Style
	FilterPrompt    lipgloss.Style
	FilterInput     lipgloss.Style
	NoLogs          lipgloss.Style
	ServiceColor    lipgloss.Style
	Footer          lipgloss.Style
	FooterLabel     lipgloss.Style
	FooterValue     lipgloss.Style
	CopyModeCursor  lipgloss.Style
	CopyModeSelect  lipgloss.Style
	CopyModeStatus  lipgloss.Style
	StatusRunning   lipgloss.Style
	StatusStopped   lipgloss.Style
	StatusStarting  lipgloss.Style
	StatusFailed    lipgloss.Style
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
		CopyModeCursor: lipgloss.NewStyle().
			Background(lipgloss.Color("#374151")),
		CopyModeSelect: lipgloss.NewStyle().
			Background(lipgloss.Color("#4C1D95")).
			Foreground(lipgloss.Color("#F9FAFB")),
		CopyModeStatus: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true),
		StatusRunning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true),
		StatusStopped: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		StatusStarting: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")),
		StatusFailed: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true),
	}
}

// NewLogPanel creates a new log panel
func NewLogPanel() *LogPanel {
	ti := textinput.New()
	ti.Placeholder = "Filter logs..."
	ti.CharLimit = 100

	return &LogPanel{
		filterInput: ti,
		autoScroll:  true,
		styles:      DefaultLogPanelStyles(),
	}
}

// SetSize sets the panel dimensions
func (l *LogPanel) SetSize(width, height int) {
	l.width = width
	l.height = height

	// Calculate view height for borders and title
	vpHeight := height - 4
	if l.filtering {
		vpHeight -= 1
	}
	if vpHeight < 1 {
		vpHeight = 1
	}

	l.viewHeight = vpHeight
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

// SetStatus sets the current service status
func (l *LogPanel) SetStatus(status process.Status) {
	l.serviceStatus = status
}

// formatStatus returns a formatted status string with color
func (l *LogPanel) formatStatus() string {
	if l.serviceID.Service == "" {
		return ""
	}

	switch l.serviceStatus {
	case process.StatusRunning:
		return l.styles.StatusRunning.Render("[running]")
	case process.StatusStarting:
		return l.styles.StatusStarting.Render("[starting]")
	case process.StatusStopping:
		return l.styles.StatusStarting.Render("[stopping]")
	case process.StatusFailed:
		return l.styles.StatusFailed.Render("[failed]")
	default:
		return l.styles.StatusStopped.Render("[stopped]")
	}
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

// LogLevel represents detected log level
type LogLevel int

const (
	LogLevelNormal LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Update updates the log panel with new entries
func (l *LogPanel) Update(buffer *log.Buffer) {
	// Don't update in copy mode (freeze logs)
	if l.copyMode {
		return
	}

	entries := buffer.GetFiltered(l.serviceID, l.filter)

	l.lines = nil
	l.rawLines = nil
	for _, entry := range entries {
		// Sanitize the line - remove ANSI codes and control chars
		cleanLine := sanitizeLine(entry.Line)

		// Store raw line for copying
		rawLine := fmt.Sprintf("%s %s", entry.Timestamp.Format("15:04:05"), cleanLine)
		l.rawLines = append(l.rawLines, rawLine)

		// Detect log level
		level := detectLogLevel(cleanLine)

		// Format timestamp with service color if available
		timestamp := l.formatTimestamp(entry.Timestamp.Format("15:04:05"))

		// Format line based on level and stderr
		var line string
		if entry.IsStderr {
			line = l.styles.LineStderr.Render(cleanLine)
		} else {
			line = l.formatLineByLevel(cleanLine, level)
		}

		l.lines = append(l.lines, fmt.Sprintf("%s %s", timestamp, line))
	}

	if l.autoScroll {
		l.scrollToBottom()
	}
}

// formatTimestamp formats timestamp with service color if available
func (l *LogPanel) formatTimestamp(ts string) string {
	if l.serviceConfig != nil && l.serviceConfig.Color != "" {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(l.serviceConfig.Color))
		return style.Render(ts)
	}
	return l.styles.Timestamp.Render(ts)
}

// formatLineByLevel applies color based on log level
func (l *LogPanel) formatLineByLevel(line string, level LogLevel) string {
	switch level {
	case LogLevelError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(line)
	case LogLevelWarn:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render(line)
	case LogLevelDebug:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(line)
	default:
		return l.styles.Line.Render(line)
	}
}

// detectLogLevel detects the log level from line content
func detectLogLevel(line string) LogLevel {
	upper := strings.ToUpper(line)

	// Check for error indicators
	if strings.Contains(upper, "ERROR") ||
		strings.Contains(upper, "FATAL") ||
		strings.Contains(upper, "EXCEPTION") ||
		strings.Contains(upper, "FAILED") {
		return LogLevelError
	}

	// Check for warning indicators
	if strings.Contains(upper, "WARN") ||
		strings.Contains(upper, "WARNING") {
		return LogLevelWarn
	}

	// Check for debug indicators
	if strings.Contains(upper, "DEBUG") ||
		strings.Contains(upper, "TRACE") ||
		strings.Contains(upper, "VERBOSE") {
		return LogLevelDebug
	}

	return LogLevelNormal
}

// sanitizeLine removes control characters and ANSI codes that break the layout
func sanitizeLine(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	inEscape := false
	for _, r := range s {
		// Skip ANSI escape sequences
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		// Skip carriage return and newline
		if r == '\r' || r == '\n' {
			continue
		}
		// Replace tab with spaces
		if r == '\t' {
			result.WriteString("    ")
			continue
		}
		// Skip other control characters
		if r < 32 {
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}

// scrollToBottom scrolls to the bottom of the logs
func (l *LogPanel) scrollToBottom() {
	maxOffset := len(l.lines) - l.viewHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	l.scrollOffset = maxOffset
}

// ScrollUp scrolls up
func (l *LogPanel) ScrollUp() {
	l.autoScroll = false
	if l.scrollOffset > 0 {
		l.scrollOffset--
	}
}

// ScrollDown scrolls down
func (l *LogPanel) ScrollDown() {
	maxOffset := len(l.lines) - l.viewHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if l.scrollOffset < maxOffset {
		l.scrollOffset++
	}
	if l.scrollOffset >= maxOffset {
		l.autoScroll = true
	}
}

// PageUp scrolls up a page
func (l *LogPanel) PageUp() {
	l.autoScroll = false
	l.scrollOffset -= l.viewHeight / 2
	if l.scrollOffset < 0 {
		l.scrollOffset = 0
	}
}

// PageDown scrolls down a page
func (l *LogPanel) PageDown() {
	maxOffset := len(l.lines) - l.viewHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	l.scrollOffset += l.viewHeight / 2
	if l.scrollOffset > maxOffset {
		l.scrollOffset = maxOffset
	}
	if l.scrollOffset >= maxOffset {
		l.autoScroll = true
	}
}

// GoToTop scrolls to top
func (l *LogPanel) GoToTop() {
	l.autoScroll = false
	l.scrollOffset = 0
}

// GoToBottom scrolls to bottom
func (l *LogPanel) GoToBottom() {
	l.autoScroll = true
	l.scrollToBottom()
}

// Copy Mode methods

// EnterCopyMode enters copy mode
func (l *LogPanel) EnterCopyMode() {
	if len(l.lines) == 0 {
		return
	}
	l.copyMode = true
	l.autoScroll = false
	l.copySelecting = false
	// Position cursor at the last visible line
	l.copyCursor = l.scrollOffset + l.viewHeight - 1
	if l.copyCursor >= len(l.lines) {
		l.copyCursor = len(l.lines) - 1
	}
}

// ExitCopyMode exits copy mode
func (l *LogPanel) ExitCopyMode() {
	l.copyMode = false
	l.copySelecting = false
	l.autoScroll = true
}

// IsCopyMode returns true if in copy mode
func (l *LogPanel) IsCopyMode() bool {
	return l.copyMode
}

// CopyModeCursorUp moves cursor up in copy mode
func (l *LogPanel) CopyModeCursorUp() {
	if !l.copyMode {
		return
	}
	if l.copyCursor > 0 {
		l.copyCursor--
		// Scroll if cursor goes above visible area
		if l.copyCursor < l.scrollOffset {
			l.scrollOffset = l.copyCursor
		}
	}
}

// CopyModeCursorDown moves cursor down in copy mode
func (l *LogPanel) CopyModeCursorDown() {
	if !l.copyMode {
		return
	}
	if l.copyCursor < len(l.lines)-1 {
		l.copyCursor++
		// Scroll if cursor goes below visible area
		if l.copyCursor >= l.scrollOffset+l.viewHeight {
			l.scrollOffset = l.copyCursor - l.viewHeight + 1
		}
	}
}

// CopyModeToggleSelect toggles selection in copy mode
func (l *LogPanel) CopyModeToggleSelect() {
	if !l.copyMode {
		return
	}
	if l.copySelecting {
		l.copySelecting = false
	} else {
		l.copySelecting = true
		l.copySelectStart = l.copyCursor
	}
}

// CopyModeGetSelectedText returns the selected text for copying
func (l *LogPanel) CopyModeGetSelectedText() string {
	if !l.copyMode || len(l.rawLines) == 0 {
		return ""
	}

	var start, end int
	if l.copySelecting {
		start = l.copySelectStart
		end = l.copyCursor
		if start > end {
			start, end = end, start
		}
	} else {
		// Just copy current line
		start = l.copyCursor
		end = l.copyCursor
	}

	// Bounds check
	if start < 0 {
		start = 0
	}
	if end >= len(l.rawLines) {
		end = len(l.rawLines) - 1
	}

	var lines []string
	for i := start; i <= end; i++ {
		lines = append(lines, l.rawLines[i])
	}

	return strings.Join(lines, "\n")
}

// CopyModeIsLineSelected returns true if the line at index is selected
func (l *LogPanel) CopyModeIsLineSelected(index int) bool {
	if !l.copyMode {
		return false
	}
	if !l.copySelecting {
		return index == l.copyCursor
	}
	start, end := l.copySelectStart, l.copyCursor
	if start > end {
		start, end = end, start
	}
	return index >= start && index <= end
}

// CopyModeIsCursor returns true if the line at index is the cursor
func (l *LogPanel) CopyModeIsCursor(index int) bool {
	return l.copyMode && index == l.copyCursor
}

// View renders the log panel
func (l *LogPanel) View(buffer *log.Buffer) string {
	var b strings.Builder

	// Title with status
	title := "Logs"
	if l.serviceID.Service != "" {
		title = fmt.Sprintf("Logs: %s/%s", l.serviceID.Project, l.serviceID.Service)
	}

	// Add status indicator
	statusText := l.formatStatus()
	if statusText != "" {
		title += " " + statusText
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

	// Calculate content width (account for borders)
	contentWidth := l.width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Render log lines
	if len(l.lines) == 0 {
		noLogsMsg := "No logs yet. Start a service to see output."
		if l.filter != "" {
			noLogsMsg = "No logs match the filter."
		}
		b.WriteString(l.styles.NoLogs.Render(noLogsMsg))
	} else {
		// Calculate visible range
		start := l.scrollOffset
		end := start + l.viewHeight
		if end > len(l.lines) {
			end = len(l.lines)
		}
		if start > len(l.lines) {
			start = len(l.lines)
		}

		// Render visible lines with truncation
		for i := start; i < end; i++ {
			if i > start {
				b.WriteString("\n")
			}
			line := l.lines[i]
			// Truncate line to fit width
			if lipgloss.Width(line) > contentWidth {
				line = truncateString(line, contentWidth)
			}

			// Apply copy mode highlighting
			if l.copyMode {
				if l.CopyModeIsLineSelected(i) {
					// Use raw line for consistent styling in copy mode
					rawLine := ""
					if i < len(l.rawLines) {
						rawLine = l.rawLines[i]
						if len(rawLine) > contentWidth {
							rawLine = rawLine[:contentWidth-1] + "…"
						}
					}
					line = l.styles.CopyModeSelect.Render(rawLine)
					// Pad to width
					padLen := contentWidth - lipgloss.Width(line)
					if padLen > 0 {
						line = l.styles.CopyModeSelect.Render(rawLine + strings.Repeat(" ", padLen))
					}
				}
			}

			b.WriteString(line)
		}
	}

	// Filter input
	if l.filtering {
		b.WriteString("\n")
		b.WriteString(l.styles.FilterPrompt.Render("/"))
		b.WriteString(l.filterInput.View())
	}

	// Copy mode status
	if l.copyMode {
		b.WriteString("\n")
		status := "[COPY] "
		if l.copySelecting {
			lines := l.copyCursor - l.copySelectStart
			if lines < 0 {
				lines = -lines
			}
			lines++
			status += fmt.Sprintf("%d lines selected │ ", lines)
		}
		status += "↑↓:move  v:select  y:copy  Esc:exit"
		b.WriteString(l.styles.CopyModeStatus.Render(status))
	} else if l.serviceConfig != nil && !l.filtering {
		// Footer with env/port info (only when not in copy mode)
		footer := l.renderFooter()
		if footer != "" {
			b.WriteString("\n")
			b.WriteString(l.styles.Footer.Render(footer))
		}
	}

	// Build content with manual borders
	content := b.String()
	return l.renderWithBorder(content)
}

// renderWithBorder renders content with manual box-drawing borders
func (l *LogPanel) renderWithBorder(content string) string {
	lines := strings.Split(content, "\n")
	innerWidth := l.width - 2   // Account for left/right borders
	innerHeight := l.height - 2 // Account for top/bottom borders

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Pad or trim lines to fit
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	// Border color
	borderColor := "#374151"
	if l.focused {
		borderColor = "#8B5CF6"
	}
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))

	var result strings.Builder

	// Top border
	result.WriteString(borderStyle.Render("╭" + strings.Repeat("─", innerWidth) + "╮"))
	result.WriteString("\n")

	// Content lines with side borders
	for _, line := range lines {
		result.WriteString(borderStyle.Render("│"))
		// Pad line to inner width
		visWidth := lipgloss.Width(line)
		if visWidth < innerWidth {
			line = line + strings.Repeat(" ", innerWidth-visWidth)
		}
		result.WriteString(line)
		result.WriteString(borderStyle.Render("│"))
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString(borderStyle.Render("╰" + strings.Repeat("─", innerWidth) + "╯"))

	return result.String()
}

// ServiceID returns the current service ID
func (l *LogPanel) ServiceID() config.ServiceID {
	return l.serviceID
}

// truncateString truncates a string to maxWidth, handling ANSI escape codes
func truncateString(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	// Fast path: if already fits, return as-is
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Truncate character by character, preserving ANSI codes
	var result strings.Builder
	visibleWidth := 0
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}

		if inEscape {
			result.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		// Regular character - check if it fits
		charWidth := 1
		if r > 127 {
			charWidth = 2 // Assume wide characters for safety
		}

		if visibleWidth+charWidth > maxWidth-1 {
			result.WriteString("…")
			break
		}

		result.WriteRune(r)
		visibleWidth += charWidth
	}

	// Reset any open ANSI sequences
	result.WriteString("\x1b[0m")

	return result.String()
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
