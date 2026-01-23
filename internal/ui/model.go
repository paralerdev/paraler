package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/paralerdev/paraler/internal/config"
	"github.com/paralerdev/paraler/internal/log"
	"github.com/paralerdev/paraler/internal/process"
	"github.com/paralerdev/paraler/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// Focus represents which panel is focused
type Focus int

const (
	FocusSidebar Focus = iota
	FocusLogs
	FocusAddProject
)

// Model is the root Bubble Tea model
type Model struct {
	// Config
	config     *config.Config
	configPath string

	// Process management
	manager *process.Manager

	// Log buffer
	logBuffer *log.Buffer

	// UI components
	sidebar          *components.Sidebar
	logPanel         *components.LogPanel
	statusBar        *components.StatusBar
	addProjectModal  *components.AddProjectModal
	confirmModal     *components.ConfirmModal
	moveServiceModal *components.MoveServiceModal
	renameModal      *components.RenameModal

	// UI state
	focus            Focus
	showHelp         bool
	showAddProject   bool
	showConfirm      bool
	showMoveService  bool
	showRename       bool
	fullscreen       bool
	width            int
	height           int
	ready            bool

	// Key bindings
	keys KeyMap
}

// NewModel creates a new root model
func NewModel(cfg *config.Config, configPath string) *Model {
	manager := process.NewManager(cfg)

	m := &Model{
		config:           cfg,
		configPath:       configPath,
		manager:          manager,
		logBuffer:        log.NewBuffer(1000),
		sidebar:          components.NewSidebar(cfg),
		logPanel:         components.NewLogPanel(),
		statusBar:        components.NewStatusBar(),
		addProjectModal:  components.NewAddProjectModal(),
		confirmModal:     components.NewConfirmModal(),
		moveServiceModal: components.NewMoveServiceModal(),
		renameModal:      components.NewRenameModal(),
		focus:            FocusSidebar,
		keys:             DefaultKeyMap(),
	}

	// Select first service if available
	if m.sidebar.ServiceCount() > 0 {
		m.sidebar.SelectFirst()
		m.updateLogPanelService()
	}

	return m
}

// Config returns the current config
func (m *Model) Config() *config.Config {
	return m.config
}

// ConfigPath returns the config file path
func (m *Model) ConfigPath() string {
	return m.configPath
}

// ReloadConfig reloads the configuration and rebuilds the UI
func (m *Model) ReloadConfig() {
	// Stop all processes
	m.manager.StopAll()

	// Reload manager
	m.manager = process.NewManager(m.config)

	// Rebuild sidebar
	m.sidebar = components.NewSidebar(m.config)

	// Recalculate layout
	m.calculateLayout()

	// Select first service if available
	if m.sidebar.ServiceCount() > 0 {
		m.sidebar.SelectFirst()
		m.updateLogPanelService()
	}
}

// ShowAddProject shows the add project modal
func (m *Model) ShowAddProject() {
	m.showAddProject = true
	m.focus = FocusAddProject
	m.addProjectModal.Reset()
	m.addProjectModal.SetSize(m.width/2, m.height/2)
}

// HideAddProject hides the add project modal
func (m *Model) HideAddProject() {
	m.showAddProject = false
	m.focus = FocusSidebar
}

// AddProjectModal returns the add project modal
func (m *Model) AddProjectModal() *components.AddProjectModal {
	return m.addProjectModal
}

// IsAddProjectVisible returns true if add project modal is visible
func (m *Model) IsAddProjectVisible() bool {
	return m.showAddProject
}

// ShowConfirmDeleteService shows confirmation for deleting a service
func (m *Model) ShowConfirmDeleteService() {
	selected := m.sidebar.Selected()
	if selected.Service == "" {
		return
	}
	m.confirmModal.Show(components.ConfirmDeleteService, selected.Project, selected.Service)
	m.confirmModal.SetSize(m.width / 2)
	m.showConfirm = true
}

