# Linux Containerization Reference Guide

## 🔧 Essential System Calls

### Process Management

```go
// fork() - Create child process
pid := syscall.ForkExec(path, args, &syscall.ProcAttr{})

// exec() - Replace process image
syscall.Exec(path, args, env)

// wait() - Wait for child process
syscall.Wait4(pid, &status, 0, nil)

// getpid() - Get process ID
pid := syscall.Getpid()

// getppid() - Get parent process ID
ppid := syscall.Getppid()
```

### Namespace Operations

```go
// unshare() - Create new namespaces
syscall.Unshare(syscall.CLONE_NEWPID | syscall.CLONE_NEWNS)

// setns() - Join existing namespace
fd, _ := syscall.Open("/proc/1234/ns/net", syscall.O_RDONLY, 0)
syscall.Syscall(syscall.SYS_SETNS, uintptr(fd), syscall.CLONE_NEWNET, 0)

// clone() - Create process with new namespaces
syscall.Clone(syscall.CLONE_NEWPID | syscall.CLONE_NEWNS)
```

### Mount Operations

```go
// mount() - Mount filesystem
syscall.Mount(source, target, fstype, flags, data)

// umount() - Unmount filesystem
syscall.Unmount(target, flags)

// pivot_root() - Change root filesystem
syscall.PivotRoot(new_root, old_root)

// chroot() - Change root directory
syscall.Chroot(path)
```

### cgroup Operations

```go
// Write to cgroup files
ioutil.WriteFile("/sys/fs/cgroup/memory/mygroup/memory.limit_in_bytes", []byte("100M"), 0644)

// Read cgroup stats
data, _ := ioutil.ReadFile("/sys/fs/cgroup/memory/mygroup/memory.usage_in_bytes")
```

## 📋 Namespace Types Reference

### PID Namespace (CLONE_NEWPID)

- **Purpose**: Process ID isolation
- **Isolation**: Process IDs, /proc filesystem
- **Use Case**: Container process isolation
- **Key Files**: `/proc/*/ns/pid`

### Mount Namespace (CLONE_NEWNS)

- **Purpose**: Filesystem mount isolation
- **Isolation**: Mount points, filesystem hierarchy
- **Use Case**: Container filesystem isolation
- **Key Files**: `/proc/*/ns/mnt`

### Network Namespace (CLONE_NEWNET)

- **Purpose**: Network stack isolation
- **Isolation**: Network interfaces, routing tables, firewall rules
- **Use Case**: Container networking
- **Key Files**: `/proc/*/ns/net`

### UTS Namespace (CLONE_NEWUTS)

- **Purpose**: Hostname and domain name isolation
- **Isolation**: hostname, domainname
- **Use Case**: Container identity
- **Key Files**: `/proc/*/ns/uts`

### IPC Namespace (CLONE_NEWIPC)

- **Purpose**: Inter-process communication isolation
- **Isolation**: System V IPC, POSIX message queues
- **Use Case**: Container communication isolation
- **Key Files**: `/proc/*/ns/ipc`

### User Namespace (CLONE_NEWUSER)

- **Purpose**: User and group ID isolation
- **Isolation**: UIDs, GIDs, capabilities
- **Use Case**: Container security
- **Key Files**: `/proc/*/ns/user`

### Cgroup Namespace (CLONE_NEWCGROUP)

- **Purpose**: Control group isolation
- **Isolation**: cgroup hierarchy view
- **Use Case**: Resource management isolation
- **Key Files**: `/proc/*/ns/cgroup`

## 🎛️ cgroup Controllers Reference

### Memory Controller

```bash
# Files and their purposes
memory.limit_in_bytes          # Memory limit
memory.usage_in_bytes          # Current memory usage
memory.max_usage_in_bytes      # Peak memory usage
memory.swappiness              # Swap tendency (0-100)
memory.oom_control             # OOM killer control
```

### CPU Controller

```bash
# Files and their purposes
cpu.shares                     # CPU shares (relative weight)
cpu.cfs_quota_us              # CPU quota in microseconds
cpu.cfs_period_us             # CPU period in microseconds
cpu.stat                      # CPU usage statistics
```

### Block I/O Controller

```bash
# Files and their purposes
blkio.weight                  # I/O weight (100-1000)
blkio.throttle.read_bps_device # Read bandwidth limit
blkio.throttle.write_bps_device # Write bandwidth limit
```

### Device Controller

```bash
# Files and their purposes
devices.allow                 # Allow device access
devices.deny                  # Deny device access
devices.list                  # List current permissions
```

## 🔒 Capabilities Reference

### Common Capabilities

