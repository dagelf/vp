# Vibeprocess Manager - Implementation Plan

## Minimal Go Implementation with Progressive Enhancement

This document outlines the implementation plan for Vibeprocess Manager, starting with a minimal Go program that can expand to include all features from the PRD. The approach prioritizes simplicity, zero external dependencies initially, and incremental feature addition.

---

## 1. Architecture Philosophy

### Core Principles
1. **Start Minimal**: Begin with a working CLI that does one thing well
2. **Progressive Enhancement**: Add features incrementally, each building on the last
3. **Self-Contained**: Single binary with no external dependencies
4. **File-Based Storage**: JSON files for simplicity and inspectability
5. **Built-in Web UI**: Embedded HTML/CSS/JS served by Go's stdlib
6. **Unix Philosophy**: Each component does one thing well

### Why Go?
- **Single Binary**: Easy deployment and distribution
- **Excellent stdlib**: HTTP server, JSON, process management built-in
- **System Integration**: Native OS process management
- **Performance**: Minimal overhead for process monitoring
- **Concurrency**: Goroutines for real-time metrics collection
- **Cross-Platform**: Build for Linux, macOS, Windows from same codebase
- **No Runtime Dependencies**: Unlike Node.js/Python

---

## 2. Project Structure

```
vp/
├── main.go                    # Entry point, CLI interface
├── cmd/
│   ├── root.go               # Root command setup
│   ├── list.go               # List instances/templates
│   ├── start.go              # Start a process instance
│   ├── stop.go               # Stop a process instance
│   ├── template.go           # Template management commands
│   └── serve.go              # Web UI server command
├── internal/
│   ├── models/
│   │   ├── template.go       # Template data structures
│   │   ├── instance.go       # Process instance structures
│   │   └── resource.go       # Resource allocation structures
│   ├── store/
│   │   ├── store.go          # Storage interface
│   │   ├── json.go           # JSON file persistence
│   │   └── memory.go         # In-memory store (for testing)
│   ├── process/
│   │   ├── manager.go        # Process lifecycle management
│   │   ├── monitor.go        # CPU/memory monitoring
│   │   ├── interpolate.go    # Variable interpolation
│   │   └── resources.go      # Resource conflict detection
│   ├── api/
│   │   ├── server.go         # HTTP server
│   │   ├── handlers.go       # API endpoints
│   │   └── middleware.go     # Logging, CORS, etc.
│   └── ui/
│       └── embed.go          # Embedded web UI files
├── web/
│   ├── index.html            # Main dashboard page
│   ├── style.css             # Minimal CSS (or inline)
│   └── app.js                # Vanilla JavaScript (no frameworks)
├── templates/
│   ├── default/              # Default template library
│   │   ├── node-express.json
│   │   ├── postgresql.json
│   │   └── redis.json
│   └── custom/               # User-created templates
├── data/
│   ├── instances.json        # Running/stopped instances
│   ├── resources.json        # Resource allocations
│   └── logs.json             # Event logs
├── PRD.md                    # Product requirements
├── PLAN.md                   # This file
├── README.md                 # Setup and usage
├── go.mod                    # Go module file (stdlib only initially)
├── go.sum
└── Makefile                  # Build commands
```

---

## 3. Implementation Phases

### Phase 1: Minimal CLI (Day 1)

**Goal**: Working CLI that can list system processes and manage basic templates.

**Tasks:**
1. Initialize Go module (`go mod init github.com/user/vp`)
2. Create basic CLI structure using `flag` package (stdlib)
3. Implement `vp list` - list currently running system processes
4. Create Template struct with JSON marshaling
5. Implement `vp template add` - add template from JSON file
6. Implement `vp template list` - list available templates
7. File-based storage in `~/.vibeprocess/` directory

**Data Structures:**
```go
type Template struct {
    ID              string            `json:"id"`
    Label           string            `json:"label"`
    CommandTemplate string            `json:"command_template"`
    Defaults        map[string]string `json:"defaults,omitempty"`
    Variables       []string          `json:"variables,omitempty"`
    Resources       Resources         `json:"resources,omitempty"`
    Exposes         map[string]string `json:"exposes,omitempty"`
    Connections     map[string]string `json:"connections,omitempty"`
    Notes           string            `json:"notes,omitempty"`
}

type Resources struct {
    Ports []string `json:"ports,omitempty"`
    Files []string `json:"files,omitempty"`
}
```

