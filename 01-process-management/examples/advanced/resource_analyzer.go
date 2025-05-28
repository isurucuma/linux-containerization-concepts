package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SystemMetrics represents system-wide process metrics
type SystemMetrics struct {
	TotalProcesses    int
	RunningProcesses  int
	SleepingProcesses int
	ZombieProcesses   int
	TotalMemory       uint64
	UsedMemory        uint64
	TotalCPUTime      uint64
	LoadAverage       [3]float64
}

// ProcessMetrics represents individual process metrics
type ProcessMetrics struct {
	PID        int
	Name       string
	CPUPercent float64
	MemoryRSS  uint64
	MemoryVSZ  uint64
	State      string
	Threads    int
	FileDesc   int
	CPUTime    uint64
	StartTime  uint64
	LastSeen   time.Time
}

// ResourceAnalyzer analyzes system and process resources
type ResourceAnalyzer struct {
	processes       map[int]*ProcessMetrics
	history         []SystemMetrics
	mutex           sync.RWMutex
	updateInterval  time.Duration
	historySize     int
	alertThresholds map[string]float64
}

// NewResourceAnalyzer creates a new resource analyzer
func NewResourceAnalyzer() *ResourceAnalyzer {
	return &ResourceAnalyzer{
		processes:      make(map[int]*ProcessMetrics),
		history:        make([]SystemMetrics, 0),
		updateInterval: time.Second,
		historySize:    300, // 5 minutes of data
		alertThresholds: map[string]float64{
			"cpu_high":    80.0, // CPU usage > 80%
			"memory_high": 90.0, // Memory usage > 90%
			"zombie_high": 10.0, // More than 10 zombie processes
		},
	}
}

// CollectSystemMetrics collects system-wide metrics
func (ra *ResourceAnalyzer) CollectSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// Read /proc/stat for CPU information
	statData, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(statData), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) >= 8 {
				for i := 1; i < 8; i++ {
					val, _ := strconv.ParseUint(fields[i], 10, 64)
					metrics.TotalCPUTime += val
				}
			}
			break
		}
	}

	// Read /proc/meminfo for memory information
	meminfoData, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	meminfoLines := strings.Split(string(meminfoData), "\n")
	for _, line := range meminfoLines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, _ := strconv.ParseUint(fields[1], 10, 64)
				metrics.TotalMemory = val * 1024 // Convert KB to bytes
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				available, _ := strconv.ParseUint(fields[1], 10, 64)
				metrics.UsedMemory = metrics.TotalMemory - (available * 1024)
			}
		}
	}

	// Read /proc/loadavg for load average
	loadavgData, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}

	loadFields := strings.Fields(string(loadavgData))
	if len(loadFields) >= 3 {
		metrics.LoadAverage[0], _ = strconv.ParseFloat(loadFields[0], 64)
		metrics.LoadAverage[1], _ = strconv.ParseFloat(loadFields[1], 64)
		metrics.LoadAverage[2], _ = strconv.ParseFloat(loadFields[2], 64)
	}

	// Count processes by state
	ra.countProcessesByState(metrics)

	return metrics, nil
}

// countProcessesByState counts processes by their state
func (ra *ResourceAnalyzer) countProcessesByState(metrics *SystemMetrics) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		statPath := fmt.Sprintf("/proc/%d/stat", pid)
		statData, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}

		fields := strings.Fields(string(statData))
		if len(fields) < 3 {
			continue
		}

		state := fields[2]
		metrics.TotalProcesses++

		switch state {
		case "R":
			metrics.RunningProcesses++
		case "S", "D", "I":
			metrics.SleepingProcesses++
		case "Z":
			metrics.ZombieProcesses++
		}
	}
}

