# Exercise 1: Process Tree Explorer

## Objective

Build a program that displays the process tree starting from a given PID, showing parent-child relationships in a hierarchical format.

## Requirements

### Basic Requirements

1. Read process information from `/proc/<pid>/stat` and `/proc/<pid>/status`
2. Parse parent-child relationships
3. Display the tree structure with proper indentation
4. Show process names, PIDs, and states

### Advanced Requirements

1. Add memory and CPU usage information
2. Implement tree pruning (limit depth)th)
3. Add process filtering capabilities
4. Color-code different process states

## Implementation Guidelines

### Step 1: Process Information Structure

```go
type ProcessInfo struct {
    PID     int
    PPID    int
    Name    string
    State   string
    Memory  uint64
    Threads int
}
```

### Step 2: Tree Node Structure

```go
type ProcessNode struct {
    Info     ProcessInfo
    Children []*ProcessNode
    Parent   *ProcessNode
}
```

### Step 3: Core Functions to Implement

```go
// ReadProcessInfo reads process information from /proc
func ReadProcessInfo(pid int) (ProcessInfo, error) {
    // TODO: Implement reading from /proc/<pid>/stat
    // TODO: Parse the fields correctly
    // TODO: Read process name from /proc/<pid>/comm
    // TODO: Read memory info from /proc/<pid>/status
}

// BuildProcessTree builds a tree starting from rootPID
func BuildProcessTree(rootPID int) (*ProcessNode, error) {
    // TODO: Get all processes
    // TODO: Build parent-child relationships
    // TODO: Create tree structure
}

// DisplayTree displays the tree in a formatted way
func DisplayTree(root *ProcessNode, prefix string, isLast bool) {
    // TODO: Implement tree visualization
    // Use characters like ├── and └── for tree structure
}
```

## Expected Output

```
Process Tree (starting from PID 1):
init (1) [S] 8.2MB
├── systemd --user (234) [S] 12.1MB
│   ├── dbus-daemon (456) [S] 3.2MB
│   └── pulseaudio (567) [S] 15.6MB
├── kthreadd (2) [S] 0.0MB
│   ├── ksoftirqd/0 (3) [S] 0.0MB
│   └── migration/0 (4) [S] 0.0MB
└── bash (1234) [S] 2.1MB
    └── go run main.go (1456) [R] 25.3MB
        └── main (1457) [R] 18.7MB
```

## Testing

1. Test with different starting PIDs
2. Test with processes that have many children
3. Test error handling for non-existent PIDs
4. Test performance with large process trees

## Bonus Challenges

1. Add real-time updates (refresh every few seconds)
2. Implement process search functionality
3. Add the ability to send signals to processes from the tree view
4. Export tree to different formats (JSON, XML, DOT for graphviz)

## Files to Create

- `main.go` - CLI interface
- `process.go` - Process information parsing
- `tree.go` - Tree building and display logic
- `go.mod` - Module definition

## Hints

- The `/proc/<pid>/stat` file contains space-separated fields
- Process name is in parentheses and may contain spaces
- Be careful with parsing - some fields can be negative
- Handle the case where processes disappear during execution
- Consider using `bufio.Scanner` for efficient file reading

## Validation

Your solution should:

- [ ] Correctly parse process information from `/proc`
- [ ] Build accurate parent-child relationships
- [ ] Display a properly formatted tree
- [ ] Handle errors gracefully
- [ ] Work with any valid PID as starting point