**CLI Interface:**
```bash
vp list                          # List all running processes (system-wide)
vp template add node-express.json  # Add template from file
vp template list                 # List available templates
vp template show node-express    # Show template details
```

**Deliverables:**
- Working `vp` binary
- Can read/write templates to JSON files
- Can list system processes using `/proc` (Linux) or `ps` command
- Foundation for process management

**Code Example (main.go):**
```go
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: vp <command>")
        os.Exit(1)
    }

    switch os.Args[1] {
    case "list":
        listProcesses()
    case "template":
        handleTemplateCommands()
    default:
        fmt.Printf("Unknown command: %s\n", os.Args[1])
        os.Exit(1)
    }
}
```

### Phase 2: Instance Management (Day 2)

**Goal**: Start and stop processes from templates with variable interpolation.

**Tasks:**
1. Implement ProcessInstance struct
2. Create variable interpolation engine (${var} replacement)
3. Implement `vp start <template> <name>` - start process from template
4. Implement `vp stop <name>` - stop running instance
5. Implement `vp ps` - list managed instances
6. Store instance state in `instances.json`
7. Track PIDs and basic status (running/stopped)

**Data Structures:**
```go
type ProcessInstance struct {
    ID         string            `json:"id"`
    Name       string            `json:"name"`
    Status     string            `json:"status"` // stopped, starting, running, stopping, error
    PID        int               `json:"pid,omitempty"`
    Ports      []int             `json:"ports,omitempty"`
    TemplateID string            `json:"template_id"`
    Vars       map[string]string `json:"vars"`
    Command    string            `json:"command"`
    Error      string            `json:"error,omitempty"`
    CreatedAt  time.Time         `json:"created_at"`
    StartedAt  *time.Time        `json:"started_at,omitempty"`
    StoppedAt  *time.Time        `json:"stopped_at,omitempty"`
}
```

**CLI Interface:**
```bash
vp start node-express api-server --port=3000 --env=dev
vp stop api-server
vp ps                           # List managed instances
vp ps --all                     # Include stopped instances
vp status api-server            # Detailed instance status
```

**Key Functions:**
```go
// Interpolate variables in command template
func InterpolateCommand(template string, vars map[string]string) string {
    result := template
    for key, value := range vars {
        result = strings.ReplaceAll(result, "${"+key+"}", value)
    }
    return result
}

// Start process using os/exec
func StartProcess(cmd string) (*exec.Cmd, error) {
    parts := strings.Fields(cmd)
    process := exec.Command(parts[0], parts[1:]...)
    process.Stdout = os.Stdout
    process.Stderr = os.Stderr
    err := process.Start()
    return process, err
}
```

**Deliverables:**
- Can start processes from templates
- Variable interpolation working
- Process tracking with PIDs
- Persistent instance state
- Basic error handling

### Phase 3: Resource Management (Day 3)

**Goal**: Port allocation, conflict detection, and auto-increment counters.

**Tasks:**
1. Implement port conflict detection
2. Add auto-increment port counters (%tcpport, %vnc, %serial)
3. Implement `vp ports` - list port allocations
4. Add resource validation before starting instances
5. Track file resource usage
6. Implement `vp resources` - show all resource allocations
7. Store resource state in `resources.json`

**Data Structures:**
```go
type ResourceAllocation struct {
    Type       string    `json:"type"` // port, file
    Value      string    `json:"value"`
    InstanceID string    `json:"instance_id"`
    Instance   string    `json:"instance_name"`
    AllocatedAt time.Time `json:"allocated_at"`
}

type PortCounter struct {
    Name    string `json:"name"`    // tcpport, vnc, serial
    Current int    `json:"current"`
    Min     int    `json:"min"`
    Max     int    `json:"max"`
}
```

**CLI Interface:**
```bash
vp ports                        # List port allocations
vp ports --available            # Show available ports in ranges
vp resources                    # Show all resources (ports + files)
vp check-conflicts <instance>   # Check for resource conflicts
```