// CollectProcessMetrics collects metrics for all processes
func (ra *ResourceAnalyzer) CollectProcessMetrics() error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	currentTime := time.Now()
	newProcesses := make(map[int]*ProcessMetrics)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		metrics, err := ra.collectSingleProcessMetrics(pid)
		if err != nil {
			continue
		}

		metrics.LastSeen = currentTime

		// Calculate CPU percentage if we have previous data
		if prevMetrics, exists := ra.processes[pid]; exists {
			timeDelta := currentTime.Sub(prevMetrics.LastSeen).Seconds()
			cpuDelta := float64(metrics.CPUTime - prevMetrics.CPUTime)
			if timeDelta > 0 {
				metrics.CPUPercent = (cpuDelta / 100.0) / timeDelta * 100.0 // Assuming 100 HZ
			}
		}

		newProcesses[pid] = metrics
	}

	ra.mutex.Lock()
	ra.processes = newProcesses
	ra.mutex.Unlock()

	return nil
}

// collectSingleProcessMetrics collects metrics for a single process
func (ra *ResourceAnalyzer) collectSingleProcessMetrics(pid int) (*ProcessMetrics, error) {
	metrics := &ProcessMetrics{PID: pid}

	// Read /proc/[pid]/stat
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	statData, err := os.ReadFile(statPath)
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(statData))
	if len(fields) < 24 {
		return nil, fmt.Errorf("insufficient fields in stat file")
	}

	// Parse process name
	metrics.Name = strings.Trim(fields[1], "()")

	// Parse state
	metrics.State = fields[2]

	// Parse CPU times
	utime, _ := strconv.ParseUint(fields[13], 10, 64)
	stime, _ := strconv.ParseUint(fields[14], 10, 64)
	metrics.CPUTime = utime + stime

	// Parse number of threads
	metrics.Threads, _ = strconv.Atoi(fields[19])

	// Parse start time
	metrics.StartTime, _ = strconv.ParseUint(fields[21], 10, 64)

	// Parse virtual memory size
	metrics.MemoryVSZ, _ = strconv.ParseUint(fields[22], 10, 64)

	// Parse RSS
	rssPages, _ := strconv.ParseUint(fields[23], 10, 64)
	metrics.MemoryRSS = rssPages * 4096 // Assuming 4KB pages

	// Count file descriptors
	metrics.FileDesc = ra.countFileDescriptors(pid)

	return metrics, nil
}

// countFileDescriptors counts open file descriptors for a process
func (ra *ResourceAnalyzer) countFileDescriptors(pid int) int {
	fdDir := fmt.Sprintf("/proc/%d/fd", pid)
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return 0
	}
	return len(entries)
}

// StartMonitoring starts continuous monitoring
func (ra *ResourceAnalyzer) StartMonitoring() {
	ticker := time.NewTicker(ra.updateInterval)
	defer ticker.Stop()

	fmt.Println("Starting resource monitoring...")

	for range ticker.C {
		// Collect system metrics
		systemMetrics, err := ra.CollectSystemMetrics()
		if err != nil {
			log.Printf("Error collecting system metrics: %v", err)
			continue
		}

		// Collect process metrics
		err = ra.CollectProcessMetrics()
		if err != nil {
			log.Printf("Error collecting process metrics: %v", err)
			continue
		}

		// Add to history
		ra.mutex.Lock()
		ra.history = append(ra.history, *systemMetrics)
		if len(ra.history) > ra.historySize {
			ra.history = ra.history[1:]
		}
		ra.mutex.Unlock()

		// Check for alerts
		ra.checkAlerts(systemMetrics)
	}
}

// checkAlerts checks for alert conditions
func (ra *ResourceAnalyzer) checkAlerts(metrics *SystemMetrics) {
	memoryPercent := float64(metrics.UsedMemory) / float64(metrics.TotalMemory) * 100

	if memoryPercent > ra.alertThresholds["memory_high"] {
		fmt.Printf("[ALERT] High memory usage: %.1f%%\n", memoryPercent)
	}

	if float64(metrics.ZombieProcesses) > ra.alertThresholds["zombie_high"] {
		fmt.Printf("[ALERT] High zombie process count: %d\n", metrics.ZombieProcesses)
	}

	if metrics.LoadAverage[0] > 2.0 {
		fmt.Printf("[ALERT] High load average: %.2f\n", metrics.LoadAverage[0])
	}
}

