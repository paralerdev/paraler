package ui

import (
	"hash/fnv"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	ColorPrimary   = lipgloss.Color("#8B5CF6")
	ColorSecondary = lipgloss.Color("#6366F1")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorWarning   = lipgloss.Color("#F59E0B")
	ColorDanger    = lipgloss.Color("#EF4444")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorBorder    = lipgloss.Color("#374151")
	ColorBg        = lipgloss.Color("#111827")
	ColorBgLight   = lipgloss.Color("#1F2937")
	ColorText      = lipgloss.Color("#F9FAFB")
	ColorTextMuted = lipgloss.Color("#9CA3AF")
)

// Service colors for differentiation
var serviceColors = []lipgloss.Color{
	lipgloss.Color("#8B5CF6"), // Purple
	lipgloss.Color("#10B981"), // Green
	lipgloss.Color("#F59E0B"), // Yellow
	lipgloss.Color("#3B82F6"), // Blue
	lipgloss.Color("#EC4899"), // Pink
	lipgloss.Color("#14B8A6"), // Teal
	lipgloss.Color("#F97316"), // Orange
	lipgloss.Color("#6366F1"), // Indigo
}

// GetServiceColor returns a consistent color for a service based on its name
func GetServiceColor(name string) lipgloss.Color {
	h := fnv.New32a()
	h.Write([]byte(name))
	idx := h.Sum32() % uint32(len(serviceColors))
	return serviceColors[idx]
}

// Styles contains all UI styles
type Styles struct {
	// Layout
	App     lipgloss.Style
	Sidebar lipgloss.Style
	Main    lipgloss.Style
	Status  lipgloss.Style

	// Sidebar items
	ProjectHeader    lipgloss.Style
	ServiceItem      lipgloss.Style
	ServiceSelected  lipgloss.Style
	ServiceRunning   lipgloss.Style
	ServiceStopped   lipgloss.Style
	ServiceFailed    lipgloss.Style
	ServiceIndicator lipgloss.Style

	// Logs
	LogLine       lipgloss.Style
	LogTimestamp  lipgloss.Style
	LogStderr     lipgloss.Style
	LogFilter     lipgloss.Style
	LogFilterText lipgloss.Style

	// Status bar
	StatusText    lipgloss.Style
	StatusKey     lipgloss.Style
	StatusValue   lipgloss.Style
	StatusRunning lipgloss.Style
	StatusStopped lipgloss.Style

	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
	HelpSep  lipgloss.Style

	// Title
	Title      lipgloss.Style
	TitleFocus lipgloss.Style
}

// DefaultStyles returns the default UI styles
func DefaultStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Background(ColorBg),

		Sidebar: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1),

		Main: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1),

		Status: lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Padding(0, 1),

		ProjectHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginTop(1).
			MarginBottom(0),

		ServiceItem: lipgloss.NewStyle().
			Foreground(ColorText).
			PaddingLeft(2),

		ServiceSelected: lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBgLight).
			Bold(true).
			PaddingLeft(2),

		ServiceRunning: lipgloss.NewStyle().
			Foreground(ColorSuccess),

		ServiceStopped: lipgloss.NewStyle().
			Foreground(ColorMuted),

		ServiceFailed: lipgloss.NewStyle().
			Foreground(ColorDanger),

		ServiceIndicator: lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true),

		LogLine: lipgloss.NewStyle().
			Foreground(ColorText),

		LogTimestamp: lipgloss.NewStyle().
			Foreground(ColorMuted),

		LogStderr: lipgloss.NewStyle().
			Foreground(ColorDanger),

		LogFilter: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1),

		LogFilterText: lipgloss.NewStyle().
			Foreground(ColorPrimary),

		StatusText: lipgloss.NewStyle().
			Foreground(ColorTextMuted),

		StatusKey: lipgloss.NewStyle().
			Foreground(ColorMuted).
			Bold(true),

		StatusValue: lipgloss.NewStyle().
			Foreground(ColorText),

		StatusRunning: lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true),

		StatusStopped: lipgloss.NewStyle().
			Foreground(ColorMuted),

		HelpKey: lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true),

		HelpDesc: lipgloss.NewStyle().
			Foreground(ColorMuted),

		HelpSep: lipgloss.NewStyle().
			Foreground(ColorBorder),

		Title: lipgloss.NewStyle().
			Foreground(ColorMuted).
			Bold(true).
			Padding(0, 1),

		TitleFocus: lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1),
	}
}
