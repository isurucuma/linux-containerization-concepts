package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProcessDemo demonstrates key process management concepts
type ProcessDemo struct {
	runningProcesses map[int]*exec.Cmd
}

func NewProcessDemo() *ProcessDemo {
	return &ProcessDemo{
		runningProcesses: make(map[int]*exec.Cmd),
	}
}

func main() {
	demo := NewProcessDemo()

	fmt.Println("🚀 Linux Process Management Demo")
	fmt.Println("================================")
	fmt.Println()

	for {
		showMenu()
		choice := getUserInput()

		switch choice {
		case "1":
			demo.showCurrentProcess()
		case "2":
			demo.showProcessTree()
		case "3":
			demo.createChildProcess()
		case "4":
			demo.demonstrateSignals()
		case "5":
			demo.monitorProcess()
		case "6":
			demo.showRunningDemos()
		case "7":
			demo.killDemoProcess()
		case "8":
			demo.cleanupAndExit()
			return
		default:
			fmt.Println("❌ Invalid choice. Please try again.")
		}

		fmt.Println("\nPress Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func showMenu() {
	fmt.Println("\n📋 Process Management Demonstrations:")
	fmt.Println("1. Show Current Process Info")
	fmt.Println("2. Show Process Tree")
	fmt.Println("3. Create Child Process")
	fmt.Println("4. Demonstrate Signal Handling")
	fmt.Println("5. Monitor Process Resources")
	fmt.Println("6. Show Running Demo Processes")
	fmt.Println("7. Kill Demo Process")
	fmt.Println("8. Exit")
	fmt.Print("\nChoose an option (1-8): ")
}

func getUserInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// 1. Show Current Process Information
func (pd *ProcessDemo) showCurrentProcess() {
	fmt.Println("\n🔍 Current Process Information:")
	fmt.Println("==============================")

	pid := os.Getpid()
	ppid := os.Getppid()
	uid := os.Getuid()
	gid := os.Getgid()

	fmt.Printf("Process ID (PID): %d\n", pid)
	fmt.Printf("Parent Process ID (PPID): %d\n", ppid)
	fmt.Printf("User ID (UID): %d\n", uid)
	fmt.Printf("Group ID (GID): %d\n", gid)
	fmt.Printf("Working Directory: %s\n", getWorkingDirectory())
	fmt.Printf("Command Line: %v\n", os.Args)

	// Show process status from /proc
	pd.showProcStatus(pid)
}

func getWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

func (pd *ProcessDemo) showProcStatus(pid int) {
	fmt.Println("\n📊 Process Status from /proc:")

	// Read basic status info
	statusFile := fmt.Sprintf("/proc/%d/status", pid)
	if content, err := os.ReadFile(statusFile); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "State:") ||
				strings.HasPrefix(line, "VmSize:") ||
				strings.HasPrefix(line, "VmRSS:") ||
				strings.HasPrefix(line, "Threads:") {
				fmt.Printf("  %s\n", line)
			}
		}
	}
}

// 2. Show Process Tree
func (pd *ProcessDemo) showProcessTree() {
	fmt.Println("\n🌳 Process Tree (showing our process lineage):")
	fmt.Println("==============================================")

	currentPID := os.Getpid()
	pd.showProcessLineage(currentPID, 0)
}

func (pd *ProcessDemo) showProcessLineage(pid int, depth int) {
	indent := strings.Repeat("  ", depth)
	processName := pd.getProcessName(pid)

	fmt.Printf("%s├─ PID: %d (%s)\n", indent, pid, processName)

	// Get parent process
	ppid := pd.getParentPID(pid)
	if ppid > 1 && depth < 5 { // Limit depth to avoid going too far up
		pd.showProcessLineage(ppid, depth+1)
	}
}

func (pd *ProcessDemo) getProcessName(pid int) string {
	commFile := fmt.Sprintf("/proc/%d/comm", pid)
	if content, err := os.ReadFile(commFile); err == nil {
		return strings.TrimSpace(string(content))
	}
	return "unknown"
}

func (pd *ProcessDemo) getParentPID(pid int) int {
	statFile := fmt.Sprintf("/proc/%d/stat", pid)
	if content, err := os.ReadFile(statFile); err == nil {
		fields := strings.Fields(string(content))
		if len(fields) > 3 {
			if ppid, err := strconv.Atoi(fields[3]); err == nil {
				return ppid
			}
		}
	}
	return -1
}

// 3. Create Child Process
func (pd *ProcessDemo) createChildProcess() {
	fmt.Println("\n👶 Creating Child Process:")
	fmt.Println("==========================")

	// Create a long-running child process
	cmd := exec.Command("sleep", "30")

	fmt.Println("Starting child process: sleep 30")

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to start child process: %v\n", err)
		return
	}

	childPID := cmd.Process.Pid
	pd.runningProcesses[childPID] = cmd

	fmt.Printf("✅ Child process started with PID: %d\n", childPID)
	fmt.Printf("Parent PID: %d\n", os.Getpid())

	// Show the process in action
	fmt.Println("\n📊 Child Process Info:")
	pd.showProcStatus(childPID)

	// Start a goroutine to wait for the process to complete
	go func() {
		err := cmd.Wait()
		delete(pd.runningProcesses, childPID)
		if err != nil {
			fmt.Printf("\n⚠️  Child process %d exited with error: %v\n", childPID, err)
		} else {
			fmt.Printf("\n✅ Child process %d completed successfully\n", childPID)
		}
	}()
}