**Key Functions:**
```go
// Check if port is available on system
func IsPortAvailable(port int) bool {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return false
    }
    ln.Close()
    return true
}

// Get next available port from counter
func GetNextPort(counter *PortCounter, allocations []ResourceAllocation) int {
    for port := counter.Current; port <= counter.Max; port++ {
        if IsPortAvailable(port) && !isPortAllocated(port, allocations) {
            counter.Current = port + 1
            return port
        }
    }
    return -1
}

// Detect resource conflicts before starting
func DetectConflicts(instance *ProcessInstance, allInstances []ProcessInstance) []string {
    var conflicts []string
    for _, port := range instance.Ports {
        for _, other := range allInstances {
            if other.ID != instance.ID && other.Status == "running" {
                for _, otherPort := range other.Ports {
                    if port == otherPort {
                        conflicts = append(conflicts,
                            fmt.Sprintf("Port %d in use by %s", port, other.Name))
                    }
                }
            }
        }
    }
    return conflicts
}
```

**Enhanced Variable Interpolation:**
```go
func InterpolateWithCounters(template string, vars map[string]string,
                             counters map[string]*PortCounter,
                             allocations []ResourceAllocation) string {
    result := template

    // Standard variables: ${var}
    for key, value := range vars {
        result = strings.ReplaceAll(result, "${"+key+"}", value)
    }

    // Auto-increment counters: %tcpport
    for name, counter := range counters {
        if strings.Contains(result, "%"+name) {
            port := GetNextPort(counter, allocations)
            result = strings.ReplaceAll(result, "%"+name, fmt.Sprintf("%d", port))
        }
    }

    return result
}
```

**Deliverables:**
- Port conflict detection working
- Auto-increment port allocation
- Resource tracking
- Pre-flight validation before starting
- Clear conflict error messages

### Phase 4: Process Monitoring (Day 4)

**Goal**: Real-time CPU/memory monitoring and uptime tracking.

**Tasks:**
1. Implement process metrics collection using `/proc` or `ps`
2. Add goroutine for continuous metric updates
3. Implement `vp stats <name>` - show process metrics
4. Add uptime calculation
5. Store metric history (last 100 samples)
6. Add process health checking (detect crashes)
7. Auto-update instance status on process exit

**Data Structures:**
```go
type ProcessMetrics struct {
    CPU    float64   `json:"cpu_usage"`    // Percentage
    Memory uint64    `json:"memory_usage"` // Bytes
    Uptime int64     `json:"uptime"`       // Seconds
    Time   time.Time `json:"timestamp"`
}

type InstanceWithMetrics struct {
    *ProcessInstance
    CurrentMetrics  *ProcessMetrics   `json:"current_metrics,omitempty"`
    MetricsHistory  []ProcessMetrics  `json:"metrics_history,omitempty"`
}
```

**CLI Interface:**
```bash
vp stats api-server             # Show current metrics
vp stats api-server --watch     # Continuously update (like top)
vp stats --all                  # Show metrics for all running instances
```

**Key Functions:**
```go
// Read process stats from /proc/[pid]/stat (Linux)
func GetProcessStats(pid int) (*ProcessMetrics, error) {
    statPath := fmt.Sprintf("/proc/%d/stat", pid)
    data, err := os.ReadFile(statPath)
    if err != nil {
        return nil, err
    }

    fields := strings.Fields(string(data))
    // Parse CPU time (fields 13, 14)
    // Parse memory (field 23)
    // Calculate percentages

    return &ProcessMetrics{
        CPU:    cpuPercent,
        Memory: memoryBytes,
        Uptime: uptime,
        Time:   time.Now(),
    }, nil
}

// Background monitoring goroutine
func MonitorProcess(instance *ProcessInstance, stopChan chan bool) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if !isProcessRunning(instance.PID) {
                instance.Status = "stopped"
                instance.StoppedAt = timePtr(time.Now())
                return
            }

            metrics, err := GetProcessStats(instance.PID)
            if err == nil {
                instance.CurrentMetrics = metrics
                // Store in history (keep last 100)
            }
        case <-stopChan:
            return
        }
    }
}
```

**Deliverables:**
- Real-time CPU/memory monitoring
- Uptime tracking
- Metrics history
- Auto-detection of crashed processes
- Watch mode for live updates

### Phase 5: Web UI - Embedded Server (Day 5)

**Goal**: Basic web interface served by Go binary.

**Tasks:**
1. Create HTTP server using `net/http` stdlib
2. Implement REST API endpoints
3. Create single-page HTML dashboard
4. Add vanilla JavaScript for API calls
5. Embed web files using `embed` package
6. Implement `vp serve` - start web UI server
7. Add basic CSS for styling

