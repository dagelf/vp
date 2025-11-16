package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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
	Cwd       string            `json:"cwd,omitempty"`       // Working directory
	Managed   bool              `json:"managed"`             // true=can stop/restart, false=monitor only
	Error     string            `json:"error,omitempty"`

  // Discovery fields - populated when discovering existing processes
	LaunchScript *ProcessInfo      `json:"launch_script,omitempty"` // The script that launched this (child of shell)
	ParentChain  []ProcessInfo     `json:"parent_chain,omitempty"`  // Parent process chain
	Discovered   bool              `json:"discovered,omitempty"`    // Was this discovered vs created by us?
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
	inst.Managed = true // Processes started by us are managed

	// Capture working directory
	if cwd, err := os.Getwd(); err == nil {
		inst.Cwd = cwd
	}

	state.Instances[name] = inst
	state.Save()

	// Start a goroutine to wait for the process and reap it
	go func() {
		proc.Wait() // This reaps the zombie when process exits
		// Process has exited, update status if instance still exists
		if inst, exists := state.Instances[name]; exists && inst.PID == proc.Process.Pid {
			inst.Status = "stopped"
			inst.PID = 0
			state.Save()
		}
	}()

	return inst, nil
}

// StopProcess stops a running process instance
func StopProcess(state *State, inst *Instance) error {
	if inst.PID == 0 {
		return fmt.Errorf("instance not running")
	}

	inst.Status = "stopping"

	// Kill the entire process group (negative PID)
	// Since we started with Setpgid:true, we need to kill the group
	pgid := inst.PID
	err := syscall.Kill(-pgid, syscall.SIGTERM)
	if err != nil {
		// If process group kill fails, try individual process
		process, err := os.FindProcess(inst.PID)
		if err != nil {
			inst.Status = "stopped"
			inst.PID = 0
			state.Save()
			return nil
		}
		process.Signal(syscall.SIGTERM)
	}

	// Wait up to 2 seconds for graceful shutdown
	for i := 0; i < 20; i++ {
		if !IsProcessRunning(inst.PID) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Force kill if still running
	if IsProcessRunning(inst.PID) {
		syscall.Kill(-pgid, syscall.SIGKILL)
		time.Sleep(100 * time.Millisecond)
	}

	// Reap any zombie processes by trying to wait
	// This is best-effort since we may not be the parent
	process, _ := os.FindProcess(inst.PID)
	if process != nil {
		process.Wait()
	}

	inst.Status = "stopped"
	inst.PID = 0
	state.Save()

	return nil
}

// RestartProcess restarts a stopped instance with the same resources and command
func RestartProcess(state *State, inst *Instance) error {
	// Instance must be stopped
	if inst.Status != "stopped" {
		return fmt.Errorf("instance %s is not stopped (status: %s)", inst.Name, inst.Status)
	}

	// Try to re-claim the same resources
	for rtype, value := range inst.Resources {
		// Check if resource type still exists
		rt := state.Types[rtype]
		if rt == nil {
			return fmt.Errorf("resource type %s no longer exists", rtype)
		}

		// Check if resource value is available
		if !CheckResource(rt, value) {
			return fmt.Errorf("resource %s=%s no longer available", rtype, value)
		}

		// Claim it
		state.ClaimResource(rtype, value, inst.Name)
	}

	// Start the process with the stored command
	parts := strings.Fields(inst.Command)
	if len(parts) == 0 {
		state.ReleaseResources(inst.Name)
		return fmt.Errorf("empty command")
	}

	proc := exec.Command(parts[0], parts[1:]...)
	proc.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}

	if err := proc.Start(); err != nil {
		state.ReleaseResources(inst.Name)
		inst.Status = "error"
		inst.Error = fmt.Sprintf("failed to restart: %v", err)
		state.Save()
		return err
	}

	inst.PID = proc.Process.Pid
	inst.Status = "running"
	inst.Started = time.Now().Unix()
	inst.Error = ""
	state.Save()

	// Reap zombie when process exits
	go func() {
		proc.Wait()
		if inst, exists := state.Instances[inst.Name]; exists && inst.PID == proc.Process.Pid {
			inst.Status = "stopped"
			inst.PID = 0
			state.Save()
		}
	}()

	return nil
}

