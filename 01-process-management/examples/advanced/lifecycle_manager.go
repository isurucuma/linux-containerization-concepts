package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ProcessLifecycleManager manages the complete lifecycle of processes
type ProcessLifecycleManager struct {
	processes       map[int]*ManagedProcess
	mutex           sync.RWMutex
	nextProcessID   int
	shutdownChannel chan bool
}

// ManagedProcess represents a process under management
type ManagedProcess struct {
	ID           int
	PID          int
	Name         string
	Command      []string
	State        ProcessState
	StartTime    time.Time
	EndTime      *time.Time
	RestartCount int
	MaxRestarts  int
	HealthCheck  HealthChecker
	Dependencies []int
	Environment  map[string]string
	WorkDir      string
	LogFile      string
	cmd          *exec.Cmd
}

// ProcessState represents the state of a managed process
type ProcessState int

const (
	StateCreated ProcessState = iota
	StateStarting
	StateRunning
	StateStopping
	StateStopped
	StateFailed
	StateRestarting
)

func (ps ProcessState) String() string {
	states := []string{"Created", "Starting", "Running", "Stopping", "Stopped", "Failed", "Restarting"}
	if int(ps) < len(states) {
		return states[ps]
	}
	return "Unknown"
}

// HealthChecker interface for process health checking
type HealthChecker interface {
	Check(process *ManagedProcess) bool
}

// DefaultHealthChecker checks if process is still running
type DefaultHealthChecker struct{}

func (dhc *DefaultHealthChecker) Check(process *ManagedProcess) bool {
	if process.cmd == nil || process.cmd.Process == nil {
		return false
	}

	// Send signal 0 to check if process exists
	err := process.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

// HTTPHealthChecker checks HTTP endpoint health
type HTTPHealthChecker struct {
	URL      string
	Timeout  time.Duration
	Expected int
}

func (hhc *HTTPHealthChecker) Check(process *ManagedProcess) bool {
	// Simplified HTTP health check - in real implementation would use HTTP client
	return process.cmd != nil && process.cmd.Process != nil
}

// NewProcessLifecycleManager creates a new process lifecycle manager
func NewProcessLifecycleManager() *ProcessLifecycleManager {
	return &ProcessLifecycleManager{
		processes:       make(map[int]*ManagedProcess),
		nextProcessID:   1,
		shutdownChannel: make(chan bool),
	}
}

// CreateProcess creates a new managed process
func (plm *ProcessLifecycleManager) CreateProcess(name string, command []string) *ManagedProcess {
	plm.mutex.Lock()
	defer plm.mutex.Unlock()

	process := &ManagedProcess{
		ID:          plm.nextProcessID,
		Name:        name,
		Command:     command,
		State:       StateCreated,
		StartTime:   time.Now(),
		MaxRestarts: 3,
		HealthCheck: &DefaultHealthChecker{},
		Environment: make(map[string]string),
		WorkDir:     "/tmp",
	}

	plm.processes[process.ID] = process
	plm.nextProcessID++

	fmt.Printf("Created process %d: %s\n", process.ID, process.Name)
	return process
}

// StartProcess starts a managed process
func (plm *ProcessLifecycleManager) StartProcess(processID int) error {
	plm.mutex.Lock()
	process, exists := plm.processes[processID]
	plm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("process %d not found", processID)
	}

	// Check dependencies
	if !plm.checkDependencies(process) {
		return fmt.Errorf("dependencies not satisfied for process %d", processID)
	}

	process.State = StateStarting
	fmt.Printf("Starting process %d: %s\n", process.ID, process.Name)

	// Prepare command
	cmd := exec.Command(process.Command[0], process.Command[1:]...)
	cmd.Dir = process.WorkDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range process.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Set up logging if specified
	if process.LogFile != "" {
		logFile, err := os.OpenFile(process.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		}
	}

	// Start the process
	err := cmd.Start()
	if err != nil {
		process.State = StateFailed
		return fmt.Errorf("failed to start process %d: %v", processID, err)
	}

	process.cmd = cmd
	process.PID = cmd.Process.Pid
	process.State = StateRunning
	process.StartTime = time.Now()

	fmt.Printf("Process %d started with PID %d\n", process.ID, process.PID)

	// Start monitoring goroutine
	go plm.monitorProcess(process)

	return nil
}