**API Endpoints:**
```go
// API routes
GET  /api/instances              # List all instances
GET  /api/instances/:id          # Get instance details
POST /api/instances              # Create new instance
POST /api/instances/:id/start    # Start instance
POST /api/instances/:id/stop     # Stop instance
DELETE /api/instances/:id        # Delete instance

GET  /api/templates              # List templates
GET  /api/templates/:id          # Get template
POST /api/templates              # Create template

GET  /api/resources              # List resource allocations
GET  /api/ports                  # List port allocations

GET  /api/logs                   # Get event logs
```

**Server Implementation:**
```go
package api

import (
    "embed"
    "encoding/json"
    "net/http"
)

//go:embed web/*
var webContent embed.FS

func StartServer(addr string) error {
    // Serve embedded web UI
    http.Handle("/", http.FileServer(http.FS(webContent)))

    // API routes
    http.HandleFunc("/api/instances", handleInstances)
    http.HandleFunc("/api/templates", handleTemplates)
    http.HandleFunc("/api/resources", handleResources)

    fmt.Printf("Server starting on http://%s\n", addr)
    return http.ListenAndServe(addr, nil)
}

func handleInstances(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        instances, _ := store.ListInstances()
        json.NewEncoder(w).Encode(instances)
    case "POST":
        var req CreateInstanceRequest
        json.NewDecoder(r.Body).Decode(&req)
        instance, err := createInstance(req)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        json.NewEncoder(w).Encode(instance)
    }
}
```

**Web UI (Minimal HTML):**
```html
<!DOCTYPE html>
<html>
<head>
    <title>Vibeprocess Manager</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: system-ui; padding: 20px; }
        .tabs { display: flex; gap: 10px; margin-bottom: 20px; }
        .tab { padding: 10px 20px; cursor: pointer; border: 1px solid #ccc; }
        .tab.active { background: #007bff; color: white; }
        .instance { border: 1px solid #ddd; padding: 15px; margin: 10px 0; }
        .status { display: inline-block; padding: 4px 8px; border-radius: 3px; }
        .status.running { background: #28a745; color: white; }
        .status.stopped { background: #6c757d; color: white; }
        button { padding: 8px 16px; margin: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <h1>Vibeprocess Manager</h1>

    <div class="tabs">
        <div class="tab active" onclick="showTab('instances')">Instances</div>
        <div class="tab" onclick="showTab('templates')">Templates</div>
        <div class="tab" onclick="showTab('resources')">Resources</div>
    </div>

    <div id="instances-tab">
        <button onclick="createInstance()">+ New Instance</button>
        <div id="instances-list"></div>
    </div>

    <script>
        async function loadInstances() {
            const res = await fetch('/api/instances');
            const instances = await res.json();
            const html = instances.map(i => `
                <div class="instance">
                    <h3>${i.name}</h3>
                    <span class="status ${i.status}">${i.status}</span>
                    <p>PID: ${i.pid || 'N/A'} | Ports: ${i.ports.join(', ')}</p>
                    <button onclick="startInstance('${i.id}')">Start</button>
                    <button onclick="stopInstance('${i.id}')">Stop</button>
                </div>
            `).join('');
            document.getElementById('instances-list').innerHTML = html;
        }

        async function startInstance(id) {
            await fetch(`/api/instances/${id}/start`, { method: 'POST' });
            loadInstances();
        }

        async function stopInstance(id) {
            await fetch(`/api/instances/${id}/stop`, { method: 'POST' });
            loadInstances();
        }

        loadInstances();
        setInterval(loadInstances, 5000); // Auto-refresh
    </script>
</body>
</html>
```

**CLI Interface:**
```bash
vp serve                        # Start web UI on :8080
vp serve --port 3000            # Start on custom port
vp serve --open                 # Start and open browser
```

**Deliverables:**
- Working web UI accessible via browser
- REST API for all operations
- Embedded files (single binary deployment)
- Auto-refreshing instance list
- Start/stop buttons functional
- No external JavaScript dependencies

### Phase 6: Connection Commands (Day 6)

**Goal**: Execute connection commands from UI and CLI.

**Tasks:**
1. Implement `vp connect <instance> [command]` CLI
2. Add connection command execution in API
3. Add connection buttons to web UI
4. Support multiple connection types per instance
5. Variable interpolation in connection commands
6. Add connection history logging

