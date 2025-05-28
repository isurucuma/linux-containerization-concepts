package process

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// ProcessInfo holds information about a process
type ProcessInfo struct {
	PID     int
	PPID    int
	Name    string
	State   string
	CPU     float64
	Memory  uint64
	Threads int
}

// ProcessNode represents a node in the process tree
type ProcessNode struct {
	Info     ProcessInfo
	Children []*ProcessNode
	Parent   *ProcessNode
}

// GetAllProcesses returns a list of all running processes
func GetAllProcesses() ([]ProcessInfo, error) {
	procDir := "/proc"
	files, err := ioutil.ReadDir(procDir)
	if err != nil {
		return nil, err
	}

	var processes []ProcessInfo

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		// Check if directory name is a number (PID)
		pid, err := strconv.Atoi(file.Name())
		if err != nil {
			continue
		}

		processInfo, err := GetProcessInfo(pid)
		if err != nil {
			// Process might have disappeared, skip it
			continue
		}

		processes = append(processes, processInfo)
	}

	return processes, nil
}

// GetProcessInfo returns detailed information about a specific process
func GetProcessInfo(pid int) (ProcessInfo, error) {
	var info ProcessInfo
	info.PID = pid

	// Read /proc/[pid]/stat
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	statData, err := ioutil.ReadFile(statPath)
	if err != nil {
		return info, err
	}

	statFields := strings.Fields(string(statData))
	if len(statFields) < 24 {
		return info, fmt.Errorf("invalid stat file format")
	}

	// Parse PPID (field 4, 0-indexed field 3)
	info.PPID, _ = strconv.Atoi(statFields[3])

	// Parse state (field 3, 0-indexed field 2)
	info.State = statFields[2]

	// Parse number of threads (field 20, 0-indexed field 19)
	info.Threads, _ = strconv.Atoi(statFields[19])

	// Read process name from /proc/[pid]/comm
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	commData, err := ioutil.ReadFile(commPath)
	if err == nil {
		info.Name = strings.TrimSpace(string(commData))
	}

	// Read memory information from /proc/[pid]/status
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	statusFile, err := os.Open(statusPath)
	if err == nil {
		defer statusFile.Close()
		scanner := bufio.NewScanner(statusFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					memKB, _ := strconv.ParseUint(fields[1], 10, 64)
					info.Memory = memKB * 1024 // Convert to bytes
				}
				break
			}
		}
	}

	// CPU usage calculation would require sampling over time
	// For now, we'll set it to 0 as a placeholder
	info.CPU = 0.0

	return info, nil
}

// BuildProcessTree builds a process tree starting from the given root PID
func BuildProcessTree(rootPID int) (*ProcessNode, error) {
	// Get all processes
	processes, err := GetAllProcesses()
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup
	processMap := make(map[int]ProcessInfo)
	for _, proc := range processes {
		processMap[proc.PID] = proc
	}

	// Find the root process
	rootInfo, exists := processMap[rootPID]
	if !exists {
		return nil, fmt.Errorf("process %d not found", rootPID)
	}

	// Build the tree recursively
	root := &ProcessNode{Info: rootInfo}
	buildTreeRecursive(root, processMap)

	return root, nil
}

// buildTreeRecursive recursively builds the process tree
func buildTreeRecursive(node *ProcessNode, processMap map[int]ProcessInfo) {
	for _, proc := range processMap {
		if proc.PPID == node.Info.PID {
			child := &ProcessNode{
				Info:   proc,
				Parent: node,
			}
			node.Children = append(node.Children, child)
			buildTreeRecursive(child, processMap)
		}
	}
}

// Monitor represents a process monitor
type Monitor struct {
	PID      int
	interval time.Duration
	running  bool
}

// NewMonitor creates a new process monitor
func NewMonitor(pid int) *Monitor {
	return &Monitor{
		PID:      pid,
		interval: 1 * time.Second,
		running:  false,
	}
}

// Start begins monitoring the process
func (m *Monitor) Start() {
	m.running = true
	fmt.Printf("Monitoring process %d (Ctrl+C to stop)\n", m.PID)
	fmt.Println("Time\t\tPID\tCPU%\tMem(MB)\tState\tThreads")
	fmt.Println("────────────────────────────────────────────────────────")

	for m.running {
		info, err := GetProcessInfo(m.PID)
		if err != nil {
			fmt.Printf("Process %d no longer exists\n", m.PID)
			break
		}

		timestamp := time.Now().Format("15:04:05")
		memMB := float64(info.Memory) / 1024 / 1024
		fmt.Printf("%s\t%d\t%.1f\t%.1f\t%s\t%d\n",
			timestamp, info.PID, info.CPU, memMB, info.State, info.Threads)

		time.Sleep(m.interval)
	}
}

// Stop stops monitoring
func (m *Monitor) Stop() {
	m.running = false
}

// Starter represents a process starter
type Starter struct{}

// NewStarter creates a new process starter
func NewStarter() *Starter {
	return &Starter{}
}

// StartProcess starts a new process with the given command and arguments
func (s *Starter) StartProcess(command string, args []string) (int, error) {
	// This is a basic implementation
	// In a real container system, this would involve more setup
	// including namespace creation, cgroup assignment, etc.

	fmt.Printf("Starting process: %s %v\n", command, args)
	fmt.Println("Note: This is a basic implementation for learning purposes")
	fmt.Println("In later sections, we'll enhance this with proper isolation")

	// For now, we'll just return a placeholder PID
	// In the actual implementation, you would use os/exec or syscalls
	return 0, fmt.Errorf("process starting not fully implemented yet - will be completed in later sections")
}
