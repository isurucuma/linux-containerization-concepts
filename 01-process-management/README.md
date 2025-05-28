# Section 1: Foundation - Linux Process Management

## 🎯 Learning Objectives

By the end of this section, you will:

- Understand the Linux process lifecycle and hierarchy
- Master process creation and management in Go
- Implement proper signal handling
- Build tools to monitor and control processes
- Understand the foundation for container process isolation

## 📚 Theory and Concepts

### What is a Process?

A process is an instance of a running program. In Linux, every process has:

- **Process ID (PID)**: Unique identifier
- **Parent Process ID (PPID)**: The process that created it
- **Process State**: Running, sleeping, zombie, etc.
- **Memory Space**: Virtual memory allocated to the process
- **File Descriptors**: Open files and I/O streams
- **Environment Variables**: Configuration data
- **Signal Handlers**: Response to system signals

**[PLACEHOLDER: Process Memory Layout Diagram]**
_This diagram should show the memory layout of a Linux process including:_

- _Text segment (code)_
- _Data segment (initialized data)_
- _BSS segment (uninitialized data)_
- _Heap (dynamic allocation)_
- _Stack (function calls and local variables)_
- _Memory mapped regions_

### Process Lifecycle

1. **Creation**: Process created via `fork()` or `clone()`
2. **Execution**: Process runs and may create child processes
3. **Termination**: Process ends normally or via signal
4. **Cleanup**: Parent process reads exit status, resources freed

**[PLACEHOLDER: Process Lifecycle State Diagram]**
_This diagram should show the process states and transitions:_

- _Running → Sleeping (waiting for I/O)_
- _Running → Stopped (SIGSTOP signal)_
- _Running → Zombie (terminated, waiting for parent)_
- _Sleeping → Running (I/O completed)_
- _Stopped → Running (SIGCONT signal)_

### Process Hierarchy

Linux processes form a tree structure:

- **init** (PID 1): Root of all processes
- **Parent-Child Relationships**: Each process (except init) has a parent
- **Process Groups**: Related processes grouped together
- **Sessions**: Groups of process groups

**[PLACEHOLDER: Process Tree Diagram]**
_This diagram should show a typical Linux process tree:_

```
init (1)
├── systemd --user (234)
├── kthreadd (2)
│   ├── ksoftirqd/0 (3)
│   └── migration/0 (4)
├── bash (1234)
│   ├── vim (1456)
│   └── go run main.go (1678)
│       └── main (1679)
└── sshd (567)
    └── sshd: user@pts/0 (890)
        └── bash (891)
```

### System Calls for Process Management

#### fork() - Process Creation

```go
// Creates an exact copy of the current process
// Returns PID of child to parent, 0 to child
pid := syscall.ForkExec(path, args, &syscall.ProcAttr{})
```

#### exec() - Process Replacement

```go
// Replaces current process image with new program
syscall.Exec(path, args, env)
```

#### wait() - Process Synchronization

```go
// Parent waits for child process to terminate
var status syscall.WaitStatus
pid, err := syscall.Wait4(-1, &status, 0, nil)
```

### Signals in Linux

Signals are software interrupts used for inter-process communication:

**Common Signals:**

- `SIGTERM` (15): Polite termination request
- `SIGKILL` (9): Forceful termination (cannot be caught)
- `SIGINT` (2): Interrupt from keyboard (Ctrl+C)
- `SIGSTOP` (19): Stop process (cannot be caught)
- `SIGCONT` (18): Continue stopped process
- `SIGCHLD` (17): Child process terminated or stopped

**[PLACEHOLDER: Signal Flow Diagram]**
_This diagram should show:_

- _Process A sending signal to Process B_
- _Kernel delivering the signal_
- _Process B's signal handler being invoked_
- _Default actions vs custom handlers_

### Go Process Management

Go provides excellent process management capabilities through:

- `os/exec` package: Running external commands
- `os` package: Process information and signals
- `syscall` package: Low-level system calls
- `golang.org/x/sys/unix`: Extended system call support

