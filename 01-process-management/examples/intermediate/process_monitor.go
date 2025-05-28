package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// ProcessMonitor represents a real-time process monitor
type ProcessMonitor struct {
	PID         int
	Name        string
	StartTime   time.Time
	LastCPUTime uint64
	LastSample  time.Time
}

// CPUStats holds CPU time information from /proc/[pid]/stat
type CPUStats struct {
	UTime uint64 // User time
	STime uint64 // System time
	Total uint64 // Total CPU time
}

// MemoryStats holds memory information
type MemoryStats struct {
	VmSize uint64 // Virtual memory size
	VmRSS  uint64 // Resident set size
	VmHWM  uint64 // Peak resident set size
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run process_monitor.go <PID>")
		fmt.Println("Example: go run process_monitor.go 1234")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Invalid PID: %s\n", os.Args[1])
		os.Exit(1)
	}

	monitor := &ProcessMonitor{
		PID:       pid,
		StartTime: time.Now(),
	}

	// Get process name
	if name, err := getProcessName(pid); err == nil {
		monitor.Name = name
	} else {
		monitor.Name = "unknown"
	}

	fmt.Printf("Monitoring process %d (%s) - Press Ctrl+C to stop\n", pid, monitor.Name)
	fmt.Println("Time\t\tCPU%\tMem(MB)\tVMem(MB)\tState\tThreads\tFDs")
	fmt.Println(strings.Repeat("─", 70))

	// Initial CPU reading
	initialCPU, err := getCPUStats(pid)
	if err != nil {
		fmt.Printf("Error reading initial CPU stats: %v\n", err)
		os.Exit(1)
	}
	monitor.LastCPUTime = initialCPU.Total
	monitor.LastSample = time.Now()

	// Monitor loop
	for {
		time.Sleep(1 * time.Second)

		if err := monitor.displayStats(); err != nil {
			fmt.Printf("Process %d no longer exists\n", pid)
			break
		}
	}
}

func (m *ProcessMonitor) displayStats() error {
	timestamp := time.Now().Format("15:04:05")

	// Get CPU stats
	cpuStats, err := getCPUStats(m.PID)
	if err != nil {
		return err
	}

	// Calculate CPU percentage
	cpuPercent := m.calculateCPUPercent(cpuStats)

	// Get memory stats
	memStats, err := getMemoryStats(m.PID)
	if err != nil {
		return err
	}

	// Get process state
	state, threads, err := getProcessState(m.PID)
	if err != nil {
		return err
	}

	// Count file descriptors
	fdCount := countFileDescriptors(m.PID)

	// Display the information
	memMB := float64(memStats.VmRSS) / 1024 / 1024
	vmemMB := float64(memStats.VmSize) / 1024 / 1024

	fmt.Printf("%s\t%.1f\t%.1f\t%.1f\t\t%s\t%d\t%d\n",
		timestamp, cpuPercent, memMB, vmemMB, state, threads, fdCount)

	return nil
}

func (m *ProcessMonitor) calculateCPUPercent(current CPUStats) float64 {
	now := time.Now()
	timeDiff := now.Sub(m.LastSample).Seconds()

	if timeDiff == 0 {
		return 0.0
	}

	// CPU time is in clock ticks, convert to seconds
	clockTicks := 100.0 // sysconf(_SC_CLK_TCK) is typically 100

	cpuTimeDiff := float64(current.Total-m.LastCPUTime) / clockTicks
	cpuPercent := (cpuTimeDiff / timeDiff) * 100.0

	// Update for next calculation
	m.LastCPUTime = current.Total
	m.LastSample = now

	return cpuPercent
}

func getProcessName(pid int) (string, error) {
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	data, err := ioutil.ReadFile(commPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func getCPUStats(pid int) (CPUStats, error) {
	var stats CPUStats

	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := ioutil.ReadFile(statPath)
	if err != nil {
		return stats, err
	}

	// Parse the stat file
	line := string(data)

	// Find the last ')' to handle process names with spaces
	lastParen := strings.LastIndex(line, ")")
	if lastParen == -1 {
		return stats, fmt.Errorf("invalid stat format")
	}

	fields := strings.Fields(line[lastParen+1:])
	if len(fields) < 15 {
		return stats, fmt.Errorf("insufficient fields in stat")
	}

	// Parse utime (field 11) and stime (field 12)
	utime, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return stats, err
	}

	stime, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return stats, err
	}

	stats.UTime = utime
	stats.STime = stime
	stats.Total = utime + stime

	return stats, nil
}

func getMemoryStats(pid int) (MemoryStats, error) {
	var stats MemoryStats

	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	file, err := os.Open(statusPath)
	if err != nil {
		return stats, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "VmSize:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				stats.VmSize = val * 1024 // Convert from KB to bytes
			}
		case "VmRSS:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				stats.VmRSS = val * 1024 // Convert from KB to bytes
			}
		case "VmHWM:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				stats.VmHWM = val * 1024 // Convert from KB to bytes
			}
		}
	}

	return stats, scanner.Err()
}

func getProcessState(pid int) (string, int, error) {
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	file, err := os.Open(statusPath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	var state string
	var threads int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "State:":
			state = fields[1]
		case "Threads:":
			if val, err := strconv.Atoi(fields[1]); err == nil {
				threads = val
			}
		}
	}

	return state, threads, scanner.Err()
}

func countFileDescriptors(pid int) int {
	fdPath := fmt.Sprintf("/proc/%d/fd", pid)
	files, err := ioutil.ReadDir(fdPath)
	if err != nil {
		return -1
	}
	return len(files)
}