// 4. Demonstrate Signal Handling
func (pd *ProcessDemo) demonstrateSignals() {
	fmt.Println("\n📡 Signal Handling Demonstration:")
	fmt.Println("=================================")

	// Create a process that handles signals
	cmd := exec.Command("bash", "-c", `
		trap 'echo "Received SIGTERM, gracefully shutting down..."; exit 0' TERM
		trap 'echo "Received SIGINT, ignoring..."; ' INT
		echo "Signal handler process started (PID: $$)"
		echo "Try sending SIGTERM or SIGINT to this process"
		for i in {1..60}; do
			echo "Working... ($i/60)"
			sleep 1
		done
		echo "Process completed normally"
	`)

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to start signal demo process: %v\n", err)
		return
	}

	childPID := cmd.Process.Pid
	pd.runningProcesses[childPID] = cmd

	fmt.Printf("✅ Signal demo process started with PID: %d\n", childPID)
	fmt.Println("\nAvailable signals to send:")
	fmt.Printf("  kill -TERM %d   (graceful shutdown)\n", childPID)
	fmt.Printf("  kill -INT %d    (will be ignored)\n", childPID)
	fmt.Printf("  kill -KILL %d   (force kill)\n", childPID)

	// Monitor the process
	go func() {
		err := cmd.Wait()
		delete(pd.runningProcesses, childPID)
		if err != nil {
			fmt.Printf("\n⚠️  Signal demo process %d exited: %v\n", childPID, err)
		} else {
			fmt.Printf("\n✅ Signal demo process %d completed\n", childPID)
		}
	}()
}

// 5. Monitor Process Resources
func (pd *ProcessDemo) monitorProcess() {
	fmt.Println("\n📈 Process Resource Monitoring:")
	fmt.Println("===============================")

	if len(pd.runningProcesses) == 0 {
		fmt.Println("No running demo processes to monitor.")
		fmt.Println("Try creating a child process first (option 3).")
		return
	}

	fmt.Println("Monitoring resources for 10 seconds...")

	for i := 0; i < 10; i++ {
		fmt.Printf("\n--- Monitoring cycle %d ---\n", i+1)

		for pid := range pd.runningProcesses {
			if pd.processExists(pid) {
				pd.showProcessResources(pid)
			}
		}

		if i < 9 {
			time.Sleep(1 * time.Second)
		}
	}
}

func (pd *ProcessDemo) processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Try to send signal 0 (no-op) to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (pd *ProcessDemo) showProcessResources(pid int) {
	fmt.Printf("\n🔍 PID %d Resources:\n", pid)

	// Memory info from /proc/pid/status
	statusFile := fmt.Sprintf("/proc/%d/status", pid)
	if content, err := os.ReadFile(statusFile); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "VmSize:") ||
				strings.HasPrefix(line, "VmRSS:") {
				fmt.Printf("  %s\n", line)
			}
		}
	}

	// CPU and other stats from /proc/pid/stat
	statFile := fmt.Sprintf("/proc/%d/stat", pid)
	if content, err := os.ReadFile(statFile); err == nil {
		fields := strings.Fields(string(content))
		if len(fields) > 13 {
			fmt.Printf("  CPU User Time: %s jiffies\n", fields[13])
			fmt.Printf("  CPU System Time: %s jiffies\n", fields[14])
		}
	}
}

// 6. Show Running Demo Processes
func (pd *ProcessDemo) showRunningDemos() {
	fmt.Println("\n🏃 Running Demo Processes:")
	fmt.Println("==========================")

	if len(pd.runningProcesses) == 0 {
		fmt.Println("No demo processes are currently running.")
		return
	}

	for pid, cmd := range pd.runningProcesses {
		if pd.processExists(pid) {
			fmt.Printf("✅ PID: %d - Command: %v\n", pid, cmd.Args)
		} else {
			fmt.Printf("❌ PID: %d - Process no longer exists\n", pid)
			delete(pd.runningProcesses, pid)
		}
	}
}

// 7. Kill Demo Process
func (pd *ProcessDemo) killDemoProcess() {
	fmt.Println("\n💀 Kill Demo Process:")
	fmt.Println("=====================")

	if len(pd.runningProcesses) == 0 {
		fmt.Println("No demo processes are currently running.")
		return
	}

	fmt.Println("Running processes:")
	for pid := range pd.runningProcesses {
		if pd.processExists(pid) {
			fmt.Printf("  PID: %d\n", pid)
		}
	}

	fmt.Print("\nEnter PID to kill (or 0 to cancel): ")
	input := getUserInput()

	if input == "0" {
		fmt.Println("Operation cancelled.")
		return
	}

	pid, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("❌ Invalid PID: %s\n", input)
		return
	}

	if cmd, exists := pd.runningProcesses[pid]; exists {
		fmt.Printf("Killing process %d...\n", pid)

		if err := cmd.Process.Kill(); err != nil {
			fmt.Printf("❌ Failed to kill process: %v\n", err)
		} else {
			fmt.Printf("✅ Process %d killed successfully\n", pid)
			delete(pd.runningProcesses, pid)
		}
	} else {
		fmt.Printf("❌ PID %d not found in running demo processes\n", pid)
	}
}

// 8. Cleanup and Exit
func (pd *ProcessDemo) cleanupAndExit() {
	fmt.Println("\n🧹 Cleaning up and exiting...")

	// Kill all running demo processes
	for pid, cmd := range pd.runningProcesses {
		if pd.processExists(pid) {
			fmt.Printf("Terminating process %d...\n", pid)
			cmd.Process.Kill()
		}
	}

	// Setup signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("✅ Cleanup completed. Goodbye!")
}
