package signal

import (
	"fmt"
	"os"
	"syscall"
)

// SignalMap maps signal names to their numeric values
var SignalMap = map[string]syscall.Signal{
	"SIGTERM": syscall.SIGTERM,
	"SIGKILL": syscall.SIGKILL,
	"SIGINT":  syscall.SIGINT,
	"SIGSTOP": syscall.SIGSTOP,
	"SIGCONT": syscall.SIGCONT,
	"SIGHUP":  syscall.SIGHUP,
	"SIGUSR1": syscall.SIGUSR1,
	"SIGUSR2": syscall.SIGUSR2,
	"SIGQUIT": syscall.SIGQUIT,
	"SIGABRT": syscall.SIGABRT,
}

// SendSignal sends a signal to a process
func SendSignal(pid int, signalName string) error {
	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d not found: %v", pid, err)
	}

	// Get the signal from the map
	sig, exists := SignalMap[signalName]
	if !exists {
		return fmt.Errorf("unknown signal: %s", signalName)
	}

	// Send the signal
	err = process.Signal(sig)
	if err != nil {
		return fmt.Errorf("failed to send signal %s to process %d: %v", signalName, pid, err)
	}

	return nil
}

// ListSignals returns a list of available signals
func ListSignals() []string {
	var signals []string
	for name := range SignalMap {
		signals = append(signals, name)
	}
	return signals
}

// GetSignalDescription returns a description of the signal
func GetSignalDescription(signalName string) string {
	descriptions := map[string]string{
		"SIGTERM": "Termination signal - polite request to terminate",
		"SIGKILL": "Kill signal - immediate termination (cannot be caught)",
		"SIGINT":  "Interrupt signal - typically from Ctrl+C",
		"SIGSTOP": "Stop signal - suspend process (cannot be caught)",
		"SIGCONT": "Continue signal - resume stopped process",
		"SIGHUP":  "Hangup signal - terminal disconnection",
		"SIGUSR1": "User-defined signal 1",
		"SIGUSR2": "User-defined signal 2",
		"SIGQUIT": "Quit signal - typically from Ctrl+\\",
		"SIGABRT": "Abort signal - abnormal termination",
	}

	desc, exists := descriptions[signalName]
	if !exists {
		return "Unknown signal"
	}
	return desc
}