// ShowConfirmDeleteProject shows confirmation for deleting a project
func (m *Model) ShowConfirmDeleteProject() {
	selected := m.sidebar.Selected()
	if selected.Project == "" {
		return
	}
	m.confirmModal.Show(components.ConfirmDeleteProject, selected.Project, "")
	m.confirmModal.SetSize(m.width / 2)
	m.showConfirm = true
}

// HideConfirm hides the confirmation modal
func (m *Model) HideConfirm() {
	m.confirmModal.Hide()
	m.showConfirm = false
}

// ConfirmModal returns the confirm modal
func (m *Model) ConfirmModal() *components.ConfirmModal {
	return m.confirmModal
}

// IsConfirmVisible returns true if confirm modal is visible
func (m *Model) IsConfirmVisible() bool {
	return m.showConfirm
}

// DeleteService removes a service from config
func (m *Model) DeleteService(projectName, serviceName string) error {
	project, ok := m.config.Projects[projectName]
	if !ok {
		return nil
	}

	// Stop the service if running
	id := config.ServiceID{Project: projectName, Service: serviceName}
	m.manager.Stop(id)

	// Remove from config
	delete(project.Services, serviceName)
	m.config.Projects[projectName] = project

	// Save config
	if err := m.config.Save(m.configPath); err != nil {
		return err
	}

	// Reload UI
	m.ReloadConfig()
	return nil
}

// DeleteProject removes a project from config
func (m *Model) DeleteProject(projectName string) error {
	// Stop all services in the project
	m.manager.StopProject(projectName)

	// Remove from config
	m.config.RemoveProject(projectName)

	// Save config
	if err := m.config.Save(m.configPath); err != nil {
		return err
	}

	// Reload UI
	m.ReloadConfig()
	return nil
}

// ShowMoveService shows the move service modal
func (m *Model) ShowMoveService() {
	selected := m.sidebar.Selected()
	if selected.Service == "" {
		return
	}

	// Get all project names
	projectNames := m.config.ProjectNames()
	if len(projectNames) < 2 {
		// Need at least 2 projects to move
		return
	}

	m.moveServiceModal.Show(selected.Service, selected.Project, projectNames)
	m.moveServiceModal.SetSize(m.width / 2)
	m.showMoveService = true
}

// HideMoveService hides the move service modal
func (m *Model) HideMoveService() {
	m.moveServiceModal.Hide()
	m.showMoveService = false
}

// MoveServiceModal returns the move service modal
func (m *Model) MoveServiceModal() *components.MoveServiceModal {
	return m.moveServiceModal
}

// IsMoveServiceVisible returns true if move service modal is visible
func (m *Model) IsMoveServiceVisible() bool {
	return m.showMoveService
}

// MoveService moves a service to another project
func (m *Model) MoveService(serviceName, fromProject, toProject string) error {
	// Stop the service if running
	id := config.ServiceID{Project: fromProject, Service: serviceName}
	m.manager.Stop(id)

	// Move in config
	if err := m.config.MoveService(serviceName, fromProject, toProject); err != nil {
		return err
	}

	// Save config
	if err := m.config.Save(m.configPath); err != nil {
		return err
	}

	// Reload UI
	m.ReloadConfig()
	return nil
}

// ShowRename shows the rename modal for the project of the selected service
func (m *Model) ShowRename() {
	selected := m.sidebar.Selected()
	if selected.Project == "" {
		return
	}
	m.renameModal.ShowRenameProject(selected.Project)
	m.renameModal.SetSize(m.width / 2)
	m.showRename = true
}

// HideRename hides the rename modal
func (m *Model) HideRename() {
	m.renameModal.Hide()
	m.showRename = false
}

// RenameModal returns the rename modal
func (m *Model) RenameModal() *components.RenameModal {
	return m.renameModal
}

// IsRenameVisible returns true if rename modal is visible
func (m *Model) IsRenameVisible() bool {
	return m.showRename
}

