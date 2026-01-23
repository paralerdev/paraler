package process

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PortStatus represents information about a port
type PortStatus struct {
	Port      int
	InUse     bool
	PID       int
	Process   string // Process name
	Command   string // Full command (if available)
}

// GetPortStatus checks if a port is in use and returns info about the process
func GetPortStatus(port int) PortStatus {
	status := PortStatus{Port: port}

	// Quick check if port is available
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		// Port is free
		status.InUse = false
		return status
	}
	conn.Close()
	status.InUse = true

	// Try to find what's using the port using lsof
	status.PID, status.Process, status.Command = getProcessOnPort(port)

	return status
}

// getProcessOnPort uses lsof to find process using a port (macOS/Linux)
func getProcessOnPort(port int) (pid int, name string, command string) {
	// lsof -i :PORT -t gives PID
	// lsof -i :PORT gives full info
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		return 0, "", ""
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, "", ""
	}

	// Parse lsof output (skip header)
	// COMMAND  PID  USER   FD   TYPE  DEVICE  SIZE/OFF  NODE  NAME
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		// Only look for LISTEN state
		if !strings.Contains(line, "LISTEN") {
			continue
		}

		fields := regexp.MustCompile(`\s+`).Split(line, -1)
		if len(fields) < 2 {
			continue
		}

		name = fields[0]
		if p, err := strconv.Atoi(fields[1]); err == nil {
			pid = p
		}

		// Get full command line
		if pid > 0 {
			command = getCommandLine(pid)
		}

		return pid, name, command
	}

	return 0, "", ""
}

// getCommandLine gets the full command line for a process
func getCommandLine(pid int) string {
	// ps -p PID -o args=
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "args=")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// KillProcessOnPort kills the process using a specific port
func KillProcessOnPort(port int) error {
	status := GetPortStatus(port)
	if !status.InUse {
		return nil // Port is free
	}

	if status.PID == 0 {
		return fmt.Errorf("could not find process on port %d", port)
	}

	// Kill the process
	cmd := exec.Command("kill", "-9", strconv.Itoa(status.PID))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill process %d: %w", status.PID, err)
	}

	// Wait a bit for port to be released
	time.Sleep(100 * time.Millisecond)

	return nil
}
