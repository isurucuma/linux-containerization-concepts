# Exercise 3: Signal Playground

## Objective

Build a comprehensive signal demonstration program that shows how different signals work, how they can be handled, masked, and inherited by child processes.

## Requirements

### Basic Requirements

1. Create a program that can send and receive various signals
2. Demonstrate signal handlers vs default actions
3. Show signal masking and blocking
4. Demonstrate signal inheritance in child processes

### Advanced Requirements

1. Create a signal proxy that forwards signals between processes
2. Implement custom signal handling with context preservation
3. Show signal queuing and real-time signals
4. Create a signal-based inter-process communication system

## Implementation Guidelines

### Step 1: Signal Handler Structure

```go
type SignalHandler struct {
    PID         int
    Handlers    map[os.Signal]func(os.Signal)
    Masked      []os.Signal
    Blocked     []os.Signal
    Received    []SignalEvent
}

type SignalEvent struct {
    Signal    os.Signal
    Timestamp time.Time
    Sender    int
    Handled   bool
}
```

### Step 2: Core Functions to Implement

```go
// SetupSignalHandlers configures custom signal handlers
func (sh *SignalHandler) SetupSignalHandlers() {
    // TODO: Setup handlers for various signals
    // TODO: Use signal.Notify() appropriately
}

// SendSignalToPID sends a signal to another process
func SendSignalToPID(pid int, sig os.Signal) error {
    // TODO: Use os.FindProcess and Signal() method
}

// MaskSignals temporarily blocks signals
func (sh *SignalHandler) MaskSignals(signals []os.Signal) {
    // TODO: Use signal.Stop() to mask signals
}

// CreateChildWithSignalHandling forks a child with custom signal setup
func CreateChildWithSignalHandling() (int, error) {
    // TODO: Create child process with inherited/custom signal handling
}
```

## Demo Scenarios

### Scenario 1: Basic Signal Handling

```bash
$ go run main.go demo basic
Starting signal demo process (PID: 12345)
Setting up handlers for SIGINT, SIGTERM, SIGUSR1, SIGUSR2

Send signals to this process:
- kill -INT 12345   (SIGINT - Interrupt)
- kill -TERM 12345  (SIGTERM - Terminate)
- kill -USR1 12345  (SIGUSR1 - User signal 1)
- kill -USR2 12345  (SIGUSR2 - User signal 2)

Press Ctrl+C or send SIGTERM to exit gracefully...

[14:30:15] Received SIGUSR1 - Custom action performed
[14:30:18] Received SIGINT - Preparing for graceful shutdown
[14:30:20] Received SIGTERM - Terminating now
```

### Scenario 2: Signal Masking

```bash
$ go run main.go demo masking
Demonstrating signal masking...

Phase 1: Normal signal handling
- SIGINT will be handled normally

Phase 2: Masking SIGINT for 5 seconds
- SIGINT will be blocked and queued
- Send SIGINT now - it will be delayed

Phase 3: Unmasking SIGINT
- Queued SIGINT will now be delivered

[14:31:00] Masking SIGINT...
[14:31:02] SIGINT blocked (queued)
[14:31:05] Unmasking SIGINT...
[14:31:05] Received queued SIGINT - Interrupt handled
```

### Scenario 3: Parent-Child Signal Communication

```bash
$ go run main.go demo parent-child
Parent process (PID: 12345) creating child...
Child process (PID: 12346) started

Parent will send signals to child every 2 seconds:
- SIGUSR1: Increment counter
- SIGUSR2: Print status
- SIGTERM: Terminate child

Child signal handlers:
- SIGUSR1: Counter = 1
- SIGUSR1: Counter = 2
- SIGUSR2: Status - Counter: 2, Uptime: 4s
- SIGUSR1: Counter = 3
- SIGTERM: Child terminating gracefully
```

## Advanced Features

### 1. Signal Proxy

Create a program that acts as a signal relay between processes:

```go
type SignalProxy struct {
    SourcePID int
    TargetPID int
    Filter    map[os.Signal]bool
    Log       []ProxyEvent
}
```

### 2. Signal-Based IPC

Implement a simple message system using signals:

```go
// Use SIGUSR1/SIGUSR2 with shared memory or files
// to implement basic message passing
```

### 3. Real-time Signals