// RenameProject renames a project
func (m *Model) RenameProject(oldName, newName string) error {
	// Stop all services in the project
	m.manager.StopProject(oldName)

	// Rename in config
	if err := m.config.RenameProject(oldName, newName); err != nil {
		return err
	}

	// Save config
	if err := m.config.Save(m.configPath); err != nil {
		return err
	}

	// Reload UI
	m.ReloadConfig()
	return nil
}

// RenameService renames a service
func (m *Model) RenameService(projectName, oldName, newName string) error {
	// Stop the service if running
	id := config.ServiceID{Project: projectName, Service: oldName}
	m.manager.Stop(id)

	// Rename in config
	if err := m.config.RenameService(projectName, oldName, newName); err != nil {
		return err
	}

	// Save config
	if err := m.config.Save(m.configPath); err != nil {
		return err
	}

	// Reload UI
	m.ReloadConfig()
	return nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.listenForOutput(),
		m.tickHealth(),
	)
}

// Manager returns the process manager
func (m *Model) Manager() *process.Manager {
	return m.manager
}

// updateLogPanelService updates the log panel to show the selected service
func (m *Model) updateLogPanelService() {
	selected := m.sidebar.Selected()
	m.logPanel.SetService(selected)

	// Set service config for footer
	if selected.Service != "" {
		if project, ok := m.config.Projects[selected.Project]; ok {
			if service, ok := project.Services[selected.Service]; ok {
				m.logPanel.SetServiceConfig(&service)
				return
			}
		}
	}
	m.logPanel.SetServiceConfig(nil)
}

// updateLogPanelStatus updates the log panel with current service status
func (m *Model) updateLogPanelStatus() {
	selected := m.sidebar.Selected()
	if selected.Service == "" {
		m.logPanel.SetStatus(process.StatusStopped)
		return
	}

	proc := m.manager.Get(selected)
	if proc != nil {
		m.logPanel.SetStatus(proc.Status())
	} else {
		m.logPanel.SetStatus(process.StatusStopped)
	}
}

// setFocus sets the focus to a specific panel
func (m *Model) setFocus(focus Focus) {
	m.focus = focus
	m.sidebar.SetFocused(focus == FocusSidebar)
	m.logPanel.SetFocused(focus == FocusLogs)
}

// toggleFocus switches focus between panels
func (m *Model) toggleFocus() {
	if m.fullscreen {
		// In fullscreen, always focus on logs
		return
	}
	if m.focus == FocusSidebar {
		m.setFocus(FocusLogs)
	} else {
		m.setFocus(FocusSidebar)
	}
}

// toggleFullscreen toggles fullscreen mode for logs
func (m *Model) toggleFullscreen() {
	m.fullscreen = !m.fullscreen
	if m.fullscreen {
		// In fullscreen, always focus on logs
		m.setFocus(FocusLogs)
	} else {
		// Exit fullscreen, return focus to sidebar
		m.setFocus(FocusSidebar)
	}
	m.calculateLayout()
}

// IsFullscreen returns true if in fullscreen mode
func (m *Model) IsFullscreen() bool {
	return m.fullscreen
}

// startSelected starts the selected service(s)
func (m *Model) startSelected() tea.Cmd {
	// Check for multi-select
	if m.sidebar.HasMultiSelect() {
		ids := m.sidebar.GetMultiSelected()
		return func() tea.Msg {
			for _, id := range ids {
				m.manager.Start(id)
			}
			m.sidebar.ClearMultiSelect()
			return ProcessStatusChangedMsg{}
		}
	}

	selected := m.sidebar.Selected()
	if selected.Service == "" {
		return nil
	}
	return func() tea.Msg {
		m.manager.Start(selected)
		return ProcessStatusChangedMsg{}
	}
}

// stopSelected stops the selected service(s)
func (m *Model) stopSelected() tea.Cmd {
	// Check for multi-select
	if m.sidebar.HasMultiSelect() {
		ids := m.sidebar.GetMultiSelected()
		return func() tea.Msg {
			for _, id := range ids {
				m.manager.Stop(id)
			}
			m.sidebar.ClearMultiSelect()
			return ProcessStatusChangedMsg{}
		}
	}

	selected := m.sidebar.Selected()
	if selected.Service == "" {
		return nil
	}
	return func() tea.Msg {
		m.manager.Stop(selected)
		return ProcessStatusChangedMsg{}
	}
}