// MonitorProcess adds an existing process to vibeprocess as monitored (not managed)
func MonitorProcess(state *State, pid int, name string) (*Instance, error) {
	// Check if instance name already exists
	if state.Instances[name] != nil {
		return nil, fmt.Errorf("instance %s already exists", name)
	}

	// Check if process exists
	if !IsProcessRunning(pid) {
		return nil, fmt.Errorf("process %d not running", pid)
	}

	// Read process info using existing functionality
	procInfo, err := ReadProcessInfo(pid)
	if err != nil {
		return nil, fmt.Errorf("cannot read process %d: %w", pid, err)
	}

	cmdline := procInfo.Cmdline
	if cmdline == "" {
		return nil, fmt.Errorf("cannot read process %d", pid)
	}

	cwd := procInfo.Cwd

	// Detect resources (ports)
	resources := make(map[string]string)
	if len(procInfo.Ports) > 0 {
		// Just record the first port as tcpport
		resources["tcpport"] = fmt.Sprintf("%d", procInfo.Ports[0])
	}

	// Check if we can manage this process (send signals to it)
	managed := canManageProcess(pid)

	inst := &Instance{
		Name:      name,
		Command:   cmdline,
		PID:       pid,
		Status:    "running",
		Resources: resources,
		Cwd:       cwd,
		Managed:   managed, // true if we can send signals, false if different user
		Started:   time.Now().Unix(),
	}

	// Claim resources (monitored processes DO use resources!)
	for rtype, value := range resources {
		state.ClaimResource(rtype, value, name)
	}

	state.Instances[name] = inst
	state.Save()

	// Start monitoring goroutine to detect when process exits
	go func() {
		for {
			time.Sleep(2 * time.Second)
			if !IsProcessRunning(pid) {
				if inst, exists := state.Instances[name]; exists && inst.PID == pid {
					inst.Status = "stopped"
					inst.PID = 0
					state.Save()
				}
				break
			}
		}
	}()

	return inst, nil
}

// canManageProcess checks if we have permission to send signals to a process
func canManageProcess(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Try to send signal 0 (null signal) to test permissions
	// If we get EPERM, we can't manage it. If we get no error, we can.
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// EPERM means process exists but we can't signal it
		return false
	}
	return true
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

// DiscoverAndImportProcess discovers a process by PID and imports it as an instance
func DiscoverAndImportProcess(state *State, pid int, name string) (*Instance, error) {
	// Check if instance name already exists
	if state.Instances[name] != nil {
		return nil, fmt.Errorf("instance %s already exists", name)
	}

	// Discover the process with parent chain
	procInfo, err := DiscoverProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to discover process: %w", err)
	}

	// Build full chain including the process itself
	fullChain := append([]ProcessInfo{*procInfo}, procInfo.ParentChain...)

	// Find the launch script
	launchScript := FindLaunchScript(fullChain)

	// Create instance
	inst := &Instance{
		Name:         name,
		Template:     "discovered",
		Command:      procInfo.Cmdline,
		PID:          pid,
		Status:       "running",
		Resources:    make(map[string]string),
		Started:      time.Now().Unix(),
		Discovered:   true,
		ParentChain:  procInfo.ParentChain,
		LaunchScript: launchScript,
	}

	state.Instances[name] = inst
	state.Save()

	return inst, nil
}

// DiscoverAndImportProcessOnPort discovers a process listening on a port and imports it
func DiscoverAndImportProcessOnPort(state *State, port int, name string) (*Instance, error) {
	// Check if instance name already exists
	if state.Instances[name] != nil {
		return nil, fmt.Errorf("instance %s already exists", name)
	}

	// Discover process on port
	procInfo, launchScript, err := DiscoverProcessOnPort(port)
	if err != nil {
		return nil, fmt.Errorf("failed to discover process on port %d: %w", port, err)
	}

	// Create instance
	inst := &Instance{
		Name:         name,
		Template:     "discovered",
		Command:      procInfo.Cmdline,
		PID:          procInfo.PID,
		Status:       "running",
		Resources:    make(map[string]string),
		Started:      time.Now().Unix(),
		Discovered:   true,
		ParentChain:  procInfo.ParentChain,
		LaunchScript: launchScript,
	}

	// Record the port as a resource
	inst.Resources["tcpport"] = fmt.Sprintf("%d", port)

	state.Instances[name] = inst
	state.Save()

	return inst, nil
}

