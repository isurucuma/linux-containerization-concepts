#!/bin/bash

# Linux Containerization Learning Environment Setup Script
# This script prepares your system for the containerization learning journey

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}  Linux Containerization Setup  ${NC}"
    echo -e "${BLUE}================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

check_linux() {
    if [[ "$OSTYPE" != "linux-gnu"* ]]; then
        print_error "This learning material requires Linux operating system"
        print_info "Please run this on a Linux machine or VM"
        exit 1
    fi
    print_success "Running on Linux"
}

check_root_access() {
    if [[ $EUID -ne 0 ]]; then
        print_warning "This script needs root privileges for some operations"
        print_info "Please run with sudo or as root"
        
        # Check if sudo is available
        if ! command -v sudo &> /dev/null; then
            print_error "sudo is not available"
            exit 1
        fi
        
        print_info "Continuing with sudo..."
        SUDO="sudo"
    else
        print_success "Running with root privileges"
        SUDO=""
    fi
}

check_kernel_version() {
    KERNEL_VERSION=$(uname -r | cut -d. -f1-2)
    REQUIRED_VERSION="4.0"
    
    if ! printf '%s\n' "$REQUIRED_VERSION" "$KERNEL_VERSION" | sort -V -C; then
        print_error "Kernel version $KERNEL_VERSION is too old"
        print_info "Required: Linux kernel 4.0 or newer"
        print_info "Current: $(uname -r)"
        exit 1
    fi
    
    print_success "Kernel version $KERNEL_VERSION is supported"
}

check_go_installation() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        print_info "Please install Go 1.19 or newer from https://golang.org/dl/"
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
    REQUIRED_GO="1.19"
    
    if ! printf '%s\n' "$REQUIRED_GO" "$GO_VERSION" | sort -V -C; then
        print_error "Go version $GO_VERSION is too old"
        print_info "Required: Go 1.19 or newer"
        exit 1
    fi
    
    print_success "Go version $GO_VERSION is installed"
}

install_system_packages() {
    print_info "Installing required system packages..."
    
    # Detect package manager
    if command -v apt &> /dev/null; then
        $SUDO apt update
        $SUDO apt install -y \
            build-essential \
            strace \
            htop \
            tree \
            jq \
            curl \
            net-tools \
            bridge-utils \
            iptables \
            util-linux \
            coreutils \
            iproute2 \
            procps
    elif command -v yum &> /dev/null; then
        $SUDO yum groupinstall -y "Development Tools"
        $SUDO yum install -y \
            strace \
            htop \
            tree \
            jq \
            curl \
            net-tools \
            bridge-utils \
            iptables \
            util-linux \
            coreutils \
            iproute \
            procps-ng
    elif command -v pacman &> /dev/null; then
        $SUDO pacman -S --noconfirm \
            base-devel \
            strace \
            htop \
            tree \
            jq \
            curl \
            net-tools \
            bridge-utils \
            iptables \
            util-linux \
            coreutils \
            iproute2 \
            procps-ng
    else
        print_warning "Unknown package manager. Please install packages manually:"
        print_info "build-essential, strace, htop, tree, jq, curl, net-tools, bridge-utils, iptables"
        print_info "util-linux, coreutils, iproute2, procps"
        return 0
    fi
    
    print_success "System packages installed"
}

check_cgroup_support() {
    if [[ ! -d "/sys/fs/cgroup" ]]; then
        print_warning "cgroup filesystem not found at /sys/fs/cgroup"
        print_info "cgroups might not be enabled in your kernel"
    else
        print_success "cgroup filesystem is available"
    fi
    
    # Check for cgroup v2
    if [[ -f "/sys/fs/cgroup/cgroup.controllers" ]]; then
        print_success "cgroup v2 is available"
    elif [[ -d "/sys/fs/cgroup/memory" ]]; then
        print_success "cgroup v1 is available"
    else
        print_warning "cgroup controllers might not be available"
    fi
}

check_namespace_support() {
    # Check if unshare command supports various namespace types
    if command -v unshare &> /dev/null; then
        print_success "unshare command is available"
        
        # Test namespace support
        if unshare --help | grep -q -- "--pid"; then
            print_success "PID namespace support detected"
        fi
        
        if unshare --help | grep -q -- "--net"; then
            print_success "Network namespace support detected"
        fi
        
        if unshare --help | grep -q -- "--mount"; then
            print_success "Mount namespace support detected"
        fi
    else
        print_error "unshare command not found"
        print_info "This is required for namespace operations"
    fi
}

# setup_development_environment() {
#     print_info "Setting up development environment..."
    
#     # Create a workspace directory if it doesn't exist
#     WORKSPACE_DIR="$HOME/containerization-learning"
    
#     if [[ ! -d "$WORKSPACE_DIR" ]]; then
#         mkdir -p "$WORKSPACE_DIR"
#         print_success "Created workspace directory: $WORKSPACE_DIR"
#     fi
    
#     # Set up Go environment
#     export GOPATH="$HOME/go"
#     export PATH="$PATH:$GOPATH/bin"
    
#     print_success "Development environment configured"
# }

# run_basic_tests() {
#     print_info "Running basic functionality tests..."
    
#     # Test namespace creation
#     if command -v unshare &> /dev/null; then
#         if unshare --pid --fork echo "PID namespace test" &> /dev/null; then
#             print_success "PID namespace creation works"
#         else
#             print_warning "PID namespace creation failed (might need root)"
#         fi
#     fi
    
#     # Test mount operations
#     if [[ -w "/tmp" ]]; then
#         TEST_DIR="/tmp/container-test-$$"
#         mkdir -p "$TEST_DIR"
#         if mount --bind "$TEST_DIR" "$TEST_DIR" 2>/dev/null; then
#             umount "$TEST_DIR" 2>/dev/null
#             rmdir "$TEST_DIR"
#             print_success "Mount operations work"
#         else
#             print_warning "Mount operations failed (might need root)"
#         fi
#     fi
# }

print_final_instructions() {
    print_success "Setup completed successfully!"
    echo ""
    print_info "Next steps:"
    echo "1. Navigate to the learning material directory"
    echo "2. Start with Section 1: Process Management"
    echo "3. Read each README.md file carefully"
    echo "4. Run examples and complete exercises"
    echo ""
    print_info "Important notes:"
    echo "• Many examples require root privileges (use sudo)"
    echo "• Consider using a virtual machine for safety"
    echo "• Read all code before running it"
    echo ""
    print_info "Happy learning! 🚀"
}

main() {
    print_header
    
    print_info "Checking system requirements..."
    check_linux
    check_root_access
    check_kernel_version
    check_go_installation
    
    print_info "Installing required packages..."
    install_system_packages
    
    print_info "Checking containerization support..."
    check_cgroup_support
    check_namespace_support
    
    # setup_development_environment
    # run_basic_tests
    print_final_instructions
}

# Run main function
main "$@"
