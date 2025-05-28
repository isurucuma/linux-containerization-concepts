package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// SignalInfo contains signal metadata
type SignalInfo struct {
	Name        string
	Number      int
	Description string
	Catchable   bool
}

// SignalPlayground manages signal operations
type SignalPlayground struct {
	processes map[int]*exec.Cmd
	mutex     sync.RWMutex
	signals   map[int]*SignalInfo
}

// NewSignalPlayground creates a new signal playground
func NewSignalPlayground() *SignalPlayground {
	sp := &SignalPlayground{
		processes: make(map[int]*exec.Cmd),
		signals:   make(map[int]*SignalInfo),
	}
	sp.initializeSignals()
	return sp
}

// initializeSignals initializes the signal information map
func (sp *SignalPlayground) initializeSignals() {
	signals := []*SignalInfo{
		{Name: "SIGHUP", Number: 1, Description: "Hangup (POSIX)", Catchable: true},
		{Name: "SIGINT", Number: 2, Description: "Terminal interrupt (ANSI)", Catchable: true},
		{Name: "SIGQUIT", Number: 3, Description: "Terminal quit (POSIX)", Catchable: true},
		{Name: "SIGILL", Number: 4, Description: "Illegal instruction (ANSI)", Catchable: true},
		{Name: "SIGTRAP", Number: 5, Description: "Trace trap (POSIX)", Catchable: true},
		{Name: "SIGABRT", Number: 6, Description: "Abort (ANSI)", Catchable: true},
		{Name: "SIGBUS", Number: 7, Description: "BUS error (4.2 BSD)", Catchable: true},
		{Name: "SIGFPE", Number: 8, Description: "Floating-point exception (ANSI)", Catchable: true},
		{Name: "SIGKILL", Number: 9, Description: "Kill (can't be caught or ignored) (POSIX)", Catchable: false},
		{Name: "SIGUSR1", Number: 10, Description: "User-defined signal 1 (POSIX)", Catchable: true},
		{Name: "SIGSEGV", Number: 11, Description: "Segmentation violation (ANSI)", Catchable: true},
		{Name: "SIGUSR2", Number: 12, Description: "User-defined signal 2 (POSIX)", Catchable: true},
		{Name: "SIGPIPE", Number: 13, Description: "Write on a pipe with no reader (POSIX)", Catchable: true},
		{Name: "SIGALRM", Number: 14, Description: "Alarm clock (POSIX)", Catchable: true},
		{Name: "SIGTERM", Number: 15, Description: "Termination (ANSI)", Catchable: true},
		{Name: "SIGSTKFLT", Number: 16, Description: "Stack fault", Catchable: true},
		{Name: "SIGCHLD", Number: 17, Description: "Child status has changed (POSIX)", Catchable: true},
		{Name: "SIGCONT", Number: 18, Description: "Continue (POSIX)", Catchable: false},
		{Name: "SIGSTOP", Number: 19, Description: "Stop (can't be caught or ignored) (POSIX)", Catchable: false},
		{Name: "SIGTSTP", Number: 20, Description: "Keyboard stop (POSIX)", Catchable: true},
		{Name: "SIGTTIN", Number: 21, Description: "Background read from tty (POSIX)", Catchable: true},
		{Name: "SIGTTOU", Number: 22, Description: "Background write to tty (POSIX)", Catchable: true},
		{Name: "SIGURG", Number: 23, Description: "Urgent condition on socket (4.2 BSD)", Catchable: true},
		{Name: "SIGXCPU", Number: 24, Description: "CPU limit exceeded (4.2 BSD)", Catchable: true},
		{Name: "SIGXFSZ", Number: 25, Description: "File size limit exceeded (4.2 BSD)", Catchable: true},
		{Name: "SIGVTALRM", Number: 26, Description: "Virtual alarm clock (4.2 BSD)", Catchable: true},
		{Name: "SIGPROF", Number: 27, Description: "Profiling alarm clock (4.2 BSD)", Catchable: true},
		{Name: "SIGWINCH", Number: 28, Description: "Window size change (4.3 BSD, Sun)", Catchable: true},
		{Name: "SIGIO", Number: 29, Description: "I/O now possible (4.2 BSD)", Catchable: true},
		{Name: "SIGPWR", Number: 30, Description: "Power failure restart (System V)", Catchable: true},
		{Name: "SIGSYS", Number: 31, Description: "Bad system call", Catchable: true},
	}

	for _, sig := range signals {
		sp.signals[sig.Number] = sig
	}
}

