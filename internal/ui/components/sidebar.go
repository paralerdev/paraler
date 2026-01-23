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
	SelectionMarker  lipgloss.Style
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
			Foreground(lipgloss.Color("#F9FAFB")),
		ItemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			Bold(true),
		SelectionMarker: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true),
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
			Background(lipgloss.Color("#374151")),
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
		newSelected := s.selected - 1
		// Skip project headers
		if s.items[newSelected].IsProject {
			if newSelected > 0 {
				newSelected--
			} else {
				// Can't go up, stay at current position
				return
			}
		}
		s.selected = newSelected
	}
}

// MoveDown moves selection down
func (s *Sidebar) MoveDown() {
	if s.selected < len(s.items)-1 {
		newSelected := s.selected + 1
		// Skip project headers
		if s.items[newSelected].IsProject {
			if newSelected < len(s.items)-1 {
				newSelected++
			} else {
				// Can't go down, stay at current position
				return
			}
		}
		s.selected = newSelected
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

// SelectedItem returns the currently selected item
func (s *Sidebar) SelectedItem() *SidebarItem {
	if s.selected >= 0 && s.selected < len(s.items) {
		return &s.items[s.selected]
	}
	return nil
}

// IsProjectSelected returns true if a project header is selected
func (s *Sidebar) IsProjectSelected() bool {
	item := s.SelectedItem()
	return item != nil && item.IsProject
}

// SelectedProjectName returns the project name of the selected item
func (s *Sidebar) SelectedProjectName() string {
	item := s.SelectedItem()
	if item == nil {
		return ""
	}
	return item.ID.Project
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
			// Project header (not selectable)
			projectName := item.Name
			maxProjectLen := s.width - 6 // borders + "▸ " prefix + margin
			if maxProjectLen < 3 {
				maxProjectLen = 3
			}
			if len(projectName) > maxProjectLen {
				projectName = projectName[:maxProjectLen-1] + "…"
			}
			b.WriteString(s.styles.ProjectHeader.Render("▸ " + projectName))
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

			serviceName := item.Name

			// Error badge (only show if errors exist)
			errorBadge := ""
			errorBadgeLen := 0
			if logBuffer != nil {
				errorCount := logBuffer.ErrorCount(item.ID)
				if errorCount > 0 {
					if errorCount > 99 {
						errorBadge = s.styles.ErrorBadge.Render(" !")
						errorBadgeLen = 2
					} else {
						errorBadge = s.styles.ErrorBadge.Render(fmt.Sprintf(" !%d", errorCount))
						errorBadgeLen = 2 + len(fmt.Sprintf("%d", errorCount))
					}
				}
			}

			// Selection marker
			selMarker := "  "
			if i == s.selected {
				selMarker = s.styles.SelectionMarker.Render("› ")
			}

			// Calculate available width for service name
			// prefix: selMarker(2) + multiMarker(1) + indicator(1) + space(1) = 5
			// suffix: healthIndicator(0-2) + errorBadge(0-4)
			prefixLen := 5
			suffixLen := len(healthIndicator) + errorBadgeLen
			innerWidth := s.width - 2 // borders
			maxNameLen := innerWidth - prefixLen - suffixLen - 1
			if maxNameLen < 3 {
				maxNameLen = 3
			}

			// Truncate service name if needed
			if len(serviceName) > maxNameLen {
				serviceName = serviceName[:maxNameLen-1] + "…"
			}

			// Item text
			text := fmt.Sprintf("%s%s%s %s%s%s", selMarker, multiMarker, indicator, serviceName, healthIndicator, errorBadge)

			// Apply style
			if i == s.selected || s.IsMultiSelected(i) {
				b.WriteString(text)
			} else {
				b.WriteString(s.styles.Item.Render(text))
			}
		}
		b.WriteString("\n")
	}

	// Build content with manual borders
	content := b.String()
	return s.renderWithBorder(content)
}

// renderWithBorder renders content with manual box-drawing borders
func (s *Sidebar) renderWithBorder(content string) string {
	lines := strings.Split(content, "\n")
	innerWidth := s.width - 2   // Account for left/right borders
	innerHeight := s.height - 2 // Account for top/bottom borders

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
	if s.focused {
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
// Returns empty string for unknown status (no health check configured)
func (s *Sidebar) getHealthIndicator(health process.HealthStatus) string {
	switch health {
	case process.HealthHealthy:
		return s.styles.HealthHealthy.Render("✓")
	case process.HealthUnhealthy:
		return s.styles.HealthUnhealthy.Render("✗")
	default:
		return ""
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
