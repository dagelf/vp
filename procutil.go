package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProcessInfo contains detailed information about a discovered process
type ProcessInfo struct {
	PID       int               `json:"pid"`
	PPID      int               `json:"ppid"`       // Parent process ID
	Name      string            `json:"name"`       // Process name
	Cmdline   string            `json:"cmdline"`    // Full command line
	Exe       string            `json:"exe"`        // Executable path
	Cwd       string            `json:"cwd"`        // Working directory
	Environ   map[string]string `json:"environ"`    // Environment variables
	Ports     []int             `json:"ports"`      // TCP ports this process listens on
	ParentChain []ProcessInfo   `json:"parent_chain,omitempty"` // Parent process chain
}

// ShellNames contains common shell executable names
var ShellNames = map[string]bool{
	"sh":      true,
	"bash":    true,
	"zsh":     true,
	"fish":    true,
	"dash":    true,
	"ksh":     true,
	"tcsh":    true,
	"csh":     true,
}

// ReadProcessInfo reads process information from /proc/[pid]
func ReadProcessInfo(pid int) (*ProcessInfo, error) {
	procDir := fmt.Sprintf("/proc/%d", pid)

	// Check if process exists
	if _, err := os.Stat(procDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("process %d does not exist", pid)
	}

	info := &ProcessInfo{
		PID:     pid,
		Environ: make(map[string]string),
	}

	// Read PPID from /proc/[pid]/stat
	statData, err := os.ReadFile(filepath.Join(procDir, "stat"))
	if err != nil {
		return nil, fmt.Errorf("failed to read stat: %w", err)
	}

	// Parse stat file - format: pid (name) state ppid ...
	// We need to handle names with spaces/parentheses
	statStr := string(statData)
	lastParen := strings.LastIndex(statStr, ")")
	if lastParen == -1 {
		return nil, fmt.Errorf("invalid stat format")
	}

	// Extract name from (name)
	firstParen := strings.Index(statStr, "(")
	if firstParen != -1 && lastParen > firstParen {
		info.Name = statStr[firstParen+1 : lastParen]
	}

	// Parse fields after name
	fields := strings.Fields(statStr[lastParen+1:])
	if len(fields) >= 2 {
		info.PPID, _ = strconv.Atoi(fields[1]) // Third field is PPID
	}

	// Read command line
	cmdlineData, err := os.ReadFile(filepath.Join(procDir, "cmdline"))
	if err == nil {
		// cmdline is null-separated, convert to space-separated
		cmdline := strings.ReplaceAll(string(cmdlineData), "\x00", " ")
		info.Cmdline = strings.TrimSpace(cmdline)
	}

	// Read executable path
	exePath, err := os.Readlink(filepath.Join(procDir, "exe"))
	if err == nil {
		info.Exe = exePath
	}

	// Read working directory
	cwdPath, err := os.Readlink(filepath.Join(procDir, "cwd"))
	if err == nil {
		info.Cwd = cwdPath
	}

	// Read environment variables
	environData, err := os.ReadFile(filepath.Join(procDir, "environ"))
	if err == nil {
		environStr := string(environData)
		for _, pair := range strings.Split(environStr, "\x00") {
			if pair == "" {
				continue
			}
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				info.Environ[parts[0]] = parts[1]
			}
		}
	}

	// Read ports this process is listening on
	ports, err := GetPortsForProcess(pid)
	if err == nil {
		info.Ports = ports
	}

	return info, nil
}

// GetParentChain traverses the parent process chain up to init (PID 1)
func GetParentChain(pid int) ([]ProcessInfo, error) {
	var chain []ProcessInfo
	currentPID := pid
	seen := make(map[int]bool) // Prevent infinite loops

	for currentPID > 0 && !seen[currentPID] {
		seen[currentPID] = true

		info, err := ReadProcessInfo(currentPID)
		if err != nil {
			break // Process no longer exists
		}

		chain = append(chain, *info)

		// Stop if we've reached init (PID 1) or if parent is 0
		if currentPID == 1 || info.PPID == 0 {
			break
		}

		currentPID = info.PPID

		// Safety: limit chain length
		if len(chain) > 100 {
			break
		}
	}

	return chain, nil
}

// FindLaunchScript finds the "launch script" in the parent chain
// This is typically the first child of a shell (e.g., "bun dev" launched from bash)
func FindLaunchScript(chain []ProcessInfo) *ProcessInfo {
	// Strategy: Find the first process whose parent is a shell
	for i := 0; i < len(chain); i++ {
		if i+1 < len(chain) {
			parent := chain[i+1]
			if IsShell(parent.Name) || IsShell(filepath.Base(parent.Exe)) {
				return &chain[i]
			}
		}
	}

	// Fallback: Return the last process in chain (closest to user action)
	// before we hit systemd/init
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].PID != 1 && chain[i].Name != "systemd" {
			return &chain[i]
		}
	}

	return nil
}

// IsShell checks if a process name is a known shell
func IsShell(name string) bool {
	return ShellNames[name]
}