**CLI Interface:**
```bash
vp connect api-server            # Show available connections
vp connect api-server browser    # Execute browser connection
vp connect api-server curl       # Execute curl connection
vp connections api-server        # List available connections
```

**Implementation:**
```go
func ExecuteConnection(instance *ProcessInstance, template *Template,
                       connType string) error {
    connCmd, ok := template.Connections[connType]
    if !ok {
        return fmt.Errorf("connection type %s not found", connType)
    }

    // Interpolate variables
    cmd := InterpolateCommand(connCmd, instance.Vars)

    // Execute command
    parts := strings.Fields(cmd)
    exec.Command(parts[0], parts[1:]...).Start()

    // Log connection
    logConnection(instance, connType)

    return nil
}
```

**Web UI Enhancement:**
```javascript
function renderConnectionButtons(instance, template) {
    const connections = Object.keys(template.connections || {});
    return connections.map(conn =>
        `<button onclick="executeConnection('${instance.id}', '${conn}')">
            ${conn}
        </button>`
    ).join('');
}

async function executeConnection(instanceId, connType) {
    await fetch(`/api/instances/${instanceId}/connect/${connType}`,
                { method: 'POST' });
}
```

**Deliverables:**
- Connection commands working from CLI
- Connection buttons in web UI
- Support for browser, curl, ssh, vnc, etc.
- Connection history tracking

### Phase 7: Logging & History (Day 7)

**Goal**: Event logging and history visualization.

**Tasks:**
1. Implement structured logging system
2. Log all instance lifecycle events
3. Log resource allocations/deallocations
4. Add `vp logs` CLI command
5. Add Logs tab to web UI
6. Implement log filtering and search
7. Add log export functionality

**Data Structures:**
```go
type LogEntry struct {
    ID         string    `json:"id"`
    Timestamp  time.Time `json:"timestamp"`
    Level      string    `json:"level"` // info, warn, error, success
    InstanceID string    `json:"instance_id,omitempty"`
    Instance   string    `json:"instance_name,omitempty"`
    Message    string    `json:"message"`
    Details    string    `json:"details,omitempty"`
}
```

**CLI Interface:**
```bash
vp logs                         # Show recent logs
vp logs --follow                # Follow logs (like tail -f)
vp logs --instance api-server   # Filter by instance
vp logs --level error           # Filter by level
vp logs --export logs.json      # Export to file
```

**Logging Functions:**
```go
func LogEvent(level, instanceID, instanceName, message string) {
    entry := LogEntry{
        ID:         generateID(),
        Timestamp:  time.Now(),
        Level:      level,
        InstanceID: instanceID,
        Instance:   instanceName,
        Message:    message,
    }

    store.AppendLog(entry)

    // Also write to stdout/stderr
    fmt.Printf("[%s] %s: %s\n", level, instanceName, message)
}

// Usage examples
LogEvent("info", inst.ID, inst.Name, "Starting process")
LogEvent("success", inst.ID, inst.Name, "Process started successfully")
LogEvent("error", inst.ID, inst.Name, "Failed to start: port conflict")
```

**Deliverables:**
- Comprehensive event logging
- Log filtering and search
- Log export
- Web UI logs tab
- Real-time log streaming

---

## 4. Data Persistence

### File Structure
```
~/.vibeprocess/
├── config.json              # Global configuration
├── templates/               # Template library
│   ├── node-express.json
│   ├── postgresql.json
│   └── custom-app.json
├── instances.json           # Active and stopped instances
├── resources.json           # Resource allocations
├── counters.json            # Port counters state
└── logs.json                # Event logs
```

### Example Files

**instances.json:**
```json
[
  {
    "id": "inst-abc123",
    "name": "api-server",
    "status": "running",
    "pid": 12345,
    "ports": [3000],
    "template_id": "node-express",
    "vars": {"port": "3000", "env": "development"},
    "command": "node server.js --port 3000 --env development",
    "created_at": "2025-11-13T10:00:00Z",
    "started_at": "2025-11-13T10:05:00Z"
  }
]
```

**counters.json:**
```json
{
  "tcpport": {"name": "tcpport", "current": 3000, "min": 3000, "max": 9999},
  "vnc": {"name": "vnc", "current": 5900, "min": 5900, "max": 5999},
  "serialport": {"name": "serialport", "current": 9600, "min": 9600, "max": 9699}
}
```