## 🔬 Practical Examples

### Example 1: Basic Process Information

```go
package main

import (
    "fmt"
    "os"
    "syscall"
)

func main() {
    // Get current process information
    pid := os.Getpid()
    ppid := os.Getppid()
    uid := os.Getuid()
    gid := os.Getgid()

    fmt.Printf("Process Information:\n")
    fmt.Printf("PID: %d\n", pid)
    fmt.Printf("PPID: %d\n", ppid)
    fmt.Printf("UID: %d\n", uid)
    fmt.Printf("GID: %d\n", gid)

    // Get process groups
    pgid := syscall.Getpgrp()
    sid, _ := syscall.Getsid(0)

    fmt.Printf("PGID: %d\n", pgid)
    fmt.Printf("SID: %d\n", sid)
}
```

### Example 2: Creating Child Processes

```go
package main

import (
    "fmt"
    "os"
    "os/exec"
    "syscall"
    "time"
)

func main() {
    fmt.Println("Parent process starting...")

    // Create child process
    cmd := exec.Command("sleep", "5")
    err := cmd.Start()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Child process created with PID: %d\n", cmd.Process.Pid)

    // Parent continues working
    for i := 0; i < 3; i++ {
        fmt.Printf("Parent working... %d\n", i+1)
        time.Sleep(1 * time.Second)
    }

    // Wait for child to complete
    err = cmd.Wait()
    if err != nil {
        fmt.Printf("Child process failed: %v\n", err)
    } else {
        fmt.Println("Child process completed successfully")
    }
}
```

### Example 3: Signal Handling

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    // Create channel to receive signals
    sigChan := make(chan os.Signal, 1)

    // Register signals we want to handle
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

    fmt.Printf("Process %d is running. Send signals to test:\n", os.Getpid())
    fmt.Println("- SIGINT (Ctrl+C): Graceful shutdown")
    fmt.Println("- SIGTERM: Termination request")
    fmt.Println("- SIGUSR1: Custom signal")

    // Main loop
    for {
        select {
        case sig := <-sigChan:
            switch sig {
            case syscall.SIGINT:
                fmt.Println("\nReceived SIGINT - Graceful shutdown")
                return
            case syscall.SIGTERM:
                fmt.Println("\nReceived SIGTERM - Terminating")
                return
            case syscall.SIGUSR1:
                fmt.Println("\nReceived SIGUSR1 - Custom action performed")
            }
        default:
            fmt.Print(".")
            time.Sleep(1 * time.Second)
        }
    }
}
```

## 🛠️ Hands-on Exercises

### Exercise 1: Process Tree Explorer

Create a program that displays the process tree starting from a given PID.

**Requirements:**

- Read process information from `/proc/<pid>/stat`
- Display parent-child relationships
- Show process names and states
- Format output as a tree structure

### Exercise 2: Simple Process Monitor

Build a process monitor that tracks resource usage.

**Requirements:**

- Monitor CPU and memory usage
- Track process lifetime
- Alert on high resource usage
- Log process events

### Exercise 3: Signal Playground

Create a program that demonstrates various signal behaviors.

**Requirements:**

- Send different signals between processes
- Show signal masking and ignoring
- Demonstrate signal handlers
- Show signal inheritance in child processes

## 🎯 Section Project: Process Manager Tool

Build a comprehensive process manager that can:

1. **List Processes**: Show running processes with details
2. **Process Tree**: Display hierarchical process relationships
3. **Start Processes**: Launch new processes with monitoring
4. **Signal Management**: Send signals to processes
5. **Resource Monitoring**: Track CPU, memory, and I/O usage

### Project Structure

```
project/
├── main.go              # CLI interface
├── process/
│   ├── manager.go       # Core process management
│   ├── monitor.go       # Resource monitoring
│   └── tree.go          # Process tree operations
├── signal/
│   └── handler.go       # Signal management
└── ui/
    └── display.go       # Output formatting
