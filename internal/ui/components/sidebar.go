package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/paralerdev/paraler/internal/config"
	"github.com/paralerdev/paraler/internal/log"
	"github.com/paralerdev/paraler/internal/process"
	"github.com/charmbracelet/lipgloss"
)

// SidebarItem represents a single item in the sidebar
type SidebarItem struct {
	ID        config.ServiceID
	IsProject bool
	Name      string
}

// Sidebar is the service list component
type Sidebar struct {
	items       []SidebarItem
	selected    int
	width       int
	height      int
	focused     bool
	styles      SidebarStyles
	multiSelect map[int]bool // Selected items for multi-select mode
}

// SidebarStyles contains sidebar-specific styles
type SidebarStyles struct {
	Container        lipgloss.Style
	Title            lipgloss.Style
	TitleFocused     lipgloss.Style
	ProjectHeader    lipgloss.Style
	Item             lipgloss.Style
	ItemSelected     lipgloss.Style
	ItemMultiSelect  lipgloss.Style
	StatusRunning    lipgloss.Style
	StatusStopped    lipgloss.Style
	StatusFailed     lipgloss.Style
	StatusStarting   lipgloss.Style
	StatusIndicator  lipgloss.Style
	HealthHealthy    lipgloss.Style
	HealthUnhealthy  lipgloss.Style
	HealthUnknown    lipgloss.Style
	MultiSelectMark  lipgloss.Style
	ErrorBadge       lipgloss.Style
}

// DefaultSidebarStyles returns the default sidebar styles
func DefaultSidebarStyles() SidebarStyles {
	return SidebarStyles{
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
		ProjectHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8B5CF6")).
			MarginTop(1),
		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			PaddingLeft(2),
		ItemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			Background(lipgloss.Color("#1F2937")).
			Bold(true).
			PaddingLeft(2),
		StatusRunning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")),
		StatusStopped: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		StatusFailed: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")),
		StatusStarting: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")),
		StatusIndicator: lipgloss.NewStyle().
			Bold(true),
		HealthHealthy: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")),
		HealthUnhealthy: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")),
		HealthUnknown: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		ItemMultiSelect: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			Background(lipgloss.Color("#374151")).
			PaddingLeft(2),
		MultiSelectMark: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true),
		ErrorBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true),
	}
}

// NewSidebar creates a new sidebar
func NewSidebar(cfg *config.Config) *Sidebar {
	s := &Sidebar{
		styles:      DefaultSidebarStyles(),
		multiSelect: make(map[int]bool),
	}
	s.buildItems(cfg)
	return s
}

// buildItems builds the sidebar items from config
func (s *Sidebar) buildItems(cfg *config.Config) {
	s.items = nil

	// Sort project names for consistent ordering
	projectNames := make([]string, 0, len(cfg.Projects))
	for name := range cfg.Projects {
		projectNames = append(projectNames, name)
	}
	sort.Strings(projectNames)

	for _, projectName := range projectNames {
		project := cfg.Projects[projectName]

		// Add project header
		s.items = append(s.items, SidebarItem{
			ID:        config.ServiceID{Project: projectName},
			IsProject: true,
			Name:      projectName,
		})

		// Sort service names
		serviceNames := make([]string, 0, len(project.Services))
		for name := range project.Services {
			serviceNames = append(serviceNames, name)
		}
		sort.Strings(serviceNames)

		// Add services
		for _, serviceName := range serviceNames {
			s.items = append(s.items, SidebarItem{
				ID: config.ServiceID{
					Project: projectName,
					Service: serviceName,
				},
				IsProject: false,
				Name:      serviceName,
			})
		}
	}
}

// SetSize sets the sidebar dimensions
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetFocused sets the focus state
func (s *Sidebar) SetFocused(focused bool) {
	s.focused = focused
}

// MoveUp moves selection up
func (s *Sidebar) MoveUp() {
	if s.selected > 0 {
		s.selected--
		// Skip project headers
		if s.items[s.selected].IsProject && s.selected > 0 {
			s.selected--
		}
	}
}

// MoveDown moves selection down
func (s *Sidebar) MoveDown() {
	if s.selected < len(s.items)-1 {
		s.selected++
		// Skip project headers
		if s.items[s.selected].IsProject && s.selected < len(s.items)-1 {
			s.selected++
		}
	}
}

// Selected returns the currently selected service ID
func (s *Sidebar) Selected() config.ServiceID {
	if s.selected >= 0 && s.selected < len(s.items) {
		item := s.items[s.selected]
		if !item.IsProject {
			return item.ID
		}
	}
	return config.ServiceID{}
}

// SelectedIndex returns the selected index
func (s *Sidebar) SelectedIndex() int {
	return s.selected
}