---

## 5. Build System

### Makefile
```makefile
.PHONY: build run test clean install

build:
	go build -o vp main.go

run:
	go run main.go

test:
	go test ./...

clean:
	rm -f vp
	rm -rf dist/

install:
	go install

# Build for multiple platforms
release:
	GOOS=linux GOARCH=amd64 go build -o dist/vp-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o dist/vp-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o dist/vp-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -o dist/vp-windows-amd64.exe

# Development with hot reload (using entr or air)
dev:
	find . -name "*.go" | entr -r go run main.go serve
```

### go.mod (Initial)
```go
module github.com/user/vp

go 1.21

// No dependencies initially - stdlib only
```

---

## 6. Testing Strategy

### Unit Tests
```go
// internal/process/interpolate_test.go
func TestInterpolateCommand(t *testing.T) {
    tests := []struct {
        template string
        vars     map[string]string
        expected string
    }{
        {
            "node server.js --port ${port}",
            map[string]string{"port": "3000"},
            "node server.js --port 3000",
        },
    }

    for _, tt := range tests {
        result := InterpolateCommand(tt.template, tt.vars)
        if result != tt.expected {
            t.Errorf("got %s, want %s", result, tt.expected)
        }
    }
}

// internal/process/resources_test.go
func TestDetectConflicts(t *testing.T) {
    instances := []ProcessInstance{
        {ID: "inst1", Status: "running", Ports: []int{3000}},
    }

    newInst := ProcessInstance{
        ID: "inst2", Ports: []int{3000},
    }

    conflicts := DetectConflicts(&newInst, instances)
    if len(conflicts) != 1 {
        t.Errorf("expected 1 conflict, got %d", len(conflicts))
    }
}
```

### Integration Tests
```bash
# Test full workflow
./vp template add templates/node-express.json
./vp start node-express test-server --port=3001
./vp ps | grep test-server
./vp stop test-server
```

---

## 7. Progressive Enhancement Roadmap

### Phase 8: Advanced Features (Week 2)
- [ ] Process restart on failure
- [ ] Scheduled process starts (cron-like)
- [ ] Process groups (start/stop multiple)
- [ ] Environment variable management
- [ ] Output capture and log streaming
- [ ] Health checks and auto-restart
- [ ] Systemd integration

### Phase 9: UI Enhancements (Week 3)
- [ ] Better CSS styling (consider adding Tailwind via CDN)
- [ ] Real-time metrics charts
- [ ] Dark mode toggle
- [ ] Process dependency visualization
- [ ] Drag-and-drop template editor
- [ ] In-browser terminal for process output

### Phase 10: Distribution (Week 4)
- [ ] Package for apt/yum/brew
- [ ] Docker image
- [ ] Installation script
- [ ] Auto-update mechanism
- [ ] Documentation site
- [ ] Video tutorials

---

## 8. Dependencies Justification

### Phase 1-7: Zero External Dependencies
- **stdlib only**: All features using Go standard library
- **Rationale**: Simplicity, security, minimal attack surface

### Future Considerations (Phase 8+)
When absolutely needed, consider:
- **cobra**: CLI framework (better than flag package for complex CLIs)
- **chi/mux**: HTTP router (better than stdlib for REST APIs)
- **sqlite**: Embedded database (better than JSON for large datasets)
- **websocket**: Real-time updates (better than polling)

**Decision criteria**: Only add dependency if:
1. Stdlib solution is significantly worse
2. Dependency is well-maintained and popular
3. Adds substantial value to users
4. No security concerns

---

## 9. Success Criteria

### Phase 1-3 (MVP)
- [ ] Can start/stop processes from templates
- [ ] Variable interpolation working
- [ ] Port conflict detection
- [ ] Persistent storage
- [ ] CLI intuitive and working

### Phase 4-5 (Usable)
- [ ] Web UI functional
- [ ] Real-time monitoring
- [ ] Auto-increment counters
- [ ] Resource tracking
- [ ] Single binary distribution

### Phase 6-7 (Complete)
- [ ] Connection commands working
- [ ] Comprehensive logging
- [ ] All PRD features implemented
- [ ] Documentation complete
- [ ] Ready for daily use

---

## 10. Timeline

