package ui

import (
	"time"

	"github.com/paralerdev/paraler/internal/log"
	"github.com/paralerdev/paraler/internal/process"
	"github.com/paralerdev/paraler/internal/ui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Messages

// OutputMsg is sent when process output is received
type OutputMsg struct {
	Line process.OutputLine
}

// ProcessStatusChangedMsg is sent when a process status changes
type ProcessStatusChangedMsg struct{}

// HealthTickMsg is sent periodically to check health
type HealthTickMsg struct{}

// ProjectScannedMsg is sent when project scanning is complete
type ProjectScannedMsg struct{}

// ProjectAddedMsg is sent when a project is added
type ProjectAddedMsg struct {
	Name string
}

// ProjectAddErrorMsg is sent when adding a project fails
type ProjectAddErrorMsg struct {
	Error error
}

// ConfigReloadedMsg is sent when config is reloaded
type ConfigReloadedMsg struct{}

// ConfigReloadErrorMsg is sent when config reload fails
type ConfigReloadErrorMsg struct {
	Error error
}

// LogsExportedMsg is sent when logs are exported
type LogsExportedMsg struct {
	Path string
}

// LogsExportErrorMsg is sent when log export fails
type LogsExportErrorMsg struct {
	Error error
}

// listenForOutput returns a command that listens for process output
func (m *Model) listenForOutput() tea.Cmd {
	return func() tea.Msg {
		line, ok := <-m.manager.OutputChannel()
		if !ok {
			return nil
		}
		return OutputMsg{Line: line}
	}
}

// tickHealth returns a command for periodic health checks
func (m *Model) tickHealth() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return HealthTickMsg{}
	})
}

// Update handles all messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := m.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.calculateLayout()
		m.ready = true

	case OutputMsg:
		// Add to log buffer
		entry := log.Entry{
			ServiceID: msg.Line.ServiceID,
			Line:      msg.Line.Line,
			IsStderr:  msg.Line.IsStderr,
			Timestamp: msg.Line.Timestamp,
		}
		m.logBuffer.Add(entry)

		// Continue listening
		cmds = append(cmds, m.listenForOutput())

	case ProcessStatusChangedMsg:
		// Status changed, UI will update automatically

	case HealthTickMsg:
		// Run health checks and auto-restart
		m.manager.CheckHealth()
		m.manager.CheckAutoRestart()
		// Continue health ticks
		cmds = append(cmds, m.tickHealth())
	}

	return m, tea.Batch(cmds...)
}

// handleKeyMsg handles keyboard input
func (m *Model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	// If confirm modal is visible, handle its input
	if m.showConfirm {
		return m.handleConfirmKeys(msg)
	}

	// If add project modal is visible, handle its input
	if m.showAddProject {
		return m.handleAddProjectKeys(msg)
	}

	// If in filter mode, handle filter input
	if m.logPanel.IsFiltering() {
		return m.handleFilterInput(msg)
	}

	// If showing help, any key closes it
	if m.showHelp {
		m.showHelp = false
		m.calculateLayout()
		return nil
	}

	// Global keys
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.manager.Shutdown()
		return tea.Quit

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		m.calculateLayout()
		return nil

	case key.Matches(msg, m.keys.Tab):
		m.toggleFocus()
		return nil

	case key.Matches(msg, m.keys.StartAll):
		return m.startAll()

	case key.Matches(msg, m.keys.StopAll):
		return m.stopAll()

	case key.Matches(msg, m.keys.AddProject):
		m.ShowAddProject()
		return nil

	case key.Matches(msg, m.keys.ReloadConfig):
		return m.reloadConfig()

	case key.Matches(msg, m.keys.ExportLogs):
		return m.exportLogs()
	}

	// Panel-specific keys
	if m.focus == FocusSidebar {
		return m.handleSidebarKeys(msg)
	}
	return m.handleLogKeys(msg)
}

// handleSidebarKeys handles keys when sidebar is focused
func (m *Model) handleSidebarKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.sidebar.MoveUp()
		m.updateLogPanelService()

	case key.Matches(msg, m.keys.Down):
		m.sidebar.MoveDown()
		m.updateLogPanelService()

	case key.Matches(msg, m.keys.Start):
		return m.startSelected()

	case key.Matches(msg, m.keys.Stop):
		return m.stopSelected()

	case key.Matches(msg, m.keys.Restart):
		return m.restartSelected()

	case key.Matches(msg, m.keys.Filter):
		m.setFocus(FocusLogs)
		m.logPanel.StartFilter()
		m.calculateLayout()

	case key.Matches(msg, m.keys.ClearLogs):
		m.clearLogs()

	case key.Matches(msg, m.keys.DeleteService):
		m.ShowConfirmDeleteService()

	case key.Matches(msg, m.keys.DeleteProject):
		m.ShowConfirmDeleteProject()

	case key.Matches(msg, m.keys.ToggleSelect):
		m.sidebar.ToggleMultiSelect()

	case key.Matches(msg, m.keys.ClearSelect):
		m.sidebar.ClearMultiSelect()
	}

	return nil
}