```go
const (
    CAP_CHOWN            = 0   // Change file ownership
    CAP_DAC_OVERRIDE     = 1   // Override access restrictions
    CAP_DAC_READ_SEARCH  = 2   // Override read restrictions
    CAP_FOWNER           = 3   // Override file owner restrictions
    CAP_KILL             = 5   // Send signals to any process
    CAP_SETGID           = 6   // Manipulate group IDs
    CAP_SETUID           = 7   // Manipulate user IDs
    CAP_NET_BIND_SERVICE = 10  // Bind to privileged ports
    CAP_NET_RAW          = 13  // Use raw sockets
    CAP_SYS_CHROOT       = 18  // Use chroot()
    CAP_SYS_ADMIN        = 21  // Perform admin operations
    CAP_MKNOD            = 27  // Create device files
)
```

## 🌐 Network Configuration Commands

### Virtual Ethernet (veth) Pairs

```bash
# Create veth pair
ip link add veth0 type veth peer name veth1

# Move to namespace
ip link set veth1 netns <namespace>

# Configure IP addresses
ip addr add 10.0.0.1/24 dev veth0
ip netns exec <namespace> ip addr add 10.0.0.2/24 dev veth1

# Bring interfaces up
ip link set veth0 up
ip netns exec <namespace> ip link set veth1 up
```

### Bridge Configuration

```bash
# Create bridge
ip link add br0 type bridge

# Add interface to bridge
ip link set veth0 master br0

# Configure bridge
ip addr add 172.17.0.1/16 dev br0
ip link set br0 up
```

### Network Namespaces

```bash
# Create network namespace
ip netns add myns

# List namespaces
ip netns list

# Execute in namespace
ip netns exec myns ip link list

# Delete namespace
ip netns delete myns
```

## 📁 Important File Paths

### Namespace Information

```
/proc/<pid>/ns/           # Namespace links
/proc/<pid>/ns/pid        # PID namespace
/proc/<pid>/ns/mnt        # Mount namespace
/proc/<pid>/ns/net        # Network namespace
/proc/<pid>/ns/uts        # UTS namespace
/proc/<pid>/ns/ipc        # IPC namespace
/proc/<pid>/ns/user       # User namespace
/proc/<pid>/ns/cgroup     # Cgroup namespace
```

### cgroup Hierarchies

```
/sys/fs/cgroup/           # cgroup v1 root
/sys/fs/cgroup/unified/   # cgroup v2 root (sometimes)
/sys/fs/cgroup/memory/    # Memory controller (v1)
/sys/fs/cgroup/cpu/       # CPU controller (v1)
/sys/fs/cgroup/devices/   # Device controller (v1)
```

### Container Runtime Paths

```
/var/lib/containers/      # Container storage
/var/run/containers/      # Runtime state
/tmp/                     # Temporary mounts
/proc/mounts             # Mount information
/proc/filesystems        # Available filesystems
```

## 🔍 Debugging Commands

### Process Inspection

```bash
# Process tree
pstree -p

# Process namespaces
lsns -p <pid>

# Process cgroups
cat /proc/<pid>/cgroup

# Process capabilities
getpcaps <pid>
```

### Network Debugging

```bash
# Network interfaces in namespace
ip netns exec <ns> ip link list

# Routing table in namespace
ip netns exec <ns> ip route

# Network connections
ss -tulpn
```

### Mount Debugging

```bash
# Current mounts
findmnt

# Mount propagation
findmnt -o TARGET,PROPAGATION

# Filesystem types
cat /proc/filesystems
```

### cgroup Inspection

```bash
# List cgroups
systemd-cgls

# cgroup tree
tree /sys/fs/cgroup/

# Process cgroup membership
cat /proc/<pid>/cgroup
```

## ⚡ Performance Monitoring

### System Resources

```bash
# Memory usage
free -h
cat /proc/meminfo

# CPU usage
top
htop
cat /proc/loadavg

# I/O statistics
iostat
iotop
```

### Container Metrics

```bash
# Memory usage in cgroup
cat /sys/fs/cgroup/memory/<group>/memory.usage_in_bytes

# CPU usage in cgroup
cat /sys/fs/cgroup/cpu/<group>/cpu.stat

# I/O usage in cgroup
cat /sys/fs/cgroup/blkio/<group>/blkio.throttle.io_service_bytes
```

## 🛡️ Security Considerations

### Safe Practices

- Always validate input parameters
- Use minimal required capabilities
- Implement proper error handling
- Clean up resources on exit
- Use temporary directories for experiments
- Verify permissions before operations

### Common Pitfalls

- Forgetting to unmount filesystems
- Not cleaning up network namespaces
- Capability escalation vulnerabilities
- Resource leaks in cgroups
- Signal handling in PID 1

## 📚 Additional Resources

### Man Pages

```bash
man 2 clone          # Process creation
man 2 unshare        # Namespace creation
man 2 mount          # Filesystem mounting
man 7 namespaces     # Namespace overview
man 7 cgroups        # Control groups
man 7 capabilities   # Linux capabilities
```

### Kernel Documentation

- `/Documentation/cgroup-v1/` - cgroup v1 documentation
- `/Documentation/admin-guide/cgroup-v2.rst` - cgroup v2 documentation
- `/Documentation/filesystems/overlayfs.txt` - OverlayFS documentation

This reference guide will be your companion throughout the learning journey!
