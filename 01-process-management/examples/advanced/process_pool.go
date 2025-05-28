package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ProcessPool manages a pool of worker processes
type ProcessPool struct {
	workers    map[int]*WorkerProcess
	mutex      sync.RWMutex
	maxWorkers int
	taskQueue  chan Task
	shutdown   chan bool
}

// WorkerProcess represents a worker in the pool
type WorkerProcess struct {
	PID       int
	StartTime time.Time
	TaskCount int
	Status    string
}

// Task represents work to be done
type Task struct {
	ID       int
	Command  string
	Duration time.Duration
}

// NewProcessPool creates a new process pool
func NewProcessPool(maxWorkers int) *ProcessPool {
	return &ProcessPool{
		workers:    make(map[int]*WorkerProcess),
		maxWorkers: maxWorkers,
		taskQueue:  make(chan Task, 100),
		shutdown:   make(chan bool),
	}
}

// Start starts the process pool
func (pp *ProcessPool) Start() {
	fmt.Printf("Starting process pool with %d workers\n", pp.maxWorkers)

	// Start worker processes
	for i := 0; i < pp.maxWorkers; i++ {
		pp.spawnWorker(i)
	}

	// Start task dispatcher
	go pp.taskDispatcher()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, cleaning up...")
		pp.Shutdown()
	}()
}

// spawnWorker spawns a new worker process
func (pp *ProcessPool) spawnWorker(workerID int) {
	// For this example, we'll simulate worker processes using goroutines
	// In a real implementation, you'd fork actual processes
	go func() {
		worker := &WorkerProcess{
			PID:       os.Getpid()*1000 + workerID, // Simulate PID
			StartTime: time.Now(),
			TaskCount: 0,
			Status:    "idle",
		}

		pp.mutex.Lock()
		pp.workers[worker.PID] = worker
		pp.mutex.Unlock()

		fmt.Printf("Worker %d (PID: %d) started\n", workerID, worker.PID)

		for {
			select {
			case task := <-pp.taskQueue:
				pp.processTask(worker, task)
			case <-pp.shutdown:
				fmt.Printf("Worker %d (PID: %d) shutting down\n", workerID, worker.PID)
				return
			}
		}
	}()
}

// processTask processes a task
func (pp *ProcessPool) processTask(worker *WorkerProcess, task Task) {
	pp.mutex.Lock()
	worker.Status = "working"
	worker.TaskCount++
	pp.mutex.Unlock()

	fmt.Printf("Worker %d processing task %d: %s\n", worker.PID, task.ID, task.Command)

	// Simulate work
	time.Sleep(task.Duration)

	pp.mutex.Lock()
	worker.Status = "idle"
	pp.mutex.Unlock()

	fmt.Printf("Worker %d completed task %d\n", worker.PID, task.ID)
}

// taskDispatcher dispatches tasks to workers
func (pp *ProcessPool) taskDispatcher() {
	taskID := 1
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	tasks := []string{
		"data_processing",
		"file_compression",
		"image_resize",
		"log_analysis",
		"backup_creation",
		"data_validation",
		"report_generation",
		"cleanup_operation",
	}

	for {
		select {
		case <-ticker.C:
			task := Task{
				ID:       taskID,
				Command:  tasks[taskID%len(tasks)],
				Duration: time.Duration(1+taskID%5) * time.Second,
			}

			select {
			case pp.taskQueue <- task:
				fmt.Printf("Queued task %d: %s\n", task.ID, task.Command)
				taskID++
			default:
				fmt.Println("Task queue full, skipping task")
			}
		case <-pp.shutdown:
			return
		}
	}
}

// GetStatus returns the current status of all workers
func (pp *ProcessPool) GetStatus() {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()

	fmt.Println("\n=== PROCESS POOL STATUS ===")
	fmt.Printf("%-8s %-10s %-12s %-10s %s\n", "PID", "STATUS", "UPTIME", "TASKS", "START_TIME")
	fmt.Println(strings.Repeat("-", 60))

	for pid, worker := range pp.workers {
		uptime := time.Since(worker.StartTime)
		fmt.Printf("%-8d %-10s %-12s %-10d %s\n",
			pid,
			worker.Status,
			formatDuration(uptime),
			worker.TaskCount,
			worker.StartTime.Format("15:04:05"))
	}
	fmt.Printf("\nQueue length: %d\n", len(pp.taskQueue))
}

// MonitorPool monitors the process pool
func (pp *ProcessPool) MonitorPool() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pp.GetStatus()
		case <-pp.shutdown:
			return
		}
	}
}

// Shutdown gracefully shuts down the process pool
func (pp *ProcessPool) Shutdown() {
	close(pp.shutdown)

	// Wait for workers to finish current tasks
	fmt.Println("Waiting for workers to complete current tasks...")
	time.Sleep(2 * time.Second)

	fmt.Println("Process pool shutdown complete")
}

// formatDuration formats a duration in human readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// InteractiveMode provides interactive control of the process pool
func (pp *ProcessPool) InteractiveMode() {
	fmt.Println("=== PROCESS POOL INTERACTIVE MODE ===")
	fmt.Println("Commands:")
	fmt.Println("  status  - Show worker status")
	fmt.Println("  add     - Add task to queue")
	fmt.Println("  quit    - Shutdown and exit")

	go pp.MonitorPool()

	for {
		var input string
		fmt.Print("pool> ")
		fmt.Scanln(&input)

		switch input {
		case "status":
			pp.GetStatus()
		case "add":
			task := Task{
				ID:       int(time.Now().Unix()),
				Command:  "manual_task",
				Duration: 3 * time.Second,
			}
			pp.taskQueue <- task
			fmt.Printf("Added manual task %d\n", task.ID)
		case "quit":
			pp.Shutdown()
			return
		default:
			fmt.Printf("Unknown command: %s\n", input)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go start [workers]    - Start process pool")
		fmt.Println("  go run main.go interactive        - Interactive mode")
		os.Exit(1)
	}

	maxWorkers := 3
	if len(os.Args) > 2 {
		if w, err := strconv.Atoi(os.Args[2]); err == nil && w > 0 {
			maxWorkers = w
		}
	}

	pool := NewProcessPool(maxWorkers)

	switch os.Args[1] {
	case "start":
		pool.Start()
		// Keep running until shutdown
		select {}

	case "interactive":
		pool.Start()
		pool.InteractiveMode()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