Demonstrate Linux real-time signals (SIGRTMIN to SIGRTMAX):

```go
// Show signal queuing and priority
// Demonstrate signal data payload (siginfo_t)
```

## Testing Instructions

### Test 1: Signal Handler Verification

1. Start the signal demo program
2. Send various signals from another terminal
3. Verify correct handling and logging
4. Test with rapid signal sending

### Test 2: Signal Masking

1. Start masking demo
2. Send signals during masked period
3. Verify signals are queued and delivered after unmasking
4. Test with multiple signal types

### Test 3: Child Process Inheritance

1. Start parent-child demo
2. Send signals to both parent and child
3. Verify inheritance and independent handling
4. Test signal forwarding from parent to child

### Test 4: Signal Race Conditions

1. Send signals rapidly from multiple sources
2. Verify all signals are handled correctly
3. Test with concurrent signal handlers

## Implementation Files

```
signal-playground/
├── main.go              # CLI interface and demo modes
├── handler.go           # Signal handler implementation
├── masking.go           # Signal masking and blocking
├── proxy.go             # Signal proxy functionality
├── parent_child.go      # Parent-child signal communication
├── realtime.go          # Real-time signal demonstration
└── utils.go             # Utility functions
```

## Sample Implementation Snippets

### Basic Signal Handler

```go
func setupBasicHandlers() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGUSR1,
        syscall.SIGUSR2,
    )

    go func() {
        for sig := range sigChan {
            timestamp := time.Now().Format("15:04:05")
            switch sig {
            case syscall.SIGINT:
                fmt.Printf("[%s] Received SIGINT - Interrupt\n", timestamp)
            case syscall.SIGTERM:
                fmt.Printf("[%s] Received SIGTERM - Terminate\n", timestamp)
                os.Exit(0)
            case syscall.SIGUSR1:
                fmt.Printf("[%s] Received SIGUSR1 - User signal 1\n", timestamp)
            case syscall.SIGUSR2:
                fmt.Printf("[%s] Received SIGUSR2 - User signal 2\n", timestamp)
            }
        }
    }()
}
```

### Signal Masking

```go
func demonstrateSignalMasking() {
    // Setup signal channel
    sigChan := make(chan os.Signal, 10)
    signal.Notify(sigChan, syscall.SIGINT)

    fmt.Println("Phase 1: Normal handling")
    time.Sleep(3 * time.Second)

    fmt.Println("Phase 2: Masking SIGINT")
    signal.Stop(sigChan)
    time.Sleep(5 * time.Second)

    fmt.Println("Phase 3: Unmasking SIGINT")
    signal.Notify(sigChan, syscall.SIGINT)

    // Process any queued signals
    for {
        select {
        case sig := <-sigChan:
            fmt.Printf("Received queued signal: %v\n", sig)
        case <-time.After(1 * time.Second):
            return
        }
    }
}
```

## Expected Learning Outcomes

After completing this exercise, you should understand:

- [ ] How Linux signals work and their purposes
- [ ] The difference between catchable and non-catchable signals
- [ ] How to implement custom signal handlers in Go
- [ ] Signal masking and blocking mechanisms
- [ ] Signal inheritance in child processes
- [ ] Race conditions and signal safety
- [ ] Signal-based inter-process communication
- [ ] Real-time signals and their advantages

## Integration with Container Learning

This exercise prepares you for later sections where you'll learn:

- How containers handle signals (especially PID 1 responsibilities)
- Signal forwarding in container runtimes
- Process lifecycle management in containers
- Init system implementation for containers

## Common Pitfalls to Avoid

1. **Signal Handler Complexity**: Keep signal handlers simple and fast
2. **Race Conditions**: Be careful with shared data in signal handlers
3. **Signal Masking**: Don't forget to unmask signals
4. **Child Process Cleanup**: Always handle SIGCHLD properly
5. **Signal Delivery**: Remember signals can be lost or coalesced

## Validation Checklist

- [ ] Can send and receive various signal types
- [ ] Implements custom signal handlers correctly
- [ ] Demonstrates signal masking and unmasking
- [ ] Shows parent-child signal inheritance
- [ ] Handles signal race conditions properly
- [ ] Provides clear logging and feedback
- [ ] Works with rapid signal sending
- [ ] Gracefully handles process termination