// StopProcess stops a managed process
func (plm *ProcessLifecycleManager) StopProcess(processID int) error {
	plm.mutex.Lock()
	process, exists := plm.processes[processID]
	plm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("process %d not found", processID)
	}

	if process.State != StateRunning {
		return fmt.Errorf("process %d is not running", processID)
	}

	process.State = StateStopping
	fmt.Printf("Stopping process %d: %s\n", process.ID, process.Name)

	if process.cmd != nil && process.cmd.Process != nil {
		// Send SIGTERM first
		err := process.cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			return fmt.Errorf("failed to send SIGTERM to process %d: %v", processID, err)
		}

		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- process.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(10 * time.Second):
			// Force kill after timeout
			fmt.Printf("Process %d didn't exit gracefully, force killing\n", process.ID)
			process.cmd.Process.Kill()
			<-done
		}
	}

	process.State = StateStopped
	now := time.Now()
	process.EndTime = &now

	fmt.Printf("Process %d stopped\n", process.ID)
	return nil
}

// RestartProcess restarts a managed process
func (plm *ProcessLifecycleManager) RestartProcess(processID int) error {
	fmt.Printf("Restarting process %d\n", processID)

	plm.mutex.Lock()
	process, exists := plm.processes[processID]
	plm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("process %d not found", processID)
	}

	if process.RestartCount >= process.MaxRestarts {
		return fmt.Errorf("process %d has exceeded maximum restart count (%d)",
			processID, process.MaxRestarts)
	}

	process.State = StateRestarting
	process.RestartCount++

	// Stop if running
	if process.State == StateRunning || process.State == StateStarting {
		plm.StopProcess(processID)
	}

	// Wait a moment before restarting
	time.Sleep(2 * time.Second)

	// Start again
	return plm.StartProcess(processID)
}

// monitorProcess monitors a process for health and lifecycle
func (plm *ProcessLifecycleManager) monitorProcess(process *ManagedProcess) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if process.State != StateRunning {
				return
			}

			// Health check
			if !process.HealthCheck.Check(process) {
				fmt.Printf("Health check failed for process %d, attempting restart\n", process.ID)
				process.State = StateFailed

				// Attempt restart
				if process.RestartCount < process.MaxRestarts {
					go plm.RestartProcess(process.ID)
				} else {
					fmt.Printf("Process %d has failed and exceeded restart limit\n", process.ID)
				}
				return
			}

		case <-plm.shutdownChannel:
			return
		}
	}
}

// checkDependencies checks if all dependencies are running
func (plm *ProcessLifecycleManager) checkDependencies(process *ManagedProcess) bool {
	for _, depID := range process.Dependencies {
		plm.mutex.RLock()
		dep, exists := plm.processes[depID]
		plm.mutex.RUnlock()

		if !exists || dep.State != StateRunning {
			fmt.Printf("Dependency %d not satisfied for process %d\n", depID, process.ID)
			return false
		}
	}
	return true
}

// GetProcessStatus returns the status of a process
func (plm *ProcessLifecycleManager) GetProcessStatus(processID int) (*ManagedProcess, error) {
	plm.mutex.RLock()
	defer plm.mutex.RUnlock()

	process, exists := plm.processes[processID]
	if !exists {
		return nil, fmt.Errorf("process %d not found", processID)
	}

	return process, nil
}

// ListProcesses lists all managed processes
func (plm *ProcessLifecycleManager) ListProcesses() {
	plm.mutex.RLock()
	defer plm.mutex.RUnlock()

	fmt.Println("=== MANAGED PROCESSES ===")
	fmt.Printf("%-4s %-8s %-20s %-12s %-8s %-10s %s\n",
		"ID", "PID", "NAME", "STATE", "RESTARTS", "UPTIME", "COMMAND")
	fmt.Println(strings.Repeat("-", 80))

	for _, process := range plm.processes {
		uptime := time.Since(process.StartTime)
		if process.EndTime != nil {
			uptime = process.EndTime.Sub(process.StartTime)
		}

		pidStr := "-"
		if process.PID != 0 {
			pidStr = strconv.Itoa(process.PID)
		}

		fmt.Printf("%-4d %-8s %-20s %-12s %-8d %-10s %s\n",
			process.ID,
			pidStr,
			truncateString(process.Name, 20),
			process.State.String(),
			process.RestartCount,
			formatDuration(uptime),
			strings.Join(process.Command, " "))
	}
	fmt.Println()
}

// ShutdownAll gracefully shuts down all processes
func (plm *ProcessLifecycleManager) ShutdownAll() {
	fmt.Println("Shutting down all managed processes...")

	close(plm.shutdownChannel)

	plm.mutex.RLock()
	var runningProcesses []*ManagedProcess
	for _, process := range plm.processes {
		if process.State == StateRunning {
			runningProcesses = append(runningProcesses, process)
		}
	}
	plm.mutex.RUnlock()

	// Stop all running processes
	for _, process := range runningProcesses {
		plm.StopProcess(process.ID)
	}

	fmt.Println("All processes shut down")
}

