package process

import (
	"fmt"
	"sync"
	"time"

	"github.com/paralerdev/paraler/internal/config"
)

const maxAutoRestarts = 5 // Maximum auto-restarts before giving up

// Manager handles multiple processes
type Manager struct {
	mu            sync.RWMutex
	processes     map[string]*Process // key: ServiceID.String()
	outputCh      chan OutputLine
	healthChecker *HealthChecker
	config        *config.Config
}

// NewManager creates a new process manager
func NewManager(cfg *config.Config) *Manager {
	outputCh := make(chan OutputLine, 1000)
	m := &Manager{
		processes:     make(map[string]*Process),
		outputCh:      outputCh,
		healthChecker: NewHealthChecker(),
		config:        cfg,
	}

	// Create processes for all services
	for projectName, project := range cfg.Projects {
		for serviceName, service := range project.Services {
			id := config.ServiceID{
				Project: projectName,
				Service: serviceName,
			}
			cwd := cfg.GetServiceCwd(projectName, serviceName)
			proc := NewProcess(id, service, cwd, outputCh)
			m.processes[id.String()] = proc
		}
	}

	return m
}

// OutputChannel returns the channel for receiving process output
func (m *Manager) OutputChannel() <-chan OutputLine {
	return m.outputCh
}

// Get returns a process by its ID
func (m *Manager) Get(id config.ServiceID) *Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processes[id.String()]
}

// All returns all processes
func (m *Manager) All() []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()

	procs := make([]*Process, 0, len(m.processes))
	for _, p := range m.processes {
		procs = append(procs, p)
	}
	return procs
}

// Start starts a specific service (with dependencies)
func (m *Manager) Start(id config.ServiceID) error {
	proc := m.Get(id)
	if proc == nil {
		return nil
	}

	// Check for port conflicts with running services
	if hasConflict, conflictID := m.CheckPortConflict(id); hasConflict {
		// Send warning to output channel
		m.sendWarning(id, fmt.Sprintf("Port %d is already in use by %s", proc.Config.Port, conflictID.String()))
	}

	// Start dependencies first
	for _, dep := range proc.Config.DependsOn {
		depID := config.ServiceID{Project: id.Project, Service: dep}
		depProc := m.Get(depID)
		if depProc != nil && depProc.Status() != StatusRunning {
			if err := depProc.Start(); err != nil {
				return err
			}
			// Wait for dependency to be ready
			m.waitForReady(depID, 10*time.Second)
		}
	}

	return proc.Start()
}