```

### Core Features

#### 1. Process Information Display

```go
type ProcessInfo struct {
    PID     int
    PPID    int
    Name    string
    State   string
    CPU     float64
    Memory  uint64
    Threads int
}
```

#### 2. Process Tree Visualization

```go
type ProcessNode struct {
    Info     ProcessInfo
    Children []*ProcessNode
    Parent   *ProcessNode
}
```

#### 3. Signal Management

```go
type SignalManager struct {
    handlers map[os.Signal]func()
    signals  chan os.Signal
}
```

### Expected Output

```bash
$ go run main.go list
PID    PPID   NAME           STATE    CPU%   MEM(MB)  THREADS
1      0      systemd        S        0.1    8.2      1
123    1      bash           S        0.0    2.1      1
456    123    process-mgr    R        1.2    15.6     3

$ go run main.go tree 123
bash (123)
├── vim (789)
└── process-mgr (456)
    ├── worker-1 (457)
    └── worker-2 (458)

$ go run main.go monitor 456
Process 456 (process-mgr):
Time: 14:30:15  CPU: 1.2%  Memory: 15.6MB  State: Running
Time: 14:30:16  CPU: 0.8%  Memory: 15.8MB  State: Running
Time: 14:30:17  CPU: 1.5%  Memory: 16.1MB  State: Running

$ go run main.go signal 456 SIGUSR1
Sent SIGUSR1 to process 456
```

## 🧪 Testing Your Understanding

### Quiz Questions

1. **What happens when a parent process terminates before its children?**

   - Answer: Children become orphans and are adopted by init (PID 1)

2. **What is a zombie process and how is it cleaned up?**

   - Answer: A terminated process whose exit status hasn't been read by its parent. Cleaned up by parent calling wait()

3. **Which signal cannot be caught or ignored?**

   - Answer: SIGKILL (9) and SIGSTOP (19)

4. **What is the difference between fork() and exec()?**
   - Answer: fork() creates a copy of the current process, exec() replaces the current process with a new program

### Practical Verification

Run these commands to verify your understanding:

```bash
# Show process tree
pstree -p

# Monitor process creation
sudo strace -f -e trace=clone,fork,vfork your_program

# Send signals to your process
kill -SIGUSR1 <pid>
kill -SIGTERM <pid>

# Check process information
cat /proc/<pid>/stat
cat /proc/<pid>/status
```

## 🔍 Deep Dive Topics

### Advanced Process Concepts

1. **Process Groups and Sessions**

   - Job control in shells
   - Terminal association
   - Signal delivery to groups

2. **Copy-on-Write (CoW)**

   - Memory efficiency in fork()
   - Page sharing between processes
   - Performance implications

3. **Process Scheduling**
   - CFS (Completely Fair Scheduler)
   - Nice values and priorities
   - Real-time scheduling classes

### Security Considerations

1. **Process Privileges**

   - Real, effective, and saved user IDs
   - Capability-based security
   - Privilege escalation prevention

2. **Process Isolation**
   - Address space isolation
   - Resource limits
   - Security boundaries

## 📖 Additional Reading

- `man 2 fork` - Process creation
- `man 2 exec` - Process replacement
- `man 2 wait` - Process synchronization
- `man 7 signal` - Signal overview
- `/proc/[pid]/` - Process information filesystem

## ✅ Section Completion Checklist

- [ ] Understand Linux process lifecycle
- [ ] Can create and manage processes in Go
- [ ] Implemented proper signal handling
- [ ] Built process monitoring tools
- [ ] Completed the process manager project
- [ ] Understand process security basics

## 🚀 Next Steps

Congratulations! You've mastered the fundamentals of Linux process management. You now understand:

- How processes are created and managed
- Process hierarchies and relationships
- Signal handling and inter-process communication
- Building process management tools in Go

In **Section 2: Namespaces**, we'll build upon this foundation to learn how to create isolated environments for processes - the cornerstone of containerization!