// DiscoverProcesses discovers running processes on the system
// If portsOnly is true, only returns processes listening on ports
func DiscoverProcesses(state *State, portsOnly bool) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// Read all PIDs from /proc
	procDir, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer procDir.Close()

	entries, err := procDir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Check if entry is a PID (numeric)
		pid, err := strconv.Atoi(entry)
		if err != nil {
			continue
		}

		// Skip if already monitored
		alreadyMonitored := false
		for _, inst := range state.Instances {
			if inst.PID == pid {
				alreadyMonitored = true
				break
			}
		}
		if alreadyMonitored {
			continue
		}

		// Read process info
		procInfo, err := ReadProcessInfo(pid)
		if err != nil {
			continue // Skip processes we can't read
		}

		// If portsOnly, skip processes not listening on ports
		if portsOnly && len(procInfo.Ports) == 0 {
			continue
		}

		// Build result entry
		entry := map[string]interface{}{
			"pid":     procInfo.PID,
			"name":    procInfo.Name,
			"command": procInfo.Cmdline,
			"cwd":     procInfo.Cwd,
			"exe":     procInfo.Exe,
			"ports":     procInfo.Ports,
			"resources": map[string]string{}, // Empty resources for discovered processes
		}

		result = append(result, entry)
	}

	return result, nil
}

// MatchAndUpdateInstances discovers running processes and updates existing instances
// if their resources and commands match
func MatchAndUpdateInstances(state *State) error {
	// Discover all processes (not just those with ports)
	processes, err := DiscoverProcesses(state, false)
	if err != nil {
		return fmt.Errorf("failed to discover processes: %w", err)
	}

	// For each discovered process, try to match it with existing instances
	for _, proc := range processes {
		pid, ok := proc["pid"].(int)
		if !ok {
			continue
		}

		// Read full process info to get ports and parent chain
		procInfo, err := ReadProcessInfo(pid)
		if err != nil {
			continue
		}

		// Build resources map for this process (mainly ports)
		procResources := make(map[string]string)
		for _, port := range procInfo.Ports {
			procResources["tcpport"] = fmt.Sprintf("%d", port)
			break // Just use the first port for now
		}

		// Try to match with existing instances
		for _, inst := range state.Instances {
			// Skip instances that are already running
			if inst.Status == "running" && IsProcessRunning(inst.PID) {
				continue
			}

			// Check if resources match
			resourcesMatch := false
			if len(inst.Resources) > 0 && len(procResources) > 0 {
				// Check if any resource matches
				for rtype, rvalue := range inst.Resources {
					if procResources[rtype] == rvalue {
						resourcesMatch = true
						break
					}
				}
			}

			if !resourcesMatch {
				continue
			}

			// Resources match - now check if command matches
			// Get full parent chain for the discovered process
			fullProcInfo, err := DiscoverProcess(pid)
			if err != nil {
				continue
			}

			// Build list of commands to check (process + parent chain)
			commandsToCheck := []string{fullProcInfo.Cmdline}
			for _, parent := range fullProcInfo.ParentChain {
				commandsToCheck = append(commandsToCheck, parent.Cmdline)
			}

			// Check if instance command matches any of the discovered commands
			commandMatches := false
			for _, cmd := range commandsToCheck {
				if strings.Contains(cmd, inst.Command) || strings.Contains(inst.Command, cmd) {
					commandMatches = true
					break
				}
			}

			if commandMatches {
				// Update the instance
				inst.PID = pid
				inst.Status = "running"
				inst.Started = time.Now().Unix()

				// Update parent chain and launch script if discovered
				inst.ParentChain = fullProcInfo.ParentChain
				fullChain := append([]ProcessInfo{*fullProcInfo}, fullProcInfo.ParentChain...)
				inst.LaunchScript = FindLaunchScript(fullChain)

				state.Save()
				break // Move to next discovered process
			}
		}
	}

	return nil
}