// ListSignals displays all available signals
func (sp *SignalPlayground) ListSignals() {
	fmt.Println("=== AVAILABLE SIGNALS ===")
	fmt.Printf("%-12s %-6s %-10s %s\n", "NAME", "NUMBER", "CATCHABLE", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", 80))

	for i := 1; i <= 31; i++ {
		if sig, exists := sp.signals[i]; exists {
			catchable := "Yes"
			if !sig.Catchable {
				catchable = "No"
			}
			fmt.Printf("%-12s %-6d %-10s %s\n", sig.Name, sig.Number, catchable, sig.Description)
		}
	}
	fmt.Println()
}

// StartTestProcess starts a test process that handles signals
func (sp *SignalPlayground) StartTestProcess() (int, error) {
	// Create a simple shell script that handles signals
	script := `#!/bin/bash
trap 'echo "Received SIGTERM, exiting gracefully..."; exit 0' TERM
trap 'echo "Received SIGINT, ignoring..."; echo "Use SIGTERM to exit gracefully or SIGKILL to force exit"' INT
trap 'echo "Received SIGUSR1, logging status..."; echo "Process is running normally"' USR1
trap 'echo "Received SIGUSR2, toggling verbose mode..."' USR2
trap 'echo "Received SIGHUP, reloading configuration..."' HUP

echo "Test process started with PID $$"
echo "This process handles the following signals:"
echo "  SIGTERM (15) - Graceful shutdown"
echo "  SIGINT (2)   - Ignored with message"
echo "  SIGUSR1 (10) - Status logging"
echo "  SIGUSR2 (12) - Toggle verbose mode"
echo "  SIGHUP (1)   - Reload configuration"
echo "Process will run for 300 seconds unless terminated..."

counter=0
while [ $counter -lt 300 ]; do
    sleep 1
    counter=$((counter + 1))
    if [ $((counter % 10)) -eq 0 ]; then
        echo "Heartbeat: $counter seconds elapsed"
    fi
done

echo "Test process exiting after 300 seconds"
`

	// Write script to temporary file
	tmpFile := "/tmp/signal_test_process.sh"
	err := os.WriteFile(tmpFile, []byte(script), 0755)
	if err != nil {
		return 0, err
	}

	// Start the process
	cmd := exec.Command("bash", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return 0, err
	}

	pid := cmd.Process.Pid

	sp.mutex.Lock()
	sp.processes[pid] = cmd
	sp.mutex.Unlock()

	fmt.Printf("Started test process with PID: %d\n", pid)
	return pid, nil
}

// SendSignal sends a signal to a process
func (sp *SignalPlayground) SendSignal(pid int, sigNum int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %v", err)
	}

	var signal os.Signal
	switch sigNum {
	case 1:
		signal = syscall.SIGHUP
	case 2:
		signal = syscall.SIGINT
	case 3:
		signal = syscall.SIGQUIT
	case 9:
		signal = syscall.SIGKILL
	case 10:
		signal = syscall.SIGUSR1
	case 12:
		signal = syscall.SIGUSR2
	case 15:
		signal = syscall.SIGTERM
	case 18:
		signal = syscall.SIGCONT
	case 19:
		signal = syscall.SIGSTOP
	case 20:
		signal = syscall.SIGTSTP
	default:
		return fmt.Errorf("signal %d not supported in this implementation", sigNum)
	}

	sigInfo := sp.signals[sigNum]
	if sigInfo != nil {
		fmt.Printf("Sending %s (%d) to process %d: %s\n",
			sigInfo.Name, sigInfo.Number, pid, sigInfo.Description)
	}

	err = process.Signal(signal)
	if err != nil {
		return fmt.Errorf("failed to send signal: %v", err)
	}

	return nil
}

// MonitorProcess monitors a process and displays signal information
func (sp *SignalPlayground) MonitorProcess(pid int) {
	fmt.Printf("=== MONITORING PROCESS %d ===\n", pid)
	fmt.Println("Press Ctrl+C to stop monitoring")

	// Set up signal handling for the monitor itself
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\nStopping process monitoring...")
			return

		case <-ticker.C:
			// Check if process still exists
			if !sp.processExists(pid) {
				fmt.Printf("Process %d has terminated\n", pid)
				return
			}

			// Read process status
			statusPath := fmt.Sprintf("/proc/%d/status", pid)
			data, err := os.ReadFile(statusPath)
			if err != nil {
				fmt.Printf("Process %d no longer accessible\n", pid)
				return
			}

			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "State:") {
					fmt.Printf("[%s] Process %d state: %s\n",
						time.Now().Format("15:04:05"), pid, strings.TrimPrefix(line, "State:\t"))
					break
				}
			}
		}
	}
}

// processExists checks if a process exists
func (sp *SignalPlayground) processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// CleanupProcess cleans up a tracked process
func (sp *SignalPlayground) CleanupProcess(pid int) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if cmd, exists := sp.processes[pid]; exists {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(sp.processes, pid)
	}
}