// sendWarning sends a warning message to the output channel
func (m *Manager) sendWarning(id config.ServiceID, message string) {
	select {
	case m.outputCh <- OutputLine{
		ServiceID: id,
		Line:      fmt.Sprintf("[WARNING] %s", message),
		IsStderr:  true,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, drop warning
	}
}

// waitForReady waits for a service to be ready (running and healthy)
func (m *Manager) waitForReady(id config.ServiceID, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		proc := m.Get(id)
		if proc == nil {
			return
		}
		if proc.Status() == StatusRunning {
			// Check health if configured
			health := m.healthChecker.CheckHealth(proc.Config)
			if health == HealthHealthy || health == HealthUnknown {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// Stop stops a specific service
func (m *Manager) Stop(id config.ServiceID) error {
	proc := m.Get(id)
	if proc == nil {
		return nil
	}
	return proc.Stop()
}

// Restart restarts a specific service
func (m *Manager) Restart(id config.ServiceID) error {
	proc := m.Get(id)
	if proc == nil {
		return nil
	}
	return proc.Restart()
}

// StartAll starts all services in dependency order
func (m *Manager) StartAll() {
	// Get services sorted by dependencies
	order := m.getDependencyOrder()

	for _, id := range order {
		proc := m.Get(id)
		if proc != nil && proc.Status() != StatusRunning {
			proc.Start()
			// Small delay between starts
			if proc.Config.Delay > 0 {
				time.Sleep(proc.Config.Delay)
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// getDependencyOrder returns services sorted by dependencies (topological sort)
func (m *Manager) getDependencyOrder() []config.ServiceID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Build dependency graph
	var allIDs []config.ServiceID
	deps := make(map[string][]string) // service -> dependencies

	for _, proc := range m.processes {
		allIDs = append(allIDs, proc.ID)
		key := proc.ID.String()
		deps[key] = nil
		for _, dep := range proc.Config.DependsOn {
			depID := config.ServiceID{Project: proc.ID.Project, Service: dep}
			deps[key] = append(deps[key], depID.String())
		}
	}

	// Topological sort using Kahn's algorithm
	inDegree := make(map[string]int)
	for id := range deps {
		inDegree[id] = 0
	}
	for _, dependencies := range deps {
		for _, dep := range dependencies {
			inDegree[dep]++
		}
	}

	// Find all nodes with no incoming edges
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var result []config.ServiceID
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		// Find the ServiceID for this key
		for _, proc := range m.processes {
			if proc.ID.String() == current {
				result = append(result, proc.ID)
				break
			}
		}

		// Reduce in-degree for dependents
		for id, dependencies := range deps {
			for _, dep := range dependencies {
				if dep == current {
					inDegree[id]--
					if inDegree[id] == 0 {
						queue = append(queue, id)
					}
				}
			}
		}
	}

	// If result doesn't contain all services, there's a cycle - just return all
	if len(result) != len(allIDs) {
		return allIDs
	}

	return result
}

// StopAll stops all services
func (m *Manager) StopAll() {
	m.mu.RLock()
	procs := make([]*Process, 0, len(m.processes))
	for _, p := range m.processes {
		procs = append(procs, p)
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, p := range procs {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			proc.Stop()
		}(p)
	}
	wg.Wait()
}

// RestartAll restarts all services
func (m *Manager) RestartAll() {
	m.StopAll()
	m.StartAll()
}

// Shutdown gracefully shuts down all processes
func (m *Manager) Shutdown() {
	m.StopAll()
	close(m.outputCh)
}

// GetByProject returns all processes for a specific project
func (m *Manager) GetByProject(projectName string) []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var procs []*Process
	for _, p := range m.processes {
		if p.ID.Project == projectName {
			procs = append(procs, p)
		}
	}
	return procs
}

// StartProject starts all services in a project
func (m *Manager) StartProject(projectName string) {
	procs := m.GetByProject(projectName)
	var wg sync.WaitGroup
	for _, p := range procs {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			proc.Start()
		}(p)
	}
	wg.Wait()
}

// StopProject stops all services in a project
func (m *Manager) StopProject(projectName string) {
	procs := m.GetByProject(projectName)
	var wg sync.WaitGroup
	for _, p := range procs {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			proc.Stop()
		}(p)
	}
	wg.Wait()
}

// RunningCount returns the number of running processes
func (m *Manager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, p := range m.processes {
		if p.IsRunning() {
			count++
		}
	}
	return count
}

// TotalCount returns the total number of processes
func (m *Manager) TotalCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.processes)
}

// CheckHealth performs health checks on all running processes
func (m *Manager) CheckHealth() {
	m.mu.RLock()
	procs := make([]*Process, 0, len(m.processes))
	for _, p := range m.processes {
		procs = append(procs, p)
	}
	m.mu.RUnlock()

	for _, p := range procs {
		if p.Status() == StatusRunning {
			health := m.healthChecker.CheckHealth(p.Config)
			p.SetHealth(health)
		} else {
			p.SetHealth(HealthUnknown)
		}
	}
}

// CheckAutoRestart checks for failed processes and restarts them if auto_restart is enabled
func (m *Manager) CheckAutoRestart() {
	m.mu.RLock()
	procs := make([]*Process, 0, len(m.processes))
	for _, p := range m.processes {
		procs = append(procs, p)
	}
	m.mu.RUnlock()

	for _, p := range procs {
		if p.Status() == StatusFailed && p.Config.AutoRestart {
			if p.RestartCount() < maxAutoRestarts {
				p.IncrementRestartCount()
				// Small delay before restart
				time.Sleep(500 * time.Millisecond)
				p.Start()
			}
		}
	}
}

// GetHealth returns the health status of a specific service
func (m *Manager) GetHealth(id config.ServiceID) HealthStatus {
	proc := m.Get(id)
	if proc == nil {
		return HealthUnknown
	}
	return proc.Health()
}

// Config returns the manager's config
func (m *Manager) Config() *config.Config {
	return m.config
}

// GetPortConflicts returns a map of ports that have conflicts
// Returns map[port] -> []ServiceID that use this port
func (m *Manager) GetPortConflicts() map[int][]config.ServiceID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	portUsage := make(map[int][]config.ServiceID)

	for _, proc := range m.processes {
		if proc.Config.Port > 0 {
			portUsage[proc.Config.Port] = append(portUsage[proc.Config.Port], proc.ID)
		}
	}

	// Filter to only conflicting ports
	conflicts := make(map[int][]config.ServiceID)
	for port, ids := range portUsage {
		if len(ids) > 1 {
			conflicts[port] = ids
		}
	}

	return conflicts
}

// CheckPortConflict checks if starting this service would conflict with another running service
func (m *Manager) CheckPortConflict(id config.ServiceID) (bool, config.ServiceID) {
	proc := m.Get(id)
	if proc == nil || proc.Config.Port == 0 {
		return false, config.ServiceID{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, other := range m.processes {
		if other.ID == id {
			continue
		}
		if other.Config.Port == proc.Config.Port && other.Status() == StatusRunning {
			return true, other.ID
		}
	}

	return false, config.ServiceID{}
}

// GetRunningPorts returns a map of ports used by running services
func (m *Manager) GetRunningPorts() map[int]config.ServiceID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ports := make(map[int]config.ServiceID)
	for _, proc := range m.processes {
		if proc.Config.Port > 0 && proc.Status() == StatusRunning {
			ports[proc.Config.Port] = proc.ID
		}
	}
	return ports
}
