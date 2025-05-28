package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// This example demonstrates basic process information gathering
func main() {
	fmt.Println("=== Basic Process Information ===")

	// Current process information
	pid := os.Getpid()
	ppid := os.Getppid()
	uid := os.Getuid()
	gid := os.Getgid()

	fmt.Printf("Process ID (PID): %d\n", pid)
	fmt.Printf("Parent Process ID (PPID): %d\n", ppid)
	fmt.Printf("User ID (UID): %d\n", uid)
	fmt.Printf("Group ID (GID): %d\n", gid)

	// Process group and session information
	pgid := syscall.Getpgrp()
	sid, err := syscall.Getsid(0)
	if err != nil {
		fmt.Printf("Error getting session ID: %v\n", err)
	} else {
		fmt.Printf("Process Group ID (PGID): %d\n", pgid)
		fmt.Printf("Session ID (SID): %d\n", sid)
	}

	// Working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
	} else {
		fmt.Printf("Working Directory: %s\n", wd)
	}

	// Environment variables count
	envCount := len(os.Environ())
	fmt.Printf("Environment Variables: %d\n", envCount)

	// Command line arguments
	fmt.Printf("Command Line Arguments: %v\n", os.Args)

	fmt.Println("\n=== Process Timing ===")
	start := time.Now()

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	elapsed := time.Since(start)
	fmt.Printf("Process has been running for: %v\n", elapsed)

	fmt.Println("\n=== File Descriptors ===")
	// Show standard file descriptors
	fmt.Printf("stdin (fd 0): %v\n", os.Stdin.Name())
	fmt.Printf("stdout (fd 1): %v\n", os.Stdout.Name())
	fmt.Printf("stderr (fd 2): %v\n", os.Stderr.Name())
}