// handleLogKeys handles keys when log panel is focused
func (m *Model) handleLogKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.logPanel.ScrollUp()

	case key.Matches(msg, m.keys.Down):
		m.logPanel.ScrollDown()

	case key.Matches(msg, m.keys.PageUp):
		m.logPanel.PageUp()

	case key.Matches(msg, m.keys.PageDown):
		m.logPanel.PageDown()

	case key.Matches(msg, m.keys.Home):
		m.logPanel.GoToTop()

	case key.Matches(msg, m.keys.End):
		m.logPanel.GoToBottom()

	case key.Matches(msg, m.keys.Filter):
		m.logPanel.StartFilter()
		m.calculateLayout()

	case key.Matches(msg, m.keys.ClearLogs):
		m.clearLogs()

	case key.Matches(msg, m.keys.Start):
		return m.startSelected()

	case key.Matches(msg, m.keys.Stop):
		return m.stopSelected()

	case key.Matches(msg, m.keys.Restart):
		return m.restartSelected()
	}

	return nil
}

// handleFilterInput handles input when filtering
func (m *Model) handleFilterInput(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Enter):
		m.logPanel.ApplyFilter()
		m.calculateLayout()
		return nil

	case key.Matches(msg, m.keys.Escape):
		m.logPanel.ClearFilter()
		m.calculateLayout()
		return nil
	}

	// Pass to text input
	input := m.logPanel.FilterInput()
	newInput, cmd := input.Update(msg)
	*input = newInput
	return cmd
}

// handleAddProjectKeys handles keys when add project modal is visible
func (m *Model) handleAddProjectKeys(msg tea.KeyMsg) tea.Cmd {
	modal := m.addProjectModal

	switch modal.State() {
	case components.AddProjectStateInput:
		switch {
		case key.Matches(msg, m.keys.Escape):
			m.HideAddProject()
			return nil

		case key.Matches(msg, m.keys.Enter):
			// If suggestion is selected, use it; otherwise scan
			if modal.HasSuggestionSelected() {
				modal.SelectSuggestion()
				return nil
			}
			return m.scanProject()

		case key.Matches(msg, m.keys.Tab):
			modal.CompleteTab()
			return nil

		case key.Matches(msg, m.keys.Up):
			modal.SuggestionUp()
			return nil

		case key.Matches(msg, m.keys.Down):
			modal.SuggestionDown()
			return nil
		}

		// Pass to text input
		input := modal.PathInput()
		newInput, cmd := input.Update(msg)
		*input = newInput

		// Update suggestions on input change
		modal.UpdateSuggestions()

		return cmd

	case components.AddProjectStatePreview:
		switch {
		case key.Matches(msg, m.keys.Escape):
			modal.BackToInput()
			return nil

		case key.Matches(msg, m.keys.Enter):
			return m.confirmAddProject()

		case key.Matches(msg, m.keys.Up):
			modal.MoveUp()

		case key.Matches(msg, m.keys.Down):
			modal.MoveDown()

		case key.Matches(msg, m.keys.Space):
			modal.ToggleSelected()
		}

	case components.AddProjectStateError:
		switch {
		case key.Matches(msg, m.keys.Escape):
			m.HideAddProject()
			return nil

		case key.Matches(msg, m.keys.Enter):
			modal.BackToInput()
		}

	case components.AddProjectStateDone:
		m.HideAddProject()
	}

	return nil
}

// scanProject scans the entered path for services
func (m *Model) scanProject() tea.Cmd {
	return func() tea.Msg {
		m.addProjectModal.Scan()
		return ProjectScannedMsg{}
	}
}

// confirmAddProject adds the project to config and saves
func (m *Model) confirmAddProject() tea.Cmd {
	return func() tea.Msg {
		modal := m.addProjectModal

		if !modal.HasSelectedServices() {
			return nil
		}

		detected := modal.GetDetectedProject()
		if detected == nil {
			return nil
		}

		// Add to config
		detected.MergeIntoConfig(m.config)

		// Save config
		if err := m.config.Save(m.configPath); err != nil {
			// Handle error - for now just log
			return ProjectAddErrorMsg{Error: err}
		}

		// Reload UI
		m.ReloadConfig()
		modal.SetDone()

		return ProjectAddedMsg{Name: detected.Name}
	}
}

// handleConfirmKeys handles keys when confirm modal is visible
func (m *Model) handleConfirmKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		// Execute the confirmed action
		modal := m.confirmModal
		action := modal.Action()
		projectName := modal.ProjectName()
		targetName := modal.TargetName()

		m.HideConfirm()

		switch action {
		case components.ConfirmDeleteService:
			return func() tea.Msg {
				m.DeleteService(projectName, targetName)
				return ServiceDeletedMsg{Project: projectName, Service: targetName}
			}
		case components.ConfirmDeleteProject:
			return func() tea.Msg {
				m.DeleteProject(projectName)
				return ProjectDeletedMsg{Name: projectName}
			}
		}

	case key.Matches(msg, m.keys.Escape):
		m.HideConfirm()

	case msg.String() == "n":
		m.HideConfirm()
	}

	return nil
}

// ServiceDeletedMsg is sent when a service is deleted
type ServiceDeletedMsg struct {
	Project string
	Service string
}

// ProjectDeletedMsg is sent when a project is deleted
type ProjectDeletedMsg struct {
	Name string
}

// reloadConfig reloads the config file
func (m *Model) reloadConfig() tea.Cmd {
	return func() tea.Msg {
		if err := m.HotReload(); err != nil {
			return ConfigReloadErrorMsg{Error: err}
		}
		return ConfigReloadedMsg{}
	}
}

// exportLogs exports logs for the selected service
func (m *Model) exportLogs() tea.Cmd {
	return func() tea.Msg {
		path, err := m.ExportLogs()
		if err != nil {
			return LogsExportErrorMsg{Error: err}
		}
		return LogsExportedMsg{Path: path}
	}
}