// View renders the sidebar
func (s *Sidebar) View(manager *process.Manager, logBuffer *log.Buffer) string {
	var b strings.Builder

	// Title
	title := "Services"
	if s.focused {
		b.WriteString(s.styles.TitleFocused.Render(title))
	} else {
		b.WriteString(s.styles.Title.Render(title))
	}
	b.WriteString("\n")

	// Calculate available height for items
	availableHeight := s.height - 4 // Title + borders

	// Render items
	for i, item := range s.items {
		if i >= availableHeight {
			break
		}

		if item.IsProject {
			// Project header
			b.WriteString(s.styles.ProjectHeader.Render("▸ " + item.Name))
		} else {
			// Service item
			proc := manager.Get(item.ID)
			status := process.StatusStopped
			health := process.HealthUnknown
			if proc != nil {
				status = proc.Status()
				health = proc.Health()
			}

			// Status indicator
			indicator := s.getStatusIndicator(status)

			// Health indicator (only show for running services)
			healthIndicator := ""
			if status == process.StatusRunning {
				healthIndicator = " " + s.getHealthIndicator(health)
			}

			// Multi-select marker
			multiMarker := " "
			if s.IsMultiSelected(i) {
				multiMarker = s.styles.MultiSelectMark.Render("▪")
			}

			// Apply custom color if set
			serviceName := item.Name
			if proc != nil && proc.Config.Color != "" {
				colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(proc.Config.Color))
				serviceName = colorStyle.Render(item.Name)
			}

			// Error badge (only show if errors exist)
			errorBadge := ""
			if logBuffer != nil {
				errorCount := logBuffer.ErrorCount(item.ID)
				if errorCount > 0 {
					if errorCount > 99 {
						errorBadge = s.styles.ErrorBadge.Render(" [!99+]")
					} else {
						errorBadge = s.styles.ErrorBadge.Render(fmt.Sprintf(" [!%d]", errorCount))
					}
				}
			}

			// Item text
			text := fmt.Sprintf("%s%s %s%s%s", multiMarker, indicator, serviceName, healthIndicator, errorBadge)

			// Apply style based on selection
			if i == s.selected {
				// Pad to full width
				text = s.padRight(text, s.width-4)
				b.WriteString(s.styles.ItemSelected.Render(text))
			} else if s.IsMultiSelected(i) {
				text = s.padRight(text, s.width-4)
				b.WriteString(s.styles.ItemMultiSelect.Render(text))
			} else {
				b.WriteString(s.styles.Item.Render(text))
			}
		}
		b.WriteString("\n")
	}

	// Apply container style
	content := b.String()
	if s.focused {
		s.styles.Container = s.styles.Container.BorderForeground(lipgloss.Color("#8B5CF6"))
	} else {
		s.styles.Container = s.styles.Container.BorderForeground(lipgloss.Color("#374151"))
	}

	return s.styles.Container.
		Width(s.width).
		Height(s.height).
		Render(content)
}

// getStatusIndicator returns the status indicator character
func (s *Sidebar) getStatusIndicator(status process.Status) string {
	switch status {
	case process.StatusRunning:
		return s.styles.StatusRunning.Render("●")
	case process.StatusStarting:
		return s.styles.StatusStarting.Render("○")
	case process.StatusStopping:
		return s.styles.StatusStarting.Render("◐")
	case process.StatusFailed:
		return s.styles.StatusFailed.Render("●")
	default:
		return s.styles.StatusStopped.Render("○")
	}
}

// getHealthIndicator returns the health indicator character
func (s *Sidebar) getHealthIndicator(health process.HealthStatus) string {
	switch health {
	case process.HealthHealthy:
		return s.styles.HealthHealthy.Render("✓")
	case process.HealthUnhealthy:
		return s.styles.HealthUnhealthy.Render("✗")
	default:
		return s.styles.HealthUnknown.Render("?")
	}
}

// padRight pads a string to the specified width
func (s *Sidebar) padRight(str string, width int) string {
	// Account for ANSI escape codes
	visibleLen := lipgloss.Width(str)
	if visibleLen >= width {
		return str
	}
	return str + strings.Repeat(" ", width-visibleLen)
}

// ItemCount returns the number of items
func (s *Sidebar) ItemCount() int {
	return len(s.items)
}

// ServiceCount returns the number of services (excluding project headers)
func (s *Sidebar) ServiceCount() int {
	count := 0
	for _, item := range s.items {
		if !item.IsProject {
			count++
		}
	}
	return count
}

// SelectFirst selects the first service
func (s *Sidebar) SelectFirst() {
	for i, item := range s.items {
		if !item.IsProject {
			s.selected = i
			return
		}
	}
}

// ToggleMultiSelect toggles multi-select for the current item
func (s *Sidebar) ToggleMultiSelect() {
	if s.selected >= 0 && s.selected < len(s.items) {
		item := s.items[s.selected]
		if !item.IsProject {
			s.multiSelect[s.selected] = !s.multiSelect[s.selected]
			if !s.multiSelect[s.selected] {
				delete(s.multiSelect, s.selected)
			}
		}
	}
}

// ClearMultiSelect clears all multi-selections
func (s *Sidebar) ClearMultiSelect() {
	s.multiSelect = make(map[int]bool)
}

// HasMultiSelect returns true if there are multi-selected items
func (s *Sidebar) HasMultiSelect() bool {
	return len(s.multiSelect) > 0
}

// GetMultiSelected returns all multi-selected service IDs
func (s *Sidebar) GetMultiSelected() []config.ServiceID {
	var ids []config.ServiceID
	for i := range s.multiSelect {
		if i >= 0 && i < len(s.items) {
			item := s.items[i]
			if !item.IsProject {
				ids = append(ids, item.ID)
			}
		}
	}
	return ids
}

// IsMultiSelected returns true if the index is multi-selected
func (s *Sidebar) IsMultiSelected(index int) bool {
	return s.multiSelect[index]
}
