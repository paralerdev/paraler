package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/paralerdev/paraler/internal/config"
)

// Status represents the current state of a process
type Status int

const (
	StatusStopped Status = iota
	StatusStarting
	StatusRunning
	StatusStopping
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopping:
		return "stopping"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Process wraps an exec.Cmd with additional functionality
type Process struct {
	ID     config.ServiceID
	Config config.Service
	Cwd    string

	mu           sync.RWMutex
	cmd          *exec.Cmd
	cancel       context.CancelFunc
	status       Status
	health       HealthStatus
	exitCode     int
	exitErr      error
	startedAt    time.Time
	stoppedAt    time.Time
	restartCount int

	// Output channels
	outputCh chan OutputLine
}

// OutputLine represents a line of output from the process
type OutputLine struct {
	ServiceID config.ServiceID
	Line      string
	IsStderr  bool
	Timestamp time.Time
}

// NewProcess creates a new process wrapper
func NewProcess(id config.ServiceID, cfg config.Service, cwd string, outputCh chan OutputLine) *Process {
	return &Process{
		ID:       id,
		Config:   cfg,
		Cwd:      cwd,
		status:   StatusStopped,
		outputCh: outputCh,
	}
}

// Status returns the current process status
func (p *Process) Status() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// StartedAt returns when the process was started
func (p *Process) StartedAt() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.startedAt
}

// ExitCode returns the exit code of the last run
func (p *Process) ExitCode() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.exitCode
}

// Start starts the process
func (p *Process) Start() error {
	p.mu.Lock()
	if p.status == StatusRunning || p.status == StatusStarting {
		p.mu.Unlock()
		return fmt.Errorf("process already running")
	}

	p.status = StatusStarting
	p.exitErr = nil
	p.exitCode = 0
	p.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	p.mu.Lock()
	p.cancel = cancel
	p.mu.Unlock()

	// Check if working directory exists
	if _, err := os.Stat(p.Cwd); os.IsNotExist(err) {
		p.setStatus(StatusFailed)
		p.emitSystemMessage(fmt.Sprintf("✖ Directory not found: %s", p.Cwd))
		return fmt.Errorf("working directory does not exist: %s", p.Cwd)
	}

	// Create command with shell
	cmd := exec.CommandContext(ctx, "sh", "-c", p.Config.Cmd)
	cmd.Dir = p.Cwd
	cmd.Env = append(cmd.Environ(), p.Config.Env...)

	// Set process group for killing children
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Get stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.setStatus(StatusFailed)
		p.emitSystemMessage(fmt.Sprintf("✖ Failed to start: %v", err))
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		p.setStatus(StatusFailed)
		p.emitSystemMessage(fmt.Sprintf("✖ Failed to start: %v", err))
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		p.setStatus(StatusFailed)
		p.emitSystemMessage(fmt.Sprintf("✖ Failed to start: %v", err))
		p.emitSystemMessage(fmt.Sprintf("  Command: %s", p.Config.Cmd))
		p.emitSystemMessage(fmt.Sprintf("  Directory: %s", p.Cwd))
		return fmt.Errorf("failed to start process: %w", err)
	}

	p.mu.Lock()
	p.cmd = cmd
	p.startedAt = time.Now()
	p.status = StatusRunning
	p.mu.Unlock()

	// Emit start message
	p.emitSystemMessage("▶ Service started")

	// Stream output in goroutines
	go p.streamOutput(stdout, false)
	go p.streamOutput(stderr, true)

	// Wait for process completion in background
	go p.wait()

	return nil
}

// Stop stops the process gracefully
func (p *Process) Stop() error {
	p.mu.Lock()
	if p.status != StatusRunning {
		p.mu.Unlock()
		return nil
	}
	p.status = StatusStopping
	cmd := p.cmd
	cancel := p.cancel
	p.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// Send SIGTERM to process group
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	// Wait for graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Process exited gracefully
	case <-time.After(5 * time.Second):
		// Force kill if still running
		if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		}
		<-done
	}

	if cancel != nil {
		cancel()
	}

	return nil
}

// Restart restarts the process
func (p *Process) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	// Small delay before restart
	time.Sleep(100 * time.Millisecond)
	return p.Start()
}

// wait waits for the process to complete and updates status
func (p *Process) wait() {
	p.mu.RLock()
	cmd := p.cmd
	p.mu.RUnlock()

	if cmd == nil {
		return
	}

	err := cmd.Wait()

	p.mu.Lock()
	p.stoppedAt = time.Now()
	p.exitErr = err

	var newStatus Status
	var exitCode int

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		if p.status != StatusStopping {
			newStatus = StatusFailed
		} else {
			newStatus = StatusStopped
		}
	} else {
		exitCode = 0
		newStatus = StatusStopped
	}

	p.exitCode = exitCode
	p.status = newStatus
	p.mu.Unlock()

	// Emit stop message
	if newStatus == StatusFailed {
		p.emitSystemMessage(fmt.Sprintf("✖ Service failed (exit code: %d)", exitCode))
		p.emitSystemMessage(fmt.Sprintf("  Command: %s", p.Config.Cmd))
		p.emitSystemMessage(fmt.Sprintf("  Directory: %s", p.Cwd))
	} else {
		p.emitSystemMessage("■ Service stopped")
	}
}

// streamOutput reads from a reader and sends lines to the output channel
func (p *Process) streamOutput(r io.Reader, isStderr bool) {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		select {
		case p.outputCh <- OutputLine{
			ServiceID: p.ID,
			Line:      line,
			IsStderr:  isStderr,
			Timestamp: time.Now(),
		}:
		default:
			// Drop line if channel is full
		}
	}
}

// setStatus sets the process status
func (p *Process) setStatus(s Status) {
	p.mu.Lock()
	p.status = s
	p.mu.Unlock()
}

// emitSystemMessage sends a system message to the output channel
func (p *Process) emitSystemMessage(msg string) {
	select {
	case p.outputCh <- OutputLine{
		ServiceID: p.ID,
		Line:      msg,
		IsStderr:  false,
		Timestamp: time.Now(),
	}:
	default:
		// Drop if channel full
	}
}

// IsRunning returns true if the process is currently running
func (p *Process) IsRunning() bool {
	return p.Status() == StatusRunning
}

// Health returns the current health status
func (p *Process) Health() HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.health
}

// SetHealth sets the health status
func (p *Process) SetHealth(h HealthStatus) {
	p.mu.Lock()
	p.health = h
	p.mu.Unlock()
}

// RestartCount returns how many times the process was auto-restarted
func (p *Process) RestartCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.restartCount
}

// IncrementRestartCount increments the restart counter
func (p *Process) IncrementRestartCount() {
	p.mu.Lock()
	p.restartCount++
	p.mu.Unlock()
}

// ResetRestartCount resets the restart counter
func (p *Process) ResetRestartCount() {
	p.mu.Lock()
	p.restartCount = 0
	p.mu.Unlock()
}

// Uptime returns how long the process has been running
func (p *Process) Uptime() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.status != StatusRunning {
		return 0
	}
	return time.Since(p.startedAt)
}
