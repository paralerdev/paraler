package components

import (
	"fmt"
	"strings"

	"github.com/paralerdev/paraler/internal/discovery"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// AddProjectState represents the current state of the add project flow
type AddProjectState int

const (
	AddProjectStateInput AddProjectState = iota
	AddProjectStateScanning
	AddProjectStatePreview
	AddProjectStateError
	AddProjectStateDone
)

// AddProjectModal is a modal for adding new projects
type AddProjectModal struct {
	state           AddProjectState
	pathInput       textinput.Model
	pathCompleter   *PathCompleter
	suggestions     []string
	suggestionIndex int  // Currently selected suggestion (-1 = none)
	detected        *discovery.DetectedProject
	selected        map[int]bool // Selected services
	cursor          int
	error           string
	width           int
	height          int
	styles          AddProjectStyles
}

// AddProjectStyles contains styles for the modal
type AddProjectStyles struct {
	Container     lipgloss.Style
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	Input         lipgloss.Style
	Label         lipgloss.Style
	Service       lipgloss.Style
	ServiceSel    lipgloss.Style
	Checkbox      lipgloss.Style
	CheckboxSel   lipgloss.Style
	Framework     lipgloss.Style
	Command       lipgloss.Style
	Error         lipgloss.Style
	Help          lipgloss.Style
	Button        lipgloss.Style
	ButtonActive  lipgloss.Style
	Suggestion    lipgloss.Style
	SuggestionSel lipgloss.Style
}

// DefaultAddProjectStyles returns default styles
func DefaultAddProjectStyles() AddProjectStyles {
	return AddProjectStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#8B5CF6")).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F9FAFB")).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			MarginBottom(1),
		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")),
		Service: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")),
		ServiceSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			Background(lipgloss.Color("#1F2937")),
		Checkbox: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		CheckboxSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")),
		Framework: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")),
		Command: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1),
		Button: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Padding(0, 2),
		ButtonActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			Background(lipgloss.Color("#8B5CF6")).
			Padding(0, 2),
		Suggestion: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			PaddingLeft(2),
		SuggestionSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			PaddingLeft(2),
	}
}

// NewAddProjectModal creates a new add project modal
func NewAddProjectModal() *AddProjectModal {
	ti := textinput.New()
	ti.Placeholder = "~/projects/myproject (Tab to autocomplete)"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return &AddProjectModal{
		state:           AddProjectStateInput,
		pathInput:       ti,
		pathCompleter:   NewPathCompleter(),
		suggestionIndex: -1,
		selected:        make(map[int]bool),
		styles:          DefaultAddProjectStyles(),
	}
}

// CompleteTab handles Tab key for path autocompletion
func (m *AddProjectModal) CompleteTab() {
	current := m.pathInput.Value()
	completed := m.pathCompleter.GetNext(current)
	m.pathInput.SetValue(completed)
	m.pathInput.CursorEnd()
	m.UpdateSuggestions()
}

// UpdateSuggestions updates the path suggestions
func (m *AddProjectModal) UpdateSuggestions() {
	current := m.pathInput.Value()
	m.suggestions = m.pathCompleter.GetSuggestions(current, 5)
	// Reset selection when suggestions change
	m.suggestionIndex = -1
}

// SuggestionUp moves selection up in suggestions
func (m *AddProjectModal) SuggestionUp() {
	if len(m.suggestions) == 0 {
		return
	}
	if m.suggestionIndex <= 0 {
		m.suggestionIndex = -1 // Back to input
	} else {
		m.suggestionIndex--
	}
}

// SuggestionDown moves selection down in suggestions
func (m *AddProjectModal) SuggestionDown() {
	if len(m.suggestions) == 0 {
		return
	}
	if m.suggestionIndex < len(m.suggestions)-1 {
		m.suggestionIndex++
	}
}

// SelectSuggestion selects the current suggestion
func (m *AddProjectModal) SelectSuggestion() bool {
	if m.suggestionIndex >= 0 && m.suggestionIndex < len(m.suggestions) {
		m.pathInput.SetValue(m.suggestions[m.suggestionIndex])
		m.pathInput.CursorEnd()
		m.UpdateSuggestions()
		return true
	}
	return false
}

// HasSuggestionSelected returns true if a suggestion is selected
func (m *AddProjectModal) HasSuggestionSelected() bool {
	return m.suggestionIndex >= 0 && m.suggestionIndex < len(m.suggestions)
}

// SetSize sets the modal dimensions
func (m *AddProjectModal) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.pathInput.Width = width - 10
}

// State returns the current state
func (m *AddProjectModal) State() AddProjectState {
	return m.state
}

// Reset resets the modal to initial state
func (m *AddProjectModal) Reset() {
	m.state = AddProjectStateInput
	m.pathInput.SetValue("")
	m.pathInput.Focus()
	m.pathCompleter.Reset()
	m.suggestions = nil
	m.suggestionIndex = -1
	m.detected = nil
	m.selected = make(map[int]bool)
	m.cursor = 0
	m.error = ""
}

// PathInput returns the path input model
func (m *AddProjectModal) PathInput() *textinput.Model {
	return &m.pathInput
}