// GetTopCPUProcesses returns top N processes by CPU usage
func (ra *ResourceAnalyzer) GetTopCPUProcesses(n int) []*ProcessMetrics {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	var processes []*ProcessMetrics
	for _, p := range ra.processes {
		processes = append(processes, p)
	}

	// Sort by CPU percentage
	for i := 0; i < len(processes)-1; i++ {
		for j := i + 1; j < len(processes); j++ {
			if processes[i].CPUPercent < processes[j].CPUPercent {
				processes[i], processes[j] = processes[j], processes[i]
			}
		}
	}

	if n > len(processes) {
		n = len(processes)
	}

	return processes[:n]
}

// GetTopMemoryProcesses returns top N processes by memory usage
func (ra *ResourceAnalyzer) GetTopMemoryProcesses(n int) []*ProcessMetrics {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	var processes []*ProcessMetrics
	for _, p := range ra.processes {
		processes = append(processes, p)
	}

	// Sort by memory RSS
	for i := 0; i < len(processes)-1; i++ {
		for j := i + 1; j < len(processes); j++ {
			if processes[i].MemoryRSS < processes[j].MemoryRSS {
				processes[i], processes[j] = processes[j], processes[i]
			}
		}
	}

	if n > len(processes) {
		n = len(processes)
	}

	return processes[:n]
}

// DisplaySystemSummary displays system resource summary
func (ra *ResourceAnalyzer) DisplaySystemSummary() {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	if len(ra.history) == 0 {
		fmt.Println("No system metrics available")
		return
	}

	latest := ra.history[len(ra.history)-1]

	fmt.Println("=== SYSTEM RESOURCE SUMMARY ===")
	fmt.Printf("Total Processes: %d\n", latest.TotalProcesses)
	fmt.Printf("Running: %d, Sleeping: %d, Zombie: %d\n",
		latest.RunningProcesses, latest.SleepingProcesses, latest.ZombieProcesses)

	memoryPercent := float64(latest.UsedMemory) / float64(latest.TotalMemory) * 100
	fmt.Printf("Memory: %s / %s (%.1f%%)\n",
		formatBytes(latest.UsedMemory), formatBytes(latest.TotalMemory), memoryPercent)

	fmt.Printf("Load Average: %.2f %.2f %.2f\n",
		latest.LoadAverage[0], latest.LoadAverage[1], latest.LoadAverage[2])
	fmt.Println()
}

// DisplayTopProcesses displays top processes
func (ra *ResourceAnalyzer) DisplayTopProcesses() {
	fmt.Println("=== TOP CPU PROCESSES ===")
	cpuProcs := ra.GetTopCPUProcesses(5)
	fmt.Printf("%-8s %-20s %-8s %-8s %-8s\n", "PID", "NAME", "CPU%", "THREADS", "STATE")
	fmt.Println(strings.Repeat("-", 55))
	for _, p := range cpuProcs {
		fmt.Printf("%-8d %-20s %-8.1f %-8d %-8s\n",
			p.PID, truncateString(p.Name, 20), p.CPUPercent, p.Threads, p.State)
	}

	fmt.Println("\n=== TOP MEMORY PROCESSES ===")
	memProcs := ra.GetTopMemoryProcesses(5)
	fmt.Printf("%-8s %-20s %-10s %-10s %-8s\n", "PID", "NAME", "RSS", "VSZ", "FDs")
	fmt.Println(strings.Repeat("-", 60))
	for _, p := range memProcs {
		fmt.Printf("%-8d %-20s %-10s %-10s %-8d\n",
			p.PID, truncateString(p.Name, 20),
			formatBytes(p.MemoryRSS), formatBytes(p.MemoryVSZ), p.FileDesc)
	}
	fmt.Println()
}