// SetProcessEnvironment sets environment variables for a process
func (plm *ProcessLifecycleManager) SetProcessEnvironment(processID int, env map[string]string) error {
	plm.mutex.Lock()
	defer plm.mutex.Unlock()

	process, exists := plm.processes[processID]
	if !exists {
		return fmt.Errorf("process %d not found", processID)
	}

	for key, value := range env {
		process.Environment[key] = value
	}

	return nil
}

// SetProcessDependencies sets dependencies for a process
func (plm *ProcessLifecycleManager) SetProcessDependencies(processID int, dependencies []int) error {
	plm.mutex.Lock()
	defer plm.mutex.Unlock()

	process, exists := plm.processes[processID]
	if !exists {
		return fmt.Errorf("process %d not found", processID)
	}

	process.Dependencies = dependencies
	return nil
}

// InteractiveMode provides interactive management interface
func (plm *ProcessLifecycleManager) InteractiveMode() {
	fmt.Println("=== PROCESS LIFECYCLE MANAGER ===")
	fmt.Println("Commands:")
	fmt.Println("  create <name> <command> [args...]  - Create process")
	fmt.Println("  start <id>                         - Start process")
	fmt.Println("  stop <id>                          - Stop process")
	fmt.Println("  restart <id>                       - Restart process")
	fmt.Println("  list                               - List processes")
	fmt.Println("  status <id>                        - Show process status")
	fmt.Println("  env <id> <key=value>               - Set environment variable")
	fmt.Println("  deps <id> <dep1,dep2,...>          - Set dependencies")
	fmt.Println("  shutdown                           - Shutdown all processes")
	fmt.Println("  quit                               - Exit manager")

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal...")
		plm.ShutdownAll()
		os.Exit(0)
	}()

	for {
		var input string
		fmt.Print("manager> ")
		if _, err := fmt.Scanln(&input); err != nil {
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "create":
			if len(parts) < 3 {
				fmt.Println("Usage: create <name> <command> [args...]")
				continue
			}
			name := parts[1]
			cmd := parts[2:]
			process := plm.CreateProcess(name, cmd)
			fmt.Printf("Created process %d\n", process.ID)

		case "start":
			if len(parts) < 2 {
				fmt.Println("Usage: start <id>")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Invalid process ID: %s\n", parts[1])
				continue
			}
			if err := plm.StartProcess(id); err != nil {
				fmt.Printf("Error starting process: %v\n", err)
			}

		case "stop":
			if len(parts) < 2 {
				fmt.Println("Usage: stop <id>")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Invalid process ID: %s\n", parts[1])
				continue
			}
			if err := plm.StopProcess(id); err != nil {
				fmt.Printf("Error stopping process: %v\n", err)
			}

		case "restart":
			if len(parts) < 2 {
				fmt.Println("Usage: restart <id>")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Invalid process ID: %s\n", parts[1])
				continue
			}
			if err := plm.RestartProcess(id); err != nil {
				fmt.Printf("Error restarting process: %v\n", err)
			}

		case "list":
			plm.ListProcesses()

		case "status":
			if len(parts) < 2 {
				fmt.Println("Usage: status <id>")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Invalid process ID: %s\n", parts[1])
				continue
			}
			process, err := plm.GetProcessStatus(id)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			fmt.Printf("Process %d Status:\n", process.ID)
			fmt.Printf("  Name: %s\n", process.Name)
			fmt.Printf("  PID: %d\n", process.PID)
			fmt.Printf("  State: %s\n", process.State)
			fmt.Printf("  Restarts: %d/%d\n", process.RestartCount, process.MaxRestarts)
			fmt.Printf("  Command: %s\n", strings.Join(process.Command, " "))

		case "shutdown":
			plm.ShutdownAll()

		case "quit", "exit":
			plm.ShutdownAll()
			return

		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}

// Utility functions
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go interactive  - Interactive mode")
		fmt.Println("  go run main.go demo         - Run demo")
		os.Exit(1)
	}

	manager := NewProcessLifecycleManager()

	switch os.Args[1] {
	case "interactive":
		manager.InteractiveMode()

	case "demo":
		// Demo mode
		fmt.Println("=== PROCESS LIFECYCLE DEMO ===")

		// Create some demo processes
		proc1 := manager.CreateProcess("sleeper", []string{"sleep", "30"})
		proc2 := manager.CreateProcess("ping", []string{"ping", "-c", "5", "localhost"})

		// Start them
		manager.StartProcess(proc1.ID)
		time.Sleep(1 * time.Second)
		manager.StartProcess(proc2.ID)

		// Show status
		time.Sleep(2 * time.Second)
		manager.ListProcesses()

		// Stop after a while
		time.Sleep(5 * time.Second)
		manager.StopProcess(proc1.ID)

		manager.ListProcesses()

		// Cleanup
		manager.ShutdownAll()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