// Scan scans the entered path for services
func (m *AddProjectModal) Scan() error {
	path := m.pathInput.Value()
	if path == "" {
		m.error = "Please enter a path"
		m.state = AddProjectStateError
		return fmt.Errorf("empty path")
	}

	m.state = AddProjectStateScanning

	detector := discovery.NewDetector()
	detected, err := detector.Detect(path)
	if err != nil {
		m.error = fmt.Sprintf("Failed to scan: %v", err)
		m.state = AddProjectStateError
		return err
	}

	if len(detected.Services) == 0 {
		m.error = "No services found in this directory"
		m.state = AddProjectStateError
		return fmt.Errorf("no services found")
	}

	m.detected = detected
	m.state = AddProjectStatePreview

	// Select all services by default
	for i := range detected.Services {
		m.selected[i] = true
	}

	return nil
}

// MoveUp moves cursor up in service list
func (m *AddProjectModal) MoveUp() {
	if m.detected != nil && m.cursor > 0 {
		m.cursor--
	}
}

// MoveDown moves cursor down in service list
func (m *AddProjectModal) MoveDown() {
	if m.detected != nil && m.cursor < len(m.detected.Services)-1 {
		m.cursor++
	}
}

// ToggleSelected toggles selection of current service
func (m *AddProjectModal) ToggleSelected() {
	m.selected[m.cursor] = !m.selected[m.cursor]
}

// GetSelectedServices returns the selected services
func (m *AddProjectModal) GetSelectedServices() []discovery.DetectedService {
	if m.detected == nil {
		return nil
	}

	var services []discovery.DetectedService
	for i, svc := range m.detected.Services {
		if m.selected[i] {
			services = append(services, svc)
		}
	}
	return services
}

// GetDetectedProject returns the detected project with only selected services
func (m *AddProjectModal) GetDetectedProject() *discovery.DetectedProject {
	if m.detected == nil {
		return nil
	}

	return &discovery.DetectedProject{
		Name:     m.detected.Name,
		Path:     m.detected.Path,
		Services: m.GetSelectedServices(),
	}
}

// HasSelectedServices returns true if at least one service is selected
func (m *AddProjectModal) HasSelectedServices() bool {
	for _, selected := range m.selected {
		if selected {
			return true
		}
	}
	return false
}

// SetDone sets the modal to done state
func (m *AddProjectModal) SetDone() {
	m.state = AddProjectStateDone
}

// BackToInput goes back to input state
func (m *AddProjectModal) BackToInput() {
	m.state = AddProjectStateInput
	m.pathInput.Focus()
	m.error = ""
}

// View renders the modal
func (m *AddProjectModal) View() string {
	var b strings.Builder

	title := m.styles.Title.Render("Add Project")
	b.WriteString(title)
	b.WriteString("\n")

	switch m.state {
	case AddProjectStateInput:
		b.WriteString(m.renderInput())
	case AddProjectStateScanning:
		b.WriteString(m.styles.Subtitle.Render("Scanning..."))
	case AddProjectStatePreview:
		b.WriteString(m.renderPreview())
	case AddProjectStateError:
		b.WriteString(m.renderError())
	case AddProjectStateDone:
		b.WriteString(m.styles.Subtitle.Render("Project added successfully!"))
	}

	return m.styles.Container.
		Width(m.width).
		Render(b.String())
}

// renderInput renders the path input view
func (m *AddProjectModal) renderInput() string {
	var b strings.Builder

	b.WriteString(m.styles.Label.Render("Enter project path:"))
	b.WriteString("\n")
	b.WriteString(m.pathInput.View())
	b.WriteString("\n")

	// Show suggestions
	if len(m.suggestions) > 0 {
		b.WriteString("\n")
		for i, s := range m.suggestions {
			if i == m.suggestionIndex {
				b.WriteString(m.styles.SuggestionSel.Render("→ " + s))
			} else {
				b.WriteString(m.styles.Suggestion.Render("  " + s))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Help.Render("Tab autocomplete • ↑↓ select • Enter scan • Esc cancel"))

	return b.String()
}

// renderPreview renders the service preview
func (m *AddProjectModal) renderPreview() string {
	var b strings.Builder

	b.WriteString(m.styles.Subtitle.Render(fmt.Sprintf("Found %d services in %s:", len(m.detected.Services), m.detected.Name)))
	b.WriteString("\n\n")

	for i, svc := range m.detected.Services {
		// Checkbox
		var checkbox string
		if m.selected[i] {
			checkbox = m.styles.CheckboxSel.Render("[✓]")
		} else {
			checkbox = m.styles.Checkbox.Render("[ ]")
		}

		// Service name and framework
		name := svc.Name
		if svc.Framework != discovery.FrameworkUnknown {
			name += " " + m.styles.Framework.Render(fmt.Sprintf("(%s)", svc.Framework))
		}

		// Build line
		line := fmt.Sprintf("%s %s", checkbox, name)

		// Apply selection style
		if i == m.cursor {
			line = m.styles.ServiceSel.Render(m.padRight(line, m.width-8))
		} else {
			line = m.styles.Service.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")

		// Show command for selected item
		if i == m.cursor && svc.DevCommand != "" {
			cmd := m.styles.Command.Render("  → " + svc.DevCommand)
			b.WriteString(cmd)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Help.Render("↑↓ navigate • Space toggle • Enter confirm • Esc back"))

	return b.String()
}

// renderError renders the error view
func (m *AddProjectModal) renderError() string {
	var b strings.Builder

	b.WriteString(m.styles.Error.Render("Error: " + m.error))
	b.WriteString("\n\n")
	b.WriteString(m.styles.Help.Render("Enter to try again • Esc to cancel"))

	return b.String()
}

// padRight pads a string to the specified width
func (m *AddProjectModal) padRight(str string, width int) string {
	visibleLen := lipgloss.Width(str)
	if visibleLen >= width {
		return str
	}
	return str + strings.Repeat(" ", width-visibleLen)
}
