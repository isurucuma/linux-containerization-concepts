package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ProcessStats represents detailed process statistics
type ProcessStats struct {
	PID       int
	Name      string
	State     string
	CPUTime   uint64
	VmSize    uint64 // Virtual memory size in bytes
	VmRSS     uint64 // Resident Set Size in bytes
	Threads   int
	FDCount   int // File descriptor count
	Priority  int
	Nice      int
	StartTime uint64
}

// ResourceMonitor manages process resource monitoring
type ResourceMonitor struct {
	processes map[int]*ProcessStats
	jiffies   int64 // Clock ticks per second
	pageSize  int64 // Memory page size
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		processes: make(map[int]*ProcessStats),
		jiffies:   100,  // Default to 100 Hz
		pageSize:  4096, // Default page size
	}
}

// parseStatFile parses /proc/[pid]/stat file
func (rm *ResourceMonitor) parseStatFile(pid int) (*ProcessStats, error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 44 {
		return nil, fmt.Errorf("insufficient fields in stat file")
	}

	stats := &ProcessStats{PID: pid}

	// Parse process name (remove parentheses)
	stats.Name = strings.Trim(fields[1], "()")

	// Parse state
	stats.State = fields[2]

	// Parse CPU times (user + system time in jiffies)
	utime, _ := strconv.ParseUint(fields[13], 10, 64)
	stime, _ := strconv.ParseUint(fields[14], 10, 64)
	stats.CPUTime = utime + stime

	// Parse priority and nice
	stats.Priority, _ = strconv.Atoi(fields[17])
	stats.Nice, _ = strconv.Atoi(fields[18])

	// Parse number of threads
	stats.Threads, _ = strconv.Atoi(fields[19])

	// Parse start time
	stats.StartTime, _ = strconv.ParseUint(fields[21], 10, 64)

	// Parse virtual memory size
	stats.VmSize, _ = strconv.ParseUint(fields[22], 10, 64)

	// Parse RSS in pages, convert to bytes
	rssPages, _ := strconv.ParseUint(fields[23], 10, 64)
	stats.VmRSS = rssPages * uint64(rm.pageSize)

	return stats, nil
}

// parseStatusFile parses additional info from /proc/[pid]/status
func (rm *ResourceMonitor) parseStatusFile(pid int) error {
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return err
	}

	process := rm.processes[pid]
	if process == nil {
		return fmt.Errorf("process not found")
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VmSize:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				size, _ := strconv.ParseUint(fields[1], 10, 64)
				process.VmSize = size * 1024 // Convert from KB to bytes
			}
		} else if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				rss, _ := strconv.ParseUint(fields[1], 10, 64)
				process.VmRSS = rss * 1024 // Convert from KB to bytes
			}
		} else if strings.HasPrefix(line, "Threads:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				process.Threads, _ = strconv.Atoi(fields[1])
			}
		}
	}

	return nil
}

// countFileDescriptors counts open file descriptors
func (rm *ResourceMonitor) countFileDescriptors(pid int) int {
	fdDir := fmt.Sprintf("/proc/%d/fd", pid)
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return 0
	}
	return len(entries)
}

// scanProcess scans and updates process information
func (rm *ResourceMonitor) scanProcess(pid int) error {
	stats, err := rm.parseStatFile(pid)
	if err != nil {
		return err
	}

	// Count file descriptors
	stats.FDCount = rm.countFileDescriptors(pid)

	// Store in map
	rm.processes[pid] = stats

	// Parse additional status information
	rm.parseStatusFile(pid)

	return nil
}

// ScanAllProcesses scans all processes in the system
func (rm *ResourceMonitor) ScanAllProcesses() error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	// Clear previous data
	rm.processes = make(map[int]*ProcessStats)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // Not a PID directory
		}

		rm.scanProcess(pid)
	}

	return nil
}

// GetTopMemoryProcesses returns top N processes by memory usage
func (rm *ResourceMonitor) GetTopMemoryProcesses(n int) []*ProcessStats {
	var processes []*ProcessStats
	for _, p := range rm.processes {
		processes = append(processes, p)
	}

	sort.Slice(processes, func(i, j int) bool {
		return processes[i].VmRSS > processes[j].VmRSS
	})

	if n > len(processes) {
		n = len(processes)
	}

	return processes[:n]
}

// GetTopCPUProcesses returns top N processes by CPU time
func (rm *ResourceMonitor) GetTopCPUProcesses(n int) []*ProcessStats {
	var processes []*ProcessStats
	for _, p := range rm.processes {
		processes = append(processes, p)
	}

	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPUTime > processes[j].CPUTime
	})

	if n > len(processes) {
		n = len(processes)
	}

	return processes[:n]
}

// GetProcessesByState returns processes filtered by state
func (rm *ResourceMonitor) GetProcessesByState(state string) []*ProcessStats {
	var processes []*ProcessStats
	for _, p := range rm.processes {
		if p.State == state {
			processes = append(processes, p)
		}
	}
	return processes
}

// formatBytes formats bytes in human readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats time duration in human readable format
func formatDuration(jiffies uint64, hz int64) string {
	seconds := float64(jiffies) / float64(hz)
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}
	minutes := int(seconds / 60)
	seconds = float64(int64(seconds) % 60)
	if minutes < 60 {
		return fmt.Sprintf("%dm%.0fs", minutes, seconds)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh%dm%.0fs", hours, minutes, seconds)
}

