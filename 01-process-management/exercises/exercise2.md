# Exercise 2: Process Resource Monitor

## Objective

Create a real-time process monitor that tracks CPU usage, memory consumption, and I/O statistics for one or more processes.

## Requirements

### Basic Requirements

1. Monitor a single process by PID
2. Display CPU percentage, memory usage, and basic I/O stats
3. Update display every second
4. Show process lifecycle (creation to termination)

### Advanced Requirements

1. Monitor multiple processes simultaneously
2. Add historical data and simple graphs (ASCII art)
3. Implement alerts for high resource usage
4. Export monitoring data to CSV/JSON

## Implementation Guidelines

### Step 1: Resource Information Structure

```go
type ResourceUsage struct {
    Timestamp   time.Time
    PID         int
    CPUPercent  float64
    MemoryBytes uint64
    IOReadBytes uint64
    IOWriteBytes uint64
    OpenFDs     int
    Threads     int
}
```

### Step 2: Monitor Structure

```go
type ProcessMonitor struct {
    PIDs        []int
    Interval    time.Duration
    History     []ResourceUsage
    MaxHistory  int
    Alerts      AlertConfig
}
```

### Step 3: Core Functions to Implement

```go
// GetResourceUsage reads current resource usage for a PID
func GetResourceUsage(pid int) (ResourceUsage, error) {
    // TODO: Read /proc/<pid>/stat for CPU times
    // TODO: Read /proc/<pid>/status for memory
    // TODO: Read /proc/<pid>/io for I/O stats
    // TODO: Count open file descriptors from /proc/<pid>/fd
}

// CalculateCPUPercent calculates CPU percentage between two samples
func CalculateCPUPercent(prev, curr ResourceUsage, interval time.Duration) float64 {
    // TODO: Calculate CPU percentage using utime and stime differences
}

// DisplayMonitor shows real-time monitoring information
func (m *ProcessMonitor) DisplayMonitor() {
    // TODO: Implement real-time display with clearing screen
    // TODO: Show tabular data with headers
    // TODO: Add simple ASCII graphs for trends
}
```

## Expected Output

```
Process Monitor - Press Ctrl+C to stop
Time: 2024-01-15 14:30:15

PID    NAME           CPU%   MEM(MB)  I/O Read   I/O Write  FDs  Threads  Status
─────────────────────────────────────────────────────────────────────────────
1234   firefox        15.2   342.1    1.2GB      45.2MB     89   12      Running
5678   code           8.7    128.4    567MB      12.1MB     45   8       Running

Resource History (last 10 samples):
PID 1234: ████████████▓▓▓▓▓▓ (15.2% CPU)
PID 5678: ██████▓▓▓▓▓▓▓▓▓▓▓▓ (8.7% CPU)

Alerts:
⚠️  PID 1234: High memory usage (342.1MB > 300MB threshold)
```

## Resource Files to Read

### `/proc/<pid>/stat`

Contains process statistics including:

- Fields 14-15: utime, stime (CPU time in clock ticks)
- Field 24: RSS (resident set size in pages)

### `/proc/<pid>/status`

Contains human-readable process status:

- `VmRSS`: Physical memory usage
- `VmSize`: Virtual memory size
- `Threads`: Number of threads

### `/proc/<pid>/io`

Contains I/O statistics:

- `read_bytes`: Bytes read from storage
- `write_bytes`: Bytes written to storage

### `/proc/<pid>/fd/`

Directory containing open file descriptors (count them)

## Advanced Features

### 1. CPU Calculation

CPU percentage = ((curr_total - prev_total) / interval_ticks) \* 100
where total = utime + stime

### 2. Memory Monitoring

- Track both physical (RSS) and virtual (VSZ) memory
- Convert from pages to bytes (page size = 4096 typically)
- Monitor memory growth trends

### 3. Alert System

```go
type AlertConfig struct {
    CPUThreshold    float64
    MemoryThreshold uint64
    IOThreshold     uint64
    Enabled         bool
}
```

### 4. Historical Data

- Keep rolling window of last N samples
- Calculate averages and peaks
- Simple trend analysis

## Testing Scenarios

1. **Normal Process**: Monitor a text editor or terminal
2. **High CPU Process**: Run `yes > /dev/null` and monitor
3. **High Memory Process**: Create a program that allocates lots of memory
4. **High I/O Process**: Run `dd if=/dev/zero of=/tmp/test bs=1M count=1000`
5. **Short-lived Process**: Monitor processes that start and stop quickly

## Bonus Challenges

1. **Web Interface**: Create a simple HTTP server to view monitoring data
2. **Process Tree Integration**: Show resource usage in context of process tree
3. **Comparison Mode**: Compare resource usage between multiple processes
4. **Predictive Alerts**: Predict resource exhaustion based on trends
5. **Container Awareness**: Detect if process is running in a container

## Files to Create

- `main.go` - CLI interface and main loop
- `monitor.go` - Core monitoring logic
- `resources.go` - Resource parsing from /proc
- `display.go` - Output formatting and visualization
- `alerts.go` - Alert system
- `history.go` - Historical data management

## Validation Checklist

- [ ] Accurately reads process resource information
- [ ] Calculates CPU percentage correctly
- [ ] Handles process termination gracefully
- [ ] Updates display in real-time
- [ ] Implements basic alerting
- [ ] Manages historical data efficiently
- [ ] Provides useful error messages
