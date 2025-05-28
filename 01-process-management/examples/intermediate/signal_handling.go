package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// This example demonstrates comprehensive signal handling
func main() {
	fmt.Printf("Signal handler demo started (PID: %d)\n", os.Getpid())
	fmt.Println("Try sending signals from another terminal:")
	fmt.Printf("  kill -SIGUSR1 %d\n", os.Getpid())
	fmt.Printf("  kill -SIGUSR2 %d\n", os.Getpid())
	fmt.Printf("  kill -SIGTERM %d\n", os.Getpid())
	fmt.Printf("  kill -SIGINT %d  (or press Ctrl+C)\n", os.Getpid())

	// Create channels for different signal types
	sigChan := make(chan os.Signal, 1)
	userSigChan := make(chan os.Signal, 1)
	termSigChan := make(chan os.Signal, 1)

	// Register for user-defined signals
	signal.Notify(userSigChan, syscall.SIGUSR1, syscall.SIGUSR2)

	// Register for termination signals
	signal.Notify(termSigChan, syscall.SIGINT, syscall.SIGTERM)

	// Register for other signals
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT)

	// Counter for user signals
	userSignalCount := 0

	// Main event loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case sig := <-userSigChan:
			userSignalCount++
			switch sig {
			case syscall.SIGUSR1:
				fmt.Printf("[%s] Received SIGUSR1 - Custom action 1 executed (count: %d)\n",
					time.Now().Format("15:04:05"), userSignalCount)
			case syscall.SIGUSR2:
				fmt.Printf("[%s] Received SIGUSR2 - Custom action 2 executed (count: %d)\n",
					time.Now().Format("15:04:05"), userSignalCount)
			}

		case sig := <-termSigChan:
			switch sig {
			case syscall.SIGINT:
				fmt.Printf("\n[%s] Received SIGINT (Ctrl+C) - Graceful shutdown initiated\n",
					time.Now().Format("15:04:05"))
			case syscall.SIGTERM:
				fmt.Printf("\n[%s] Received SIGTERM - Termination requested\n",
					time.Now().Format("15:04:05"))
			}
			fmt.Println("Cleaning up resources...")
			fmt.Printf("Total user signals received: %d\n", userSignalCount)
			fmt.Println("Goodbye!")
			return

		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				fmt.Printf("[%s] Received SIGHUP - Reloading configuration\n",
					time.Now().Format("15:04:05"))
			case syscall.SIGQUIT:
				fmt.Printf("[%s] Received SIGQUIT - Quit signal\n",
					time.Now().Format("15:04:05"))
				return
			}

		case <-ticker.C:
			fmt.Printf("[%s] Process is running... (user signals: %d)\n",
				time.Now().Format("15:04:05"), userSignalCount)
		}
	}
}