// GenerateReport generates a detailed system report
func (ra *ResourceAnalyzer) GenerateReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write report header
	fmt.Fprintf(writer, "System Resource Analysis Report\n")
	fmt.Fprintf(writer, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// System summary
	ra.mutex.RLock()
	if len(ra.history) > 0 {
		latest := ra.history[len(ra.history)-1]
		fmt.Fprintf(writer, "=== SYSTEM SUMMARY ===\n")
		fmt.Fprintf(writer, "Total Processes: %d\n", latest.TotalProcesses)
		fmt.Fprintf(writer, "Running: %d, Sleeping: %d, Zombie: %d\n",
			latest.RunningProcesses, latest.SleepingProcesses, latest.ZombieProcesses)

		memoryPercent := float64(latest.UsedMemory) / float64(latest.TotalMemory) * 100
		fmt.Fprintf(writer, "Memory Usage: %.1f%%\n", memoryPercent)
		fmt.Fprintf(writer, "Load Average: %.2f %.2f %.2f\n\n",
			latest.LoadAverage[0], latest.LoadAverage[1], latest.LoadAverage[2])
	}
	ra.mutex.RUnlock()

	// Top processes
	fmt.Fprintf(writer, "=== TOP CPU PROCESSES ===\n")
	cpuProcs := ra.GetTopCPUProcesses(10)
	for _, p := range cpuProcs {
		fmt.Fprintf(writer, "PID: %d, Name: %s, CPU: %.1f%%, Threads: %d\n",
			p.PID, p.Name, p.CPUPercent, p.Threads)
	}

	fmt.Fprintf(writer, "\n=== TOP MEMORY PROCESSES ===\n")
	memProcs := ra.GetTopMemoryProcesses(10)
	for _, p := range memProcs {
		fmt.Fprintf(writer, "PID: %d, Name: %s, RSS: %s, VSZ: %s\n",
			p.PID, p.Name, formatBytes(p.MemoryRSS), formatBytes(p.MemoryVSZ))
	}

	fmt.Printf("Report saved to: %s\n", filename)
	return nil
}

// Interactive mode for the analyzer
func (ra *ResourceAnalyzer) InteractiveMode() {
	fmt.Println("=== RESOURCE ANALYZER - INTERACTIVE MODE ===")
	fmt.Println("Commands:")
	fmt.Println("  summary  - Show system summary")
	fmt.Println("  top      - Show top processes")
	fmt.Println("  report   - Generate report")
	fmt.Println("  start    - Start monitoring")
	fmt.Println("  quit     - Exit")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("analyzer> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "summary":
			ra.CollectSystemMetrics()
			ra.DisplaySystemSummary()

		case "top":
			ra.CollectProcessMetrics()
			ra.DisplayTopProcesses()

		case "report":
			filename := "system_report.txt"
			if len(parts) > 1 {
				filename = parts[1]
			}
			ra.GenerateReport(filename)

		case "start":
			go ra.StartMonitoring()
			fmt.Println("Monitoring started in background")

		case "quit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}

// Utility functions
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
		fmt.Println("  go run main.go monitor      - Start monitoring")
		fmt.Println("  go run main.go report [file] - Generate report")
		os.Exit(1)
	}

	analyzer := NewResourceAnalyzer()

	switch os.Args[1] {
	case "interactive":
		analyzer.InteractiveMode()

	case "monitor":
		analyzer.StartMonitoring()

	case "report":
		filename := "system_report.txt"
		if len(os.Args) > 2 {
			filename = os.Args[2]
		}
		analyzer.CollectSystemMetrics()
		analyzer.CollectProcessMetrics()
		analyzer.GenerateReport(filename)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
