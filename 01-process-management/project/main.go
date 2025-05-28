package main

import (
	"fmt"
	"os"
	"strconv"

	"process-manager/process"
	"process-manager/signal"
	"process-manager/ui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		handleList()
	case "tree":
		handleTree()
	case "monitor":
		handleMonitor()
	case "signal":
		handleSignal()
	case "start":
		handleStart()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleList() {
	processes, err := process.GetAllProcesses()
	if err != nil {
		fmt.Printf("Error getting processes: %v\n", err)
		os.Exit(1)
	}

	ui.DisplayProcessList(processes)
}

func handleTree() {
	var rootPID int = 1

	if len(os.Args) > 2 {
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid PID: %s\n", os.Args[2])
			os.Exit(1)
		}
		rootPID = pid
	}

	tree, err := process.BuildProcessTree(rootPID)
	if err != nil {
		fmt.Printf("Error building process tree: %v\n", err)
		os.Exit(1)
	}

	ui.DisplayProcessTree(tree)
}

func handleMonitor() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: process-manager monitor <pid>")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Invalid PID: %s\n", os.Args[2])
		os.Exit(1)
	}

	monitor := process.NewMonitor(pid)
	monitor.Start()
}

func handleSignal() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: process-manager signal <pid> <signal>")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Invalid PID: %s\n", os.Args[2])
		os.Exit(1)
	}

	signalName := os.Args[3]

	err = signal.SendSignal(pid, signalName)
	if err != nil {
		fmt.Printf("Error sending signal: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sent %s to process %d\n", signalName, pid)
}

func handleStart() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: process-manager start <command> [args...]")
		os.Exit(1)
	}

	command := os.Args[2]
	args := []string{}
	if len(os.Args) > 3 {
		args = os.Args[3:]
	}

	starter := process.NewStarter()
	pid, err := starter.StartProcess(command, args)
	if err != nil {
		fmt.Printf("Error starting process: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Started process %s with PID %d\n", command, pid)
}

func printUsage() {
	fmt.Println("Process Manager - Linux Container Learning Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  process-manager list                    - List all processes")
	fmt.Println("  process-manager tree [pid]              - Show process tree (default: from init)")
	fmt.Println("  process-manager monitor <pid>           - Monitor process resources")
	fmt.Println("  process-manager signal <pid> <signal>   - Send signal to process")
	fmt.Println("  process-manager start <cmd> [args...]   - Start new process")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  process-manager list")
	fmt.Println("  process-manager tree 1234")
	fmt.Println("  process-manager monitor 1234")
	fmt.Println("  process-manager signal 1234 SIGTERM")
	fmt.Println("  process-manager start sleep 10")
}