// restartSelected restarts the selected service(s)
func (m *Model) restartSelected() tea.Cmd {
	// Check for multi-select
	if m.sidebar.HasMultiSelect() {
		ids := m.sidebar.GetMultiSelected()
		return func() tea.Msg {
			for _, id := range ids {
				m.manager.Restart(id)
			}
			m.sidebar.ClearMultiSelect()
			return ProcessStatusChangedMsg{}
		}
	}

	selected := m.sidebar.Selected()
	if selected.Service == "" {
		return nil
	}
	return func() tea.Msg {
		m.manager.Restart(selected)
		return ProcessStatusChangedMsg{}
	}
}

// startAll starts all services
func (m *Model) startAll() tea.Cmd {
	return func() tea.Msg {
		m.manager.StartAll()
		return ProcessStatusChangedMsg{}
	}
}

// stopAll stops all services
func (m *Model) stopAll() tea.Cmd {
	return func() tea.Msg {
		m.manager.StopAll()
		return ProcessStatusChangedMsg{}
	}
}

// clearLogs clears logs for the selected service
func (m *Model) clearLogs() {
	selected := m.sidebar.Selected()
	if selected.Service != "" {
		m.logBuffer.Clear(selected)
	}
}

// calculateLayout calculates panel sizes based on terminal dimensions
func (m *Model) calculateLayout() {
	// Status bar height
	statusHeight := 1
	if m.showHelp {
		statusHeight = 10
	}

	// Panel heights (subtract status bar)
	panelHeight := m.height - statusHeight - 1

	if m.fullscreen {
		// Fullscreen mode: logs take full width
		m.logPanel.SetSize(m.width, panelHeight)
	} else {
		// Normal mode: sidebar + logs
		// Sidebar takes ~25% width, min 20, max 40
		sidebarWidth := m.width / 4
		if sidebarWidth < 20 {
			sidebarWidth = 20
		}
		if sidebarWidth > 40 {
			sidebarWidth = 40
		}

		// Log panel takes remaining width
		logWidth := m.width - sidebarWidth - 1

		m.sidebar.SetSize(sidebarWidth, panelHeight)
		m.logPanel.SetSize(logWidth, panelHeight)
	}

	m.statusBar.SetWidth(m.width)
}

// HotReload reloads the config file and updates the UI
func (m *Model) HotReload() error {
	// Load new config
	newConfig, err := config.Load(m.configPath)
	if err != nil {
		return err
	}

	// Stop all running processes
	m.manager.StopAll()

	// Update config
	m.config = newConfig

	// Recreate manager with new config
	m.manager = process.NewManager(m.config)

	// Rebuild sidebar
	m.sidebar = components.NewSidebar(m.config)

	// Recalculate layout
	m.calculateLayout()

	// Select first service if available
	if m.sidebar.ServiceCount() > 0 {
		m.sidebar.SelectFirst()
		m.updateLogPanelService()
	}

	return nil
}

// ExportLogs exports logs for the selected service to a file
func (m *Model) ExportLogs() (string, error) {
	selected := m.sidebar.Selected()
	if selected.Service == "" {
		return "", fmt.Errorf("no service selected")
	}

	// Get logs for service
	entries := m.logBuffer.Get(selected)
	if len(entries) == 0 {
		return "", fmt.Errorf("no logs to export")
	}

	// Create logs directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	logsDir := filepath.Join(homeDir, "paraler-logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return "", err
	}

	// Generate filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s_%s.log", selected.Project, selected.Service, timestamp)
	filepath := filepath.Join(logsDir, filename)

	// Write logs
	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	for _, entry := range entries {
		line := fmt.Sprintf("[%s] %s\n", entry.Timestamp.Format("15:04:05"), entry.Line)
		file.WriteString(line)
	}

	return filepath, nil
}
