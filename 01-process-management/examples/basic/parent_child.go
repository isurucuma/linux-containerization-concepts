package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// This example demonstrates parent-child process relationships
func main() {
	fmt.Printf("Parent process PID: %d\n", os.Getpid())

	// Create multiple child processes
	var wg sync.WaitGroup
	numChildren := 3

	for i := 0; i < numChildren; i++ {
		wg.Add(1)
		go func(childNum int) {
			defer wg.Done()
			createChildProcess(childNum)
		}(i + 1)
	}

	// Wait for all children to complete
	wg.Wait()
	fmt.Println("All child processes completed")
}

func createChildProcess(childNum int) {
	fmt.Printf("Creating child process %d...\n", childNum)

	// Create a child process that runs for a few seconds
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`
		echo "Child %d started with PID $$"
		echo "Child %d parent PID: $PPID"
		sleep %d
		echo "Child %d finished"
	`, childNum, childNum, childNum, childNum))

	// Set up process attributes
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}

	// Start the process
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting child %d: %v\n", childNum, err)
		return
	}

	fmt.Printf("Child %d started with PID: %d\n", childNum, cmd.Process.Pid)

	// Monitor the child process
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("Child %d exited with error: %v\n", childNum, err)
		} else {
			fmt.Printf("Child %d completed successfully\n", childNum)
		}
	case <-time.After(10 * time.Second):
		fmt.Printf("Child %d timed out, killing...\n", childNum)
		cmd.Process.Kill()
		<-done // Wait for the process to actually exit
	}
}
