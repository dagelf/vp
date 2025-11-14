package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// Instance represents a running or stopped process instance
type Instance struct {
	Name      string            `json:"name"`      // User-provided name
	Template  string            `json:"template"`  // Template ID
	Command   string            `json:"command"`   // Final interpolated command
	PID       int               `json:"pid"`       // Process ID
	Status    string            `json:"status"`    // stopped|starting|running|stopping|error
	Resources map[string]string `json:"resources"` // resource_type -> value
	Started   int64             `json:"started"`   // Unix timestamp
	Error     string            `json:"error,omitempty"`
}

// Template defines how to start a process
type Template struct {
	ID        string            `json:"id"`        // Unique template ID
	Label     string            `json:"label"`     // Human-readable label
	Command   string            `json:"command"`   // Template with ${var} and %counter
	Resources []string          `json:"resources"` // Resource types this needs
	Vars      map[string]string `json:"vars"`      // Default variables
}

// StartProcess creates and starts a process instance from a template
func StartProcess(state *State, template *Template, name string, vars map[string]string) (*Instance, error) {
	// Check if instance already exists
	if state.Instances[name] != nil {
		return nil, fmt.Errorf("instance %s already exists", name)
	}

	inst := &Instance{
		Name:      name,
		Template:  template.ID,
		Status:    "starting",
		Resources: make(map[string]string),
	}

	// Merge template defaults with provided vars
	finalVars := make(map[string]string)
	for k, v := range template.Vars {
		finalVars[k] = v
	}
	for k, v := range vars {
		finalVars[k] = v
	}

	// Phase 1: Allocate resources declared in template
	for _, rtype := range template.Resources {
		value, err := AllocateResource(state, rtype, finalVars[rtype])
		if err != nil {
			// Rollback all allocated resources
			state.ReleaseResources(name)
			inst.Status = "error"
			inst.Error = fmt.Sprintf("resource allocation failed: %v", err)
			return inst, err
		}
		inst.Resources[rtype] = value
		state.ClaimResource(rtype, value, name)
		finalVars[rtype] = value // Make available for interpolation
	}

	// Phase 2: Interpolate command
	cmd := template.Command

	// Replace ${var} syntax
	for key, val := range finalVars {
		cmd = strings.ReplaceAll(cmd, "${"+key+"}", val)
	}

	// Handle %counter syntax (auto-allocate if not already allocated)
	re := regexp.MustCompile(`%(\w+)`)
	for {
		match := re.FindStringSubmatch(cmd)
		if match == nil {
			break
		}
		counter := match[1]

		// Allocate counter resource
		value, err := AllocateResource(state, counter, "")
		if err != nil {
			state.ReleaseResources(name)
			inst.Status = "error"
			inst.Error = fmt.Sprintf("counter allocation failed: %v", err)
			return inst, err
		}

		cmd = strings.ReplaceAll(cmd, "%"+counter, value)
		inst.Resources[counter] = value
		state.ClaimResource(counter, value, name)
	}

	inst.Command = cmd

	// Phase 3: Start process
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		state.ReleaseResources(name)
		inst.Status = "error"
		inst.Error = "empty command"
		return inst, fmt.Errorf("empty command")
	}

	proc := exec.Command(parts[0], parts[1:]...)
	proc.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}

	if err := proc.Start(); err != nil {
		state.ReleaseResources(name)
		inst.Status = "error"
		inst.Error = fmt.Sprintf("failed to start: %v", err)
		return inst, err
	}

	inst.PID = proc.Process.Pid
	inst.Status = "running"
	inst.Started = time.Now().Unix()

	state.Instances[name] = inst
	state.Save()

	return inst, nil
}

// StopProcess stops a running process instance
func StopProcess(state *State, inst *Instance) error {
	if inst.PID == 0 {
		return fmt.Errorf("instance not running")
	}

	inst.Status = "stopping"

	// Find and kill process
	process, err := os.FindProcess(inst.PID)
	if err != nil {
		inst.Status = "stopped"
		inst.PID = 0
		state.Save()
		return nil // Process doesn't exist, consider it stopped
	}

	// Try graceful shutdown first (SIGTERM)
	err = process.Signal(os.Interrupt)
	if err != nil {
		// Force kill if graceful shutdown fails (SIGKILL)
		process.Kill()
	}

	// Wait a bit for process to exit
	time.Sleep(100 * time.Millisecond)

	inst.Status = "stopped"
	inst.PID = 0
	state.Save()

	return nil
}

// IsProcessRunning checks if a process is still running
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
