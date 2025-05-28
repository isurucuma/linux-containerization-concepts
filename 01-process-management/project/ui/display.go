package ui

import (
	"fmt"
	"sort"
	"strings"

	"process-manager/process"
)

// DisplayProcessList displays a formatted list of processes
func DisplayProcessList(processes []process.ProcessInfo) {
	if len(processes) == 0 {
		fmt.Println("No processes found")
		return
	}

	// Sort processes by PID
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].PID < processes[j].PID
	})

	// Print header
	fmt.Printf("%-8s %-8s %-20s %-8s %-8s %-10s %-8s\n",
		"PID", "PPID", "NAME", "STATE", "CPU%", "MEM(MB)", "THREADS")
	fmt.Println(strings.Repeat("─", 80))

	// Print process information
	for _, proc := range processes {
		memMB := float64(proc.Memory) / 1024 / 1024
		fmt.Printf("%-8d %-8d %-20s %-8s %-8.1f %-10.1f %-8d\n",
			proc.PID, proc.PPID, truncateString(proc.Name, 20),
			proc.State, proc.CPU, memMB, proc.Threads)
	}

	fmt.Printf("\nTotal processes: %d\n", len(processes))
}

// DisplayProcessTree displays a process tree in a hierarchical format
func DisplayProcessTree(root *process.ProcessNode) {
	fmt.Printf("Process Tree (starting from PID %d):\n", root.Info.PID)
	fmt.Println(strings.Repeat("─", 50))
	displayNode(root, "", true)
}

// displayNode recursively displays a process tree node
func displayNode(node *process.ProcessNode, prefix string, isLast bool) {
	// Determine the tree characters
	var nodeChar, childPrefix string
	if isLast {
		nodeChar = "└── "
		childPrefix = prefix + "    "
	} else {
		nodeChar = "├── "
		childPrefix = prefix + "│   "
	}

	// Display current node
	memMB := float64(node.Info.Memory) / 1024 / 1024
	fmt.Printf("%s%s%s (%d) [%s] %.1fMB\n",
		prefix, nodeChar, node.Info.Name, node.Info.PID, node.Info.State, memMB)

	// Display children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		displayNode(child, childPrefix, isLastChild)
	}
}

// truncateString truncates a string to the specified length
func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}

// DisplaySignalInfo displays information about available signals
func DisplaySignalInfo() {
	fmt.Println("Available Signals:")
	fmt.Println(strings.Repeat("─", 50))

	signals := []string{
		"SIGTERM", "SIGKILL", "SIGINT", "SIGSTOP",
		"SIGCONT", "SIGHUP", "SIGUSR1", "SIGUSR2",
		"SIGQUIT", "SIGABRT",
	}

	descriptions := map[string]string{
		"SIGTERM": "Polite termination request",
		"SIGKILL": "Immediate termination (cannot be caught)",
		"SIGINT":  "Interrupt from keyboard (Ctrl+C)",
		"SIGSTOP": "Stop process (cannot be caught)",
		"SIGCONT": "Continue stopped process",
		"SIGHUP":  "Terminal disconnection",
		"SIGUSR1": "User-defined signal 1",
		"SIGUSR2": "User-defined signal 2",
		"SIGQUIT": "Quit signal (Ctrl+\\)",
		"SIGABRT": "Abnormal termination",
	}

	for _, sig := range signals {
		fmt.Printf("%-10s - %s\n", sig, descriptions[sig])
	}
}