// GetPortsForProcess finds all TCP ports that a specific process is listening on
func GetPortsForProcess(pid int) ([]int, error) {
	// Get all socket inodes for this process
	socketInodes := make(map[string]bool)
	fdDir := filepath.Join("/proc", strconv.Itoa(pid), "fd")

	fds, err := os.ReadDir(fdDir)
	if err != nil {
		return nil, err
	}

	for _, fd := range fds {
		link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
		if err != nil {
			continue
		}
		// Socket links look like "socket:[12345]"
		if strings.HasPrefix(link, "socket:[") {
			inode := strings.TrimPrefix(link, "socket:[")
			inode = strings.TrimSuffix(inode, "]")
			socketInodes[inode] = true
		}
	}

	if len(socketInodes) == 0 {
		return []int{}, nil
	}

	// Now scan /proc/net/tcp and /proc/net/tcp6 for these inodes
	ports := make(map[int]bool)

	for _, tcpFile := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		file, err := os.Open(tcpFile)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Scan() // Skip header

		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 10 {
				continue
			}

			// Field 3 is connection state (0A = LISTEN)
			if fields[3] != "0A" {
				continue // Only interested in listening sockets
			}

			// Field 9 is inode
			inode := fields[9]

			// Check if this inode belongs to our process
			if !socketInodes[inode] {
				continue
			}

			// Field 1 is local_address in format "IP:PORT" (hex)
			localAddr := fields[1]
			parts := strings.Split(localAddr, ":")
			if len(parts) != 2 {
				continue
			}

			// Parse port (hex)
			portHex := parts[1]
			portNum, err := strconv.ParseInt(portHex, 16, 64)
			if err != nil {
				continue
			}

			ports[int(portNum)] = true
		}
	}

	// Convert map to slice
	result := make([]int, 0, len(ports))
	for port := range ports {
		result = append(result, port)
	}

	return result, nil
}

// GetProcessesListeningOnPort finds all processes listening on a specific TCP port
func GetProcessesListeningOnPort(port int) ([]int, error) {
	// Read /proc/net/tcp and /proc/net/tcp6
	pids := make(map[int]bool)

	// Parse tcp files
	for _, tcpFile := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		file, err := os.Open(tcpFile)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Scan() // Skip header

		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 10 {
				continue
			}

			// Field 1 is local_address in format "IP:PORT" (hex)
			localAddr := fields[1]
			parts := strings.Split(localAddr, ":")
			if len(parts) != 2 {
				continue
			}

			// Parse port (hex)
			portHex := parts[1]
			portNum, err := strconv.ParseInt(portHex, 16, 64)
			if err != nil {
				continue
			}

			// Check if this is the port we're looking for
			if int(portNum) != port {
				continue
			}

			// Field 9 is inode
			inode := fields[9]

			// Find process using this socket
			pid, err := findProcessByInode(inode)
			if err == nil {
				pids[pid] = true
			}
		}
	}

	// Convert map to slice
	result := make([]int, 0, len(pids))
	for pid := range pids {
		result = append(result, pid)
	}

	return result, nil
}

// findProcessByInode searches /proc/*/fd/* for the given socket inode
func findProcessByInode(inode string) (int, error) {
	socketRef := fmt.Sprintf("socket:[%s]", inode)

	procDir, err := os.Open("/proc")
	if err != nil {
		return 0, err
	}
	defer procDir.Close()

	entries, err := procDir.Readdirnames(-1)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		// Check if entry is a PID (numeric)
		pid, err := strconv.Atoi(entry)
		if err != nil {
			continue
		}

		// Check all file descriptors
		fdDir := filepath.Join("/proc", entry, "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}

			if link == socketRef {
				return pid, nil
			}
		}
	}

	return 0, fmt.Errorf("no process found for inode %s", inode)
}

// DiscoverProcess discovers a process and its launch context
// This enriches process info with parent chain and identifies the launch script
func DiscoverProcess(pid int) (*ProcessInfo, error) {
	chain, err := GetParentChain(pid)
	if err != nil {
		return nil, err
	}

	if len(chain) == 0 {
		return nil, fmt.Errorf("could not read process info for PID %d", pid)
	}

	// The first element in chain is the target process itself
	info := chain[0]

	// Attach parent chain (excluding self)
	if len(chain) > 1 {
		info.ParentChain = chain[1:]
	}

	return &info, nil
}

// DiscoverProcessOnPort discovers the process listening on a port and finds its launch script
func DiscoverProcessOnPort(port int) (*ProcessInfo, *ProcessInfo, error) {
	pids, err := GetProcessesListeningOnPort(port)
	if err != nil {
		return nil, nil, err
	}

	if len(pids) == 0 {
		return nil, nil, fmt.Errorf("no process listening on port %d", port)
	}

	// Use the first PID found
	pid := pids[0]

	// Get full process info with parent chain
	procInfo, err := DiscoverProcess(pid)
	if err != nil {
		return nil, nil, err
	}

	// Build full chain including the process itself
	fullChain := append([]ProcessInfo{*procInfo}, procInfo.ParentChain...)

	// Find the launch script
	launchScript := FindLaunchScript(fullChain)

	return procInfo, launchScript, nil
}
