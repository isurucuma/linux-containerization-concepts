package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

// AdvancedProcessDemo demonstrates advanced process management concepts
func main() {
	fmt.Println("=== Advanced Process Management Demo ===")

	// Setup signal handling for clean shutdown
	setupSignalHandling()

	fmt.Println("\n1. Process Group Management")
	demonstrateProcessGroups()

	fmt.Println("\n2. Process Resource Limits")
	demonstrateResourceLimits()

	fmt.Println("\n3. Process Attributes and Security")
	demonstrateProcessSecurity()

	fmt.Println("\n4. Advanced Fork and Exec")
	demonstrateAdvancedForkExec()

	fmt.Println("\nDemo completed. Press Ctrl+C to exit or wait 30 seconds...")
	time.Sleep(30 * time.Second)
}

func setupSignalHandling() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nReceived signal, cleaning up...")
		os.Exit(0)
	}()
}

func demonstrateProcessGroups() {
	fmt.Printf("Current Process Group ID: %d\n", syscall.Getpgrp())
	fmt.Printf("Current Session ID: %d\n", getSessionID())

	// Create a new process group
	if err := syscall.Setpgid(0, 0); err != nil {
		fmt.Printf("Failed to create new process group: %v\n", err)
	} else {
		fmt.Printf("Created new process group: %d\n", syscall.Getpgrp())
	}
}

func getSessionID() int {
	sid, err := syscall.Getsid(0)
	if err != nil {
		return -1
	}
	return sid
}

func demonstrateResourceLimits() {
	// Get current resource limits
	var rlimit syscall.Rlimit

	// Memory limit
	if err := syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit); err == nil {
		fmt.Printf("Virtual Memory Limit: Soft=%d, Hard=%d\n", rlimit.Cur, rlimit.Max)
	}

	// File descriptor limit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err == nil {
		fmt.Printf("File Descriptor Limit: Soft=%d, Hard=%d\n", rlimit.Cur, rlimit.Max)
	}

	// CPU time limit
	if err := syscall.Getrlimit(syscall.RLIMIT_CPU, &rlimit); err == nil {
		fmt.Printf("CPU Time Limit: Soft=%d, Hard=%d seconds\n", rlimit.Cur, rlimit.Max)
	}

	// Process limit
	if err := syscall.Getrlimit(syscall.RLIMIT_NPROC, &rlimit); err == nil {
		fmt.Printf("Process Limit: Soft=%d, Hard=%d\n", rlimit.Cur, rlimit.Max)
	}
}

func demonstrateProcessSecurity() {
	fmt.Printf("Real UID: %d, Effective UID: %d\n", syscall.Getuid(), syscall.Geteuid())
	fmt.Printf("Real GID: %d, Effective GID: %d\n", syscall.Getgid(), syscall.Getegid())

	// Get supplementary groups
	groups, err := syscall.Getgroups()
	if err == nil {
		fmt.Printf("Supplementary Groups: %v\n", groups)
	}

	// Process priority (nice value)
	priority, err := syscall.Getpriority(syscall.PRIO_PROCESS, 0)
	if err == nil {
		fmt.Printf("Process Priority (nice): %d\n", priority)
	}
}

func demonstrateAdvancedForkExec() {
	fmt.Println("Creating child process with custom attributes...")

	// Create command with custom process attributes
	cmd := exec.Command("sleep", "5")

	// Set custom process attributes
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Create new process group
		Setpgid: true,
		Pgid:    0,
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start child process: %v\n", err)
		return
	}

	fmt.Printf("Child process started with PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Child process group: %d\n", getProcessGroup(cmd.Process.Pid))

	// Wait for child to complete
	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("Child process failed: %v\n", err)
		} else {
			fmt.Printf("Child process completed successfully\n")
		}
	}()

	// Give child some time to run
	time.Sleep(2 * time.Second)

	// Send signal to child process group
	fmt.Printf("Sending SIGTERM to child process group...\n")
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
		fmt.Printf("Failed to send signal to process group: %v\n", err)
	}
}

func getProcessGroup(pid int) int {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return -1
	}
	return pgid
}
