package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ProcessInfo holds information about a process
type ProcessInfo struct {
	PID     int
	PPID    int
	Name    string
	State   string
	Memory  uint64
	Threads int
}

// ProcessNode represents a node in the process tree
type ProcessNode struct {
	Info     ProcessInfo
	Children []*ProcessNode
	Parent   *ProcessNode
}

// ReadProcessInfo reads process information from /proc/<pid>/
func ReadProcessInfo(pid int) (ProcessInfo, error) {
	var info ProcessInfo
	info.PID = pid

	// Read /proc/<pid>/stat
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	statData, err := ioutil.ReadFile(statPath)
	if err != nil {
		return info, fmt.Errorf("failed to read stat file: %v", err)
	}

	// Parse stat file - handle process names with spaces/parentheses
	statStr := string(statData)

	// Find the last ')' to handle process names with parentheses
	lastParen := strings.LastIndex(statStr, ")")
	if lastParen == -1 {
		return info, fmt.Errorf("invalid stat file format")
	}

	// Extract process name (between first '(' and last ')')
	firstParen := strings.Index(statStr, "(")
	if firstParen == -1 || firstParen >= lastParen {
		return info, fmt.Errorf("invalid stat file format")
	}
	info.Name = statStr[firstParen+1 : lastParen]

	// Parse fields after the last ')'
	fields := strings.Fields(statStr[lastParen+1:])
	if len(fields) < 20 {
		return info, fmt.Errorf("insufficient fields in stat file")
	}

	// Parse PPID (field 1 after name)
	if ppid, err := strconv.Atoi(fields[1]); err == nil {
		info.PPID = ppid
	}

	// Parse state (field 0 after name)
	info.State = strings.TrimSpace(fields[0])

	// Parse number of threads (field 17 after name)
	if threads, err := strconv.Atoi(fields[17]); err == nil {
		info.Threads = threads
	}

	// Read memory information from /proc/<pid>/status
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
					if memKB, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						info.Memory = memKB * 1024 // Convert to bytes
					}
				}
				break
			}
		}
	}

	return info, nil
}

// GetAllProcesses returns a list of all processes
func GetAllProcesses() ([]ProcessInfo, error) {
	procDir := "/proc"
	files, err := ioutil.ReadDir(procDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc directory: %v", err)
	}

	var processes []ProcessInfo
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		// Check if directory name is a number (PID)
		pid, err := strconv.Atoi(file.Name())
		if err != nil {
			continue // Skip non-numeric directories
		}

		processInfo, err := ReadProcessInfo(pid)
		if err != nil {
			// Process might have disappeared, skip it
			continue
		}

		processes = append(processes, processInfo)
	}

	return processes, nil
}

// BuildProcessTree builds a process tree starting from the given root PID
func BuildProcessTree(rootPID int) (*ProcessNode, error) {
	// Get all processes
	processes, err := GetAllProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %v", err)
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
	var children []ProcessInfo

	// Find all children of this process
	for _, proc := range processMap {
		if proc.PPID == node.Info.PID {
			children = append(children, proc)
		}
	}

	// Sort children by PID for consistent output
	sort.Slice(children, func(i, j int) bool {
		return children[i].PID < children[j].PID
	})

	// Create child nodes and recurse
	for _, child := range children {
		childNode := &ProcessNode{
			Info:   child,
			Parent: node,
		}
		node.Children = append(node.Children, childNode)
		buildTreeRecursive(childNode, processMap)
	}
}

// DisplayTree displays the process tree in a formatted way
func DisplayTree(root *ProcessNode, maxDepth int) {
	fmt.Printf("Process Tree (starting from PID %d):\n", root.Info.PID)
	fmt.Println(strings.Repeat("─", 60))
	displayTreeRecursive(root, "", true, 0, maxDepth)
}

// displayTreeRecursive recursively displays the process tree
func displayTreeRecursive(node *ProcessNode, prefix string, isLast bool, depth int, maxDepth int) {
	// Check depth limit
	if maxDepth > 0 && depth >= maxDepth {
		if len(node.Children) > 0 {
			fmt.Printf("%s%s... (%d children pruned)\n",
				prefix, getConnector(isLast), len(node.Children))
		}
		return
	}

	// Display current node
	connector := getConnector(isLast)
	memMB := float64(node.Info.Memory) / 1024 / 1024

	fmt.Printf("%s%s%s (%d) [%s] %.1fMB (%d threads)\n",
		prefix, connector, node.Info.Name, node.Info.PID,
		node.Info.State, memMB, node.Info.Threads)

	// Prepare prefix for children
	var childPrefix string
	if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + "│   "
	}

	// Display children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		displayTreeRecursive(child, childPrefix, isLastChild, depth+1, maxDepth)
	}
}

// getConnector returns the appropriate tree connector character
func getConnector(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}

// PrintUsage prints usage information
func PrintUsage() {
	fmt.Println("Process Tree Explorer")
	fmt.Println("Usage: go run main.go [PID] [options]")
	fmt.Println("")
	fmt.Println("Arguments:")
	fmt.Println("  PID     Process ID to start tree from (default: 1)")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --depth N    Limit tree depth to N levels")
	fmt.Println("  --help       Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run main.go           # Show full tree from init")
	fmt.Println("  go run main.go 1234      # Show tree from PID 1234")
	fmt.Println("  go run main.go 1 --depth 3  # Limit to 3 levels deep")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: PID argument required")
		PrintUsage()
		os.Exit(1)
	}

	// Check for help flag
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			PrintUsage()
			return
		}
	}

	// Parse PID
	rootPID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error: Invalid PID '%s'\n", os.Args[1])
		PrintUsage()
		os.Exit(1)
	}

	// Parse depth limit
	maxDepth := 0 // 0 means no limit
	for i, arg := range os.Args {
		if arg == "--depth" && i+1 < len(os.Args) {
			if depth, err := strconv.Atoi(os.Args[i+1]); err == nil && depth > 0 {
				maxDepth = depth
			} else {
				fmt.Printf("Error: Invalid depth value '%s'\n", os.Args[i+1])
				os.Exit(1)
			}
			break
		}
	}

	// Build and display the process tree
	tree, err := BuildProcessTree(rootPID)
	if err != nil {
		fmt.Printf("Error building process tree: %v\n", err)
		os.Exit(1)
	}

	DisplayTree(tree, maxDepth)

	// Print summary
	nodeCount := countNodes(tree)
	fmt.Printf("\nTotal processes in tree: %d\n", nodeCount)
	if maxDepth > 0 {
		fmt.Printf("Tree depth limited to: %d levels\n", maxDepth)
	}
}

// countNodes counts the total number of nodes in the tree
func countNodes(node *ProcessNode) int {
	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}
