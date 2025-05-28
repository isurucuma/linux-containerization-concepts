# Exercise Solutions - Section 1

This directory contains complete reference solutions for all exercises in Section 1: Process Management.

## How to Use These Solutions

1. **Try First**: Always attempt the exercises yourself before looking at solutions
2. **Compare**: Use these solutions to compare your approach
3. **Learn**: Study different techniques and optimizations used
4. **Extend**: Use these as starting points for additional features

## Solutions Overview

### Exercise 1: Process Tree Explorer

**Location**: `exercise1/`

Complete implementation of a process tree explorer that:

- Parses `/proc/[pid]/stat` and `/proc/[pid]/status` files
- Builds hierarchical process trees with parent-child relationships
- Displays formatted process information including PID, PPID, name, state
- Handles process tree traversal and visualization
- Includes error handling for non-existent or inaccessible processes

**Key Features**:

- Robust `/proc` filesystem parsing
- Recursive tree building algorithm
- Clean, formatted output display
- Process state interpretation
- Memory and CPU information extraction

**Usage**:

```bash
cd exercise1
go run main.go tree [root_pid]
go run main.go list
go run main.go info <pid>
```

### Exercise 2: Process Resource Monitor

**Location**: `exercise2/`

Advanced resource monitoring tool that tracks:

- Memory usage (RSS, VSZ) for all processes
- CPU time and percentage calculation
- File descriptor counts
- Thread information
- Process states and transitions
- Top consumers by memory and CPU
- Real-time process monitoring

**Key Features**:

- Comprehensive system resource analysis
- Real-time monitoring with configurable intervals
- Top N process listings by resource usage
- Process state categorization and counting
- Individual process tracking over time
- Formatted output with human-readable units

**Usage**:

```bash
cd exercise2
go run main.go scan                    # Scan all processes
go run main.go top-memory [N]          # Show top N memory processes
go run main.go top-cpu [N]             # Show top N CPU processes
go run main.go states                  # Show process states
go run main.go monitor <PID> [seconds] # Monitor specific process
```

### Exercise 3: Signal Playground

**Location**: `exercise3/`

Interactive signal experimentation environment featuring:

- Comprehensive signal information database
- Test process creation with signal handlers
- Signal sending and monitoring capabilities
- Real-time process state tracking
- Interactive command interface
- Signal handling demonstrations

**Key Features**:

- Complete Linux signal reference (1-31)
- Test process with configurable signal handlers
- Interactive signal sending interface
- Process state monitoring
- Signal effects demonstration
- Educational signal handling scenarios

**Usage**:

```bash
cd exercise3
go run main.go interactive              # Start interactive mode
go run main.go list                     # List all signals
go run main.go demo                     # Run signal handling demo
go run main.go send <PID> <SIGNAL>      # Send signal to process
go run main.go monitor <PID>            # Monitor process state
```

## Building and Running Solutions

Each solution is self-contained with its own `go.mod` file and can be run independently:

```bash
# Exercise 1 - Process Tree Explorer
cd exercise1
go run main.go tree

# Exercise 2 - Process Resource Monitor
cd exercise2
go run main.go top-memory 10

# Exercise 3 - Signal Playground
cd exercise3
go run main.go interactive
```

## Advanced Examples Integration

The solutions complement the advanced examples in `../examples/advanced/`:

- `process_pool.go` - Process pool management
- `resource_analyzer.go` - System-wide resource analysis
- `lifecycle_manager.go` - Complete process lifecycle management

## Learning Notes

These solutions demonstrate:

- Proper error handling for system operations
- Efficient parsing of `/proc` filesystem
- Safe concurrent programming with signals
- Clean separation of concerns
- Good CLI interface design
- Performance considerations for real-time monitoring
- Real-world Go systems programming patterns

## Extensions and Challenges

Use these solutions as base for additional features:

- Add network monitoring to process monitor
- Implement process search and filtering
- Add export capabilities (JSON, CSV)
- Create web interface for monitoring
- Add process control capabilities
- Implement process grouping and categorization
- Add container awareness
- Implement custom signal handlers
- Create process performance profiling tools