**Week 1: Core Functionality**
- Day 1: Minimal CLI, templates
- Day 2: Instance management
- Day 3: Resource management
- Day 4: Process monitoring
- Day 5: Web UI
- Day 6: Connection commands
- Day 7: Logging, polish

**Week 2: Production Ready**
- Testing and bug fixes
- Documentation
- Example templates
- Installation scripts

**Week 3+: Enhancements**
- Advanced features
- UI improvements
- Community feedback

---

## 11. Getting Started

### Setup Commands
```bash
# Initialize project
mkdir -p ~/vp
cd ~/vp
go mod init github.com/user/vp

# Create directory structure
mkdir -p cmd internal/{models,store,process,api,ui} web templates/default data

# Create main.go
cat > main.go << 'EOF'
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Vibeprocess Manager v0.1.0")
    if len(os.Args) < 2 {
        fmt.Println("Usage: vp <command>")
        os.Exit(1)
    }
}
EOF

# Build and run
go build -o vp main.go
./vp
```

### First Template
```bash
# Create node-express template
cat > templates/default/node-express.json << 'EOF'
{
  "id": "node-express",
  "label": "Node.js Express Server",
  "command_template": "node server.js --port ${port} --env ${env}",
  "defaults": {"port": "%tcpport", "env": "development"},
  "variables": ["port", "env"],
  "resources": {
    "ports": ["${port}"],
    "files": ["server.js", "package.json"]
  },
  "exposes": {"http": ":${port}"},
  "connections": {
    "curl": "curl -I http://localhost:${port}",
    "browser": "open http://localhost:${port}"
  }
}
EOF

./vp template add templates/default/node-express.json
./vp template list
```

---

## 12. Comparison: Go vs Next.js Approach

| Aspect | Go Approach | Next.js Approach |
|--------|-------------|------------------|
| **Binary Size** | ~10MB | ~200MB (with node_modules) |
| **Startup Time** | <10ms | ~1-2s |
| **Memory Usage** | ~20MB | ~100-200MB |
| **Dependencies** | 0 (stdlib only) | ~1000 npm packages |
| **Distribution** | Single binary | npm install + node |
| **Process Management** | Native | child_process wrapper |
| **System Integration** | Direct /proc access | OS commands via exec |
| **Learning Curve** | Go basics | React + Next.js + TypeScript |
| **Deployment** | Copy binary | Deploy Node.js app |

**Verdict**: Go is significantly better aligned with PRD goals:
- "Minimal Dependencies" ✓
- "Simple Architecture" ✓
- System-level process management ✓
- Single binary distribution ✓

---

## 13. Next Steps

1. **Approve this Go-based plan**
2. **Initialize Go project**
3. **Implement Phase 1: Minimal CLI**
4. **Test with real processes**
5. **Iterate based on feedback**
6. **Expand features incrementally**

---

## Appendix A: Example Session

```bash
# Install vp
curl -sf https://example.com/install.sh | sh

# Add templates
vp template add https://example.com/templates/postgresql.json

# Create and start instance
vp start postgresql dev-db --port=5433 --data_dir=/tmp/pgdata
# Output: Started 'dev-db' (PID: 12345) on port 5433

# List running instances
vp ps
# NAME      STATUS   PID     PORTS  CPU%  MEM    UPTIME
# dev-db    running  12345   5433   2.1   128MB  00:05:23

# Connect to database
vp connect dev-db psql
# Executes: psql -h localhost -p 5433 -U postgres

# View metrics
vp stats dev-db
# CPU: 2.1%  Memory: 128MB  Uptime: 5m23s

# Stop instance
vp stop dev-db
# Output: Stopped 'dev-db' (PID: 12345)

# Start web UI
vp serve
# Output: Server running at http://localhost:8080
```

---

## Appendix B: Code Organization Principles

1. **Keep it simple**: Prefer simple solutions over clever ones
2. **Stdlib first**: Only add dependencies when absolutely necessary
3. **Errors are values**: Use Go's error handling idiomatically
4. **Small interfaces**: Define minimal, focused interfaces
5. **Table-driven tests**: Use table-driven tests for comprehensive coverage
6. **Single responsibility**: Each package/function does one thing well
7. **Documentation**: Every exported function has a doc comment
8. **Examples**: Include runnable examples in docs

---

This plan provides a clear path from a minimal working program to a full-featured process manager, while staying true to the "minimal dependencies, simple architecture" philosophy.