// InteractiveMode starts interactive mode for signal experimentation
func (sp *SignalPlayground) InteractiveMode() {
	fmt.Println("=== SIGNAL PLAYGROUND - INTERACTIVE MODE ===")
	fmt.Println("Commands:")
	fmt.Println("  list                    - List all signals")
	fmt.Println("  start                   - Start a test process")
	fmt.Println("  send <PID> <SIGNAL>     - Send signal to process")
	fmt.Println("  monitor <PID>           - Monitor process state")
	fmt.Println("  kill <PID>              - Kill test process")
	fmt.Println("  help                    - Show this help")
	fmt.Println("  quit                    - Exit playground")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("signal> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]

		switch command {
		case "list":
			sp.ListSignals()

		case "start":
			pid, err := sp.StartTestProcess()
			if err != nil {
				fmt.Printf("Error starting process: %v\n", err)
			} else {
				fmt.Printf("Started test process with PID: %d\n", pid)
			}

		case "send":
			if len(parts) != 3 {
				fmt.Println("Usage: send <PID> <SIGNAL>")
				continue
			}
			pid, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Invalid PID: %s\n", parts[1])
				continue
			}
			sigNum, err := strconv.Atoi(parts[2])
			if err != nil {
				fmt.Printf("Invalid signal number: %s\n", parts[2])
				continue
			}
			if err := sp.SendSignal(pid, sigNum); err != nil {
				fmt.Printf("Error sending signal: %v\n", err)
			}

		case "monitor":
			if len(parts) != 2 {
				fmt.Println("Usage: monitor <PID>")
				continue
			}
			pid, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Invalid PID: %s\n", parts[1])
				continue
			}
			sp.MonitorProcess(pid)

		case "kill":
			if len(parts) != 2 {
				fmt.Println("Usage: kill <PID>")
				continue
			}
			pid, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Invalid PID: %s\n", parts[1])
				continue
			}
			sp.SendSignal(pid, 9) // SIGKILL
			sp.CleanupProcess(pid)

		case "help":
			fmt.Println("Commands:")
			fmt.Println("  list                    - List all signals")
			fmt.Println("  start                   - Start a test process")
			fmt.Println("  send <PID> <SIGNAL>     - Send signal to process")
			fmt.Println("  monitor <PID>           - Monitor process state")
			fmt.Println("  kill <PID>              - Kill test process")
			fmt.Println("  help                    - Show this help")
			fmt.Println("  quit                    - Exit playground")

		case "quit", "exit":
			fmt.Println("Cleaning up and exiting...")
			sp.cleanup()
			return

		default:
			fmt.Printf("Unknown command: %s (type 'help' for available commands)\n", command)
		}
	}
}

// cleanup cleans up all tracked processes
func (sp *SignalPlayground) cleanup() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	for pid, cmd := range sp.processes {
		if cmd.Process != nil {
			fmt.Printf("Cleaning up process %d\n", pid)
			cmd.Process.Kill()
		}
	}
	sp.processes = make(map[int]*exec.Cmd)
}

// DemoSignalHandling demonstrates various signal handling scenarios
func (sp *SignalPlayground) DemoSignalHandling() {
	fmt.Println("=== SIGNAL HANDLING DEMONSTRATION ===")

	// Start a test process
	fmt.Println("1. Starting test process...")
	pid, err := sp.StartTestProcess()
	if err != nil {
		log.Fatal("Failed to start test process:", err)
	}

	// Wait a moment for process to start
	time.Sleep(2 * time.Second)

	// Demonstrate different signals
	fmt.Println("\n2. Sending SIGUSR1 (status request)...")
	sp.SendSignal(pid, 10)
	time.Sleep(2 * time.Second)

	fmt.Println("\n3. Sending SIGUSR2 (toggle verbose)...")
	sp.SendSignal(pid, 12)
	time.Sleep(2 * time.Second)

	fmt.Println("\n4. Sending SIGHUP (reload config)...")
	sp.SendSignal(pid, 1)
	time.Sleep(2 * time.Second)

	fmt.Println("\n5. Sending SIGINT (should be ignored)...")
	sp.SendSignal(pid, 2)
	time.Sleep(2 * time.Second)

	fmt.Println("\n6. Sending SIGTERM (graceful shutdown)...")
	sp.SendSignal(pid, 15)
	time.Sleep(3 * time.Second)

	// Cleanup
	sp.CleanupProcess(pid)
	fmt.Println("\nDemo completed!")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go interactive              - Start interactive mode")
		fmt.Println("  go run main.go list                     - List all signals")
		fmt.Println("  go run main.go demo                     - Run signal handling demo")
		fmt.Println("  go run main.go send <PID> <SIGNAL>      - Send signal to process")
		fmt.Println("  go run main.go monitor <PID>            - Monitor process state")
		os.Exit(1)
	}

	playground := NewSignalPlayground()

	switch os.Args[1] {
	case "interactive":
		playground.InteractiveMode()

	case "list":
		playground.ListSignals()

	case "demo":
		playground.DemoSignalHandling()

	case "send":
		if len(os.Args) != 4 {
			fmt.Println("Usage: go run main.go send <PID> <SIGNAL>")
			os.Exit(1)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Invalid PID:", err)
		}
		sigNum, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal("Invalid signal number:", err)
		}
		if err := playground.SendSignal(pid, sigNum); err != nil {
			log.Fatal("Error sending signal:", err)
		}

	case "monitor":
		if len(os.Args) != 3 {
			fmt.Println("Usage: go run main.go monitor <PID>")
			os.Exit(1)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Invalid PID:", err)
		}
		playground.MonitorProcess(pid)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