// DisplayTopMemory displays top memory consuming processes
func (rm *ResourceMonitor) DisplayTopMemory(n int) {
	fmt.Printf("=== TOP %d MEMORY CONSUMING PROCESSES ===\n", n)
	fmt.Printf("%-8s %-20s %-10s %-10s %-8s %-8s\n",
		"PID", "NAME", "VMSIZE", "RSS", "STATE", "THREADS")
	fmt.Println(strings.Repeat("-", 70))

	processes := rm.GetTopMemoryProcesses(n)
	for _, p := range processes {
		fmt.Printf("%-8d %-20s %-10s %-10s %-8s %-8d\n",
			p.PID,
			truncateString(p.Name, 20),
			formatBytes(p.VmSize),
			formatBytes(p.VmRSS),
			p.State,
			p.Threads)
	}
	fmt.Println()
}

// DisplayTopCPU displays top CPU consuming processes
func (rm *ResourceMonitor) DisplayTopCPU(n int) {
	fmt.Printf("=== TOP %d CPU CONSUMING PROCESSES ===\n", n)
	fmt.Printf("%-8s %-20s %-12s %-8s %-8s %-8s\n",
		"PID", "NAME", "CPU_TIME", "PRIORITY", "NICE", "FDs")
	fmt.Println(strings.Repeat("-", 70))

	processes := rm.GetTopCPUProcesses(n)
	for _, p := range processes {
		fmt.Printf("%-8d %-20s %-12s %-8d %-8d %-8d\n",
			p.PID,
			truncateString(p.Name, 20),
			formatDuration(p.CPUTime, rm.jiffies),
			p.Priority,
			p.Nice,
			p.FDCount)
	}
	fmt.Println()
}

// DisplayProcessStates displays process count by state
func (rm *ResourceMonitor) DisplayProcessStates() {
	stateMap := map[string]string{
		"R": "Running",
		"S": "Sleeping (interruptible)",
		"D": "Waiting (uninterruptible)",
		"Z": "Zombie",
		"T": "Stopped",
		"t": "Tracing stop",
		"X": "Dead",
		"x": "Dead",
		"K": "Wakekill",
		"W": "Waking",
		"P": "Parked",
	}

	stateCounts := make(map[string]int)
	for _, p := range rm.processes {
		stateCounts[p.State]++
	}

	fmt.Println("=== PROCESS STATES ===")
	fmt.Printf("%-6s %-25s %s\n", "STATE", "DESCRIPTION", "COUNT")
	fmt.Println(strings.Repeat("-", 45))

	for state, count := range stateCounts {
		description := stateMap[state]
		if description == "" {
			description = "Unknown"
		}
		fmt.Printf("%-6s %-25s %d\n", state, description, count)
	}
	fmt.Println()
}

// MonitorProcess monitors a specific process over time
func (rm *ResourceMonitor) MonitorProcess(pid int, duration time.Duration) {
	fmt.Printf("=== MONITORING PROCESS %d ===\n", pid)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	var prevCPUTime uint64

	for {
		select {
		case <-ticker.C:
			if time.Since(startTime) > duration {
				return
			}

			err := rm.scanProcess(pid)
			if err != nil {
				fmt.Printf("Process %d no longer exists\n", pid)
				return
			}

			process := rm.processes[pid]
			cpuDelta := process.CPUTime - prevCPUTime
			prevCPUTime = process.CPUTime

			fmt.Printf("[%s] PID=%d CPU_TIME=%s (+%d jiffies) RSS=%s VmSize=%s Threads=%d FDs=%d State=%s\n",
				time.Now().Format("15:04:05"),
				process.PID,
				formatDuration(process.CPUTime, rm.jiffies),
				cpuDelta,
				formatBytes(process.VmRSS),
				formatBytes(process.VmSize),
				process.Threads,
				process.FDCount,
				process.State)
		}
	}
}

// truncateString truncates string to specified length
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go scan                    - Scan all processes")
		fmt.Println("  go run main.go top-memory [N]          - Show top N memory processes")
		fmt.Println("  go run main.go top-cpu [N]             - Show top N CPU processes")
		fmt.Println("  go run main.go states                  - Show process states")
		fmt.Println("  go run main.go monitor <PID> [seconds] - Monitor specific process")
		os.Exit(1)
	}

	monitor := NewResourceMonitor()

	switch os.Args[1] {
	case "scan":
		fmt.Println("Scanning all processes...")
		if err := monitor.ScanAllProcesses(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Found %d processes\n", len(monitor.processes))

	case "top-memory":
		n := 10
		if len(os.Args) > 2 {
			if parsed, err := strconv.Atoi(os.Args[2]); err == nil {
				n = parsed
			}
		}
		if err := monitor.ScanAllProcesses(); err != nil {
			log.Fatal(err)
		}
		monitor.DisplayTopMemory(n)

	case "top-cpu":
		n := 10
		if len(os.Args) > 2 {
			if parsed, err := strconv.Atoi(os.Args[2]); err == nil {
				n = parsed
			}
		}
		if err := monitor.ScanAllProcesses(); err != nil {
			log.Fatal(err)
		}
		monitor.DisplayTopCPU(n)

	case "states":
		if err := monitor.ScanAllProcesses(); err != nil {
			log.Fatal(err)
		}
		monitor.DisplayProcessStates()

	case "monitor":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go monitor <PID> [seconds]")
			os.Exit(1)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Invalid PID:", err)
		}

		duration := 30 * time.Second
		if len(os.Args) > 3 {
			if seconds, err := strconv.Atoi(os.Args[3]); err == nil {
				duration = time.Duration(seconds) * time.Second
			}
		}

		monitor.MonitorProcess(pid, duration)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
