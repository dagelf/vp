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

**Approach**: UI-First Development
Build the web interface early to visualize system processes, then progressively add management capabilities.

**Phase Overview:**
1. **Data Models & Process Reading** - Define structures, read from /proc
2. **Web UI & API Server** - Display current processes in browser
3. **Template Management** - Create/edit process templates (simple text format)
4. **Instance Management** - Start/stop processes from templates
5. **Resource Management** - Generic resource allocation and conflict detection
6. **Connection Commands** - One-click access to running services
7. **Logging & History** - Event tracking and visualization

---

### Phase 1: Data Models & Process Reading

**Goal**: Define data structures and read current system processes.

**Tasks:**
1. Initialize Go module (`go mod init github.com/user/vp`)
2. Define core data structures (Template, ProcessInstance, Resources)
3. Implement process discovery from `/proc` (Linux) or `ps` command
4. Parse process information (PID, command, CPU, memory)
5. Create process list data structure
6. Implement basic JSON marshaling for API responses
7. Write unit tests for process parsing

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

type ProcessInstance struct {
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Status       string            `json:"status"` // stopped, starting, running, stopping, error
    PID          int               `json:"pid,omitempty"`
    Ports        []int             `json:"ports,omitempty"`
    TemplateID   string            `json:"template_id,omitempty"`
    Vars         map[string]string `json:"vars,omitempty"`
    Command      string            `json:"command"`
    Error        string            `json:"error,omitempty"`
    CPUUsage     float64           `json:"cpu_usage,omitempty"`
    MemoryUsage  uint64            `json:"memory_usage,omitempty"`
    Uptime       int64             `json:"uptime,omitempty"`
    CreatedAt    time.Time         `json:"created_at,omitempty"`
    StartedAt    *time.Time        `json:"started_at,omitempty"`
    StoppedAt    *time.Time        `json:"stopped_at,omitempty"`
}

type Resources struct {
    // Generic resource map: resource_name -> []values
    // Examples: "port" -> ["3000", "8080"], "gpu" -> ["0"], "db_conn" -> ["mydb"]
    Allocations map[string][]string `json:"allocations,omitempty"`
}
```

**Key Functions:**
```go
// Read all processes from /proc
func ListSystemProcesses() ([]ProcessInstance, error) {
    var processes []ProcessInstance

    entries, err := os.ReadDir("/proc")
    if err != nil {
        return nil, err
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        pid, err := strconv.Atoi(entry.Name())
        if err != nil {
            continue
        }

        proc, err := readProcessInfo(pid)
        if err == nil {
            processes = append(processes, proc)
        }
    }

    return processes, nil
}

// Read process info from /proc/[pid]/
func readProcessInfo(pid int) (ProcessInstance, error) {
    // Read /proc/[pid]/cmdline for command
    // Read /proc/[pid]/stat for CPU, memory, uptime
    // Parse and return ProcessInstance
}
```

**Deliverables:**
- Core data structures defined
- Process discovery working on Linux
- Can read all running system processes
- Process metrics (CPU, memory) parsed correctly
- JSON serialization working
- Unit tests passing

### Phase 2: Web UI & API Server

**Goal**: Create web interface that displays current system processes.

**Tasks:**
1. Create HTTP server using `net/http` stdlib
2. Implement REST API endpoint: `GET /api/processes`
3. Create single-page HTML dashboard
4. Add vanilla JavaScript to fetch and display process list
5. Embed web files using `embed` package
6. Add auto-refresh every 5 seconds
7. Basic CSS styling for process table

**API Endpoints:**
```go
GET  /api/processes             # List all system processes
GET  /api/processes/:pid        # Get single process details
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
    http.HandleFunc("/api/processes", handleProcesses)

    fmt.Printf("Server starting on http://%s\n", addr)
    return http.ListenAndServe(addr, nil)
}

func handleProcesses(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    processes, err := process.ListSystemProcesses()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(processes)
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
        body { font-family: system-ui; padding: 20px; background: #f5f5f5; }
        h1 { margin-bottom: 20px; }
        table { width: 100%; background: white; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #333; color: white; font-weight: 600; }
        tr:hover { background: #f9f9f9; }
        .status { display: inline-block; padding: 4px 8px; border-radius: 3px; }
        .status.running { background: #28a745; color: white; }
        .metrics { font-family: monospace; font-size: 0.9em; }
    </style>
</head>
<body>
    <h1>Vibeprocess Manager</h1>
    <p>Showing <span id="count">0</span> running processes</p>

    <table id="processes-table">
        <thead>
            <tr>
                <th>PID</th>
                <th>Name</th>
                <th>Status</th>
                <th>CPU %</th>
                <th>Memory</th>
                <th>Uptime</th>
                <th>Command</th>
            </tr>
        </thead>
        <tbody id="processes-list"></tbody>
    </table>

    <script>
        function formatBytes(bytes) {
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024*1024) return (bytes/1024).toFixed(1) + ' KB';
            return (bytes/(1024*1024)).toFixed(1) + ' MB';
        }

        function formatUptime(seconds) {
            const h = Math.floor(seconds / 3600);
            const m = Math.floor((seconds % 3600) / 60);
            const s = seconds % 60;
            return `${h}h ${m}m ${s}s`;
        }

        async function loadProcesses() {
            try {
                const res = await fetch('/api/processes');
                const processes = await res.json();

                document.getElementById('count').textContent = processes.length;

                const html = processes.map(p => `
                    <tr>
                        <td>${p.pid}</td>
                        <td>${p.name}</td>
                        <td><span class="status running">running</span></td>
                        <td class="metrics">${(p.cpu_usage || 0).toFixed(1)}</td>
                        <td class="metrics">${formatBytes(p.memory_usage || 0)}</td>
                        <td class="metrics">${formatUptime(p.uptime || 0)}</td>
                        <td style="max-width: 400px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                            ${p.command}
                        </td>
                    </tr>
                `).join('');

                document.getElementById('processes-list').innerHTML = html;
            } catch (err) {
                console.error('Failed to load processes:', err);
            }
        }

        // Load initially
        loadProcesses();

        // Auto-refresh every 5 seconds
        setInterval(loadProcesses, 5000);
    </script>
</body>
</html>
```

**CLI Interface:**
```bash
vp serve                        # Start web UI on :8080
vp serve --port 3000            # Start on custom port
```

**Deliverables:**
- Working HTTP server
- REST API returning process list as JSON
- Web UI displaying all running processes
- Auto-refresh working
- CPU, memory, uptime displayed correctly
- Single binary with embedded web files
- Clean, readable table interface

### Phase 3: Template Management

**Goal**: Create and manage process templates with web UI.

**Tasks:**
1. Implement template storage (JSON files in `~/.vibeprocess/templates/`)
2. Add template CRUD API endpoints
3. Add "Templates" tab to web UI
4. Create template editor form
5. Load default templates (node-express, postgresql, redis)
6. Add template validation
7. Display template details and variables

**API Endpoints:**
```go
GET    /api/templates          # List all templates
GET    /api/templates/:id      # Get template by ID
POST   /api/templates          # Create new template
PUT    /api/templates/:id      # Update template
DELETE /api/templates/:id      # Delete template
```

**Storage Implementation:**
```go
type TemplateStore struct {
    baseDir string
}

func NewTemplateStore() *TemplateStore {
    homeDir, _ := os.UserHomeDir()
    baseDir := filepath.Join(homeDir, ".vibeprocess", "templates")
    os.MkdirAll(baseDir, 0755)
    return &TemplateStore{baseDir: baseDir}
}

func (s *TemplateStore) List() ([]Template, error) {
    var templates []Template
    files, err := os.ReadDir(s.baseDir)
    if err != nil {
        return nil, err
    }

    for _, file := range files {
        if filepath.Ext(file.Name()) == ".json" {
            data, _ := os.ReadFile(filepath.Join(s.baseDir, file.Name()))
            var tmpl Template
            json.Unmarshal(data, &tmpl)
            templates = append(templates, tmpl)
        }
    }
    return templates, nil
}

func (s *TemplateStore) Save(tmpl Template) error {
    data, err := json.MarshalIndent(tmpl, "", "  ")
    if err != nil {
        return err
    }
    path := filepath.Join(s.baseDir, tmpl.ID+".json")
    return os.WriteFile(path, data, 0644)
}
```

**Web UI Enhancement (Templates Tab):**
```javascript
async function loadTemplates() {
    const res = await fetch('/api/templates');
    const templates = await res.json();

    const html = templates.map(t => `
        <div class="template-card">
            <h3>${t.label}</h3>
            <p><strong>ID:</strong> ${t.id}</p>
            <p><strong>Command:</strong> <code>${t.command_template}</code></p>
            <p><strong>Variables:</strong> ${Object.keys(t.defaults || {}).join(', ')}</p>
            <button onclick="editTemplate('${t.id}')">Edit</button>
            <button onclick="deleteTemplate('${t.id}')">Delete</button>
            <button onclick="createInstance('${t.id}')">Create Instance</button>
        </div>
    `).join('');

    document.getElementById('templates-list').innerHTML = html;
}
```

**Default Templates:**
Create 3 default template files on first run:
- `node-express.json` - Node.js web server
- `postgresql.json` - PostgreSQL database
- `redis.json` - Redis cache server

**Deliverables:**
- Template CRUD operations working
- Templates persisted to JSON files
- Web UI templates tab functional
- Template editor form working
- Default templates loaded
- Template validation implemented

### Phase 4: Instance Management

**Goal**: Start and stop processes from templates with variable interpolation.

**Tasks:**
1. Implement variable interpolation engine (`${var}` replacement)
2. Add instance storage (`~/.vibeprocess/instances.json`)
3. Implement process spawning using `os/exec`
4. Add API endpoints for instance lifecycle
5. Add "Create Instance" modal in web UI
6. Implement start/stop functionality
7. Track instance state and PIDs

**API Endpoints:**
```go
GET    /api/instances              # List managed instances
GET    /api/instances/:id          # Get instance details
POST   /api/instances              # Create new instance
POST   /api/instances/:id/start    # Start instance
POST   /api/instances/:id/stop     # Stop instance
DELETE /api/instances/:id          # Delete instance
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

// Start process from template
func CreateAndStartInstance(tmpl Template, name string, vars map[string]string) (*ProcessInstance, error) {
    // Merge defaults with provided vars
    finalVars := make(map[string]string)
    for k, v := range tmpl.Defaults {
        finalVars[k] = v
    }
    for k, v := range vars {
        finalVars[k] = v
    }

    // Interpolate command
    command := InterpolateCommand(tmpl.CommandTemplate, finalVars)

    // Create instance
    instance := &ProcessInstance{
        ID:         generateID(),
        Name:       name,
        Status:     "stopped",
        TemplateID: tmpl.ID,
        Vars:       finalVars,
        Command:    command,
        CreatedAt:  time.Now(),
    }

    // Start the process
    err := StartInstance(instance)
    return instance, err
}

// Start instance process
func StartInstance(instance *ProcessInstance) error {
    instance.Status = "starting"

    // Parse command into parts
    parts := strings.Fields(instance.Command)
    cmd := exec.Command(parts[0], parts[1:]...)

    // Start process
    err := cmd.Start()
    if err != nil {
        instance.Status = "error"
        instance.Error = err.Error()
        return err
    }

    // Update instance
    instance.PID = cmd.Process.Pid
    instance.Status = "running"
    now := time.Now()
    instance.StartedAt = &now

    return nil
}

// Stop instance process
func StopInstance(instance *ProcessInstance) error {
    if instance.PID == 0 {
        return fmt.Errorf("instance not running")
    }

    instance.Status = "stopping"

    // Find and kill process
    process, err := os.FindProcess(instance.PID)
    if err != nil {
        return err
    }

    err = process.Signal(os.Interrupt)
    if err != nil {
        // Force kill if graceful shutdown fails
        process.Kill()
    }

    instance.Status = "stopped"
    now := time.Now()
    instance.StoppedAt = &now
    instance.PID = 0

    return nil
}
```

**Web UI Enhancement (Instances Tab):**
```javascript
async function createInstanceModal(templateId) {
    const template = await fetch(`/api/templates/${templateId}`).then(r => r.json());

    // Show modal with form for variables
    const form = Object.keys(template.defaults || {}).map(varName => `
        <label>${varName}: <input name="${varName}" value="${template.defaults[varName]}"></label>
    `).join('');

    // Show modal, collect input, then:
    const vars = getFormData();
    const instance = await fetch('/api/instances', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({
            template_id: templateId,
            name: vars.name,
            vars: vars
        })
    }).then(r => r.json());

    // Auto-start instance
    await fetch(`/api/instances/${instance.id}/start`, {method: 'POST'});
    loadInstances();
}
```

**Deliverables:**
- Variable interpolation working
- Process spawning functional
- Start/stop operations working
- Instance state persisted
- Web UI can create instances from templates
- Process lifecycle managed correctly

### Phase 5: Resource Management

**Goal**: Port allocation, conflict detection, and auto-increment counters.

**Tasks:**
1. Implement port conflict detection
2. Add auto-increment port counters (%tcpport, %vnc, %serial)
3. Track port allocations for instances
4. Add resource validation before starting instances
5. Implement resource conflict prevention
6. Add "Resources" tab to web UI
7. Show port allocations and availability

**Key Functions:**
```go
// Check if port is available
func IsPortAvailable(port int) bool {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return false
    }
    ln.Close()
    return true
}

// Get next available port from counter
func GetNextPort(counterName string, min, max int) (int, error) {
    // Read current counter from ~/.vibeprocess/counters/{counterName}
    current := readCounter(counterName, min)

    for port := current; port <= max; port++ {
        if IsPortAvailable(port) {
            writeCounter(counterName, port+1)
            return port, nil
        }
    }
    return -1, fmt.Errorf("no available ports in range %d-%d", min, max)
}

// Detect resource conflicts before starting
func CheckPortConflict(port int, instances []ProcessInstance) error {
    for _, inst := range instances {
        if inst.Status == "running" {
            for _, p := range inst.Ports {
                if p == port {
                    return fmt.Errorf("port %d already in use by %s", port, inst.Name)
                }
            }
        }
    }
    return nil
}

// Auto-increment variable interpolation
func InterpolateWithCounters(template string, vars map[string]string) string {
    result := template

    // Standard variables: ${var}
    for key, value := range vars {
        result = strings.ReplaceAll(result, "${"+key+"}", value)
    }

    // Auto-increment counters: %tcpport, %vnc, %serialport
    if strings.Contains(result, "%tcpport") {
        port, _ := GetNextPort("tcpport", 3000, 9999)
        result = strings.ReplaceAll(result, "%tcpport", fmt.Sprintf("%d", port))
    }
    if strings.Contains(result, "%vnc") {
        port, _ := GetNextPort("vnc", 5900, 5999)
        result = strings.ReplaceAll(result, "%vnc", fmt.Sprintf("%d", port))
    }
    if strings.Contains(result, "%serialport") {
        port, _ := GetNextPort("serialport", 9600, 9699)
        result = strings.ReplaceAll(result, "%serialport", fmt.Sprintf("%d", port))
    }

    return result
}
```

**Deliverables:**
- Port conflict detection working
- Auto-increment port counters (%tcpport, %vnc, %serialport)
- Resource tracking and validation
- Pre-flight checks before starting
- Resources tab in web UI

### Phase 6: Connection Commands

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

### Phase 7: Logging & History

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

**Philosophy**:
- **Configuration = Simple Text Files** (easily editable, script-friendly, convention over config)
- **State/History = JSON** (structured data, makes code leaner)

### File Structure
```
~/.vibeprocess/
├── templates/               # Template library (simple text format)
│   ├── node-express.tpl
│   ├── postgresql.tpl
│   └── custom-app.tpl
├── counters/                # Port counter values (plain text files)
│   ├── tcpport              # Contains just the number: "3001"
│   ├── vnc                  # Contains just the number: "5901"
│   └── serialport           # Contains just the number: "9601"
├── instances.json           # Active and stopped instances (JSON state)
├── logs.json                # Event logs (JSON history)
└── config                   # Global config (KEY=VALUE format)
```

### Template Format (Simple Text)

**~/.vibeprocess/templates/node-express.tpl:**
```
id: node-express
label: Node.js Express Server
command: node server.js --port ${port} --env ${env}

[defaults]
port = %tcpport
env = development

[variables]
port
env

[resources]
port = ${port}
file = server.js
file = package.json

[exposes]
http = :${port}

[connections]
curl = curl -I http://localhost:${port}
browser = open http://localhost:${port}
```

**Simple parsing rules:**
- Lines with `key: value` or `key = value` are properties
- `[section]` headers denote sections
- Multiple lines with same key create a list (e.g., multiple `file =` lines)
- Blank lines ignored
- `#` for comments

**Generic Resources:**
- Resources are not opinionated - any resource type can be defined
- Format: `resource_type = value`
- Examples: `port = 3000`, `file = data.db`, `gpu = 0`, `db_connection = mydb`
- Multiple values: repeat the key on new lines
- This allows extensibility for any resource type you need to track

### Counter Files (Plain Text)

**~/.vibeprocess/counters/tcpport:**
```
3001
```

**~/.vibeprocess/counters/vnc:**
```
5901
```

Simple integer, one per file. Easy to read/write with scripts.

### State Files (JSON for Lean Code)

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

### Config File (KEY=VALUE)

**~/.vibeprocess/config:**
```
WEB_PORT=8080
AUTO_START=false
LOG_LEVEL=info
DEFAULT_PORT_MIN=3000
DEFAULT_PORT_MAX=9999
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

## 10. Implementation Sequence

**Core Functionality (Sequential Phases)**
Execute these phases in order. Each phase builds on the previous:

1. Data models & process reading from /proc
2. Web UI & API server showing current processes
3. Template management (simple text format)
4. Instance management (start/stop from templates)
5. Resource management (generic resources, conflicts, counters)
6. Connection commands (click to connect)
7. Logging, history, polish

**Post-Core Enhancement Phases** (Optional, as needed)
- Phase 8: Advanced features (health checks, auto-restart, process groups)
- Phase 9: UI enhancements (better styling, charts, dark mode)
- Phase 10: Distribution (packaging, installation scripts, docs)

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
# Create node-express template (simple text format)
mkdir -p ~/.vibeprocess/templates
cat > ~/.vibeprocess/templates/node-express.tpl << 'EOF'
id: node-express
label: Node.js Express Server
command: node server.js --port ${port} --env ${env}

[defaults]
port = %tcpport
env = development

[variables]
port
env

[resources]
port = ${port}
file = server.js
file = package.json

[exposes]
http = :${port}

[connections]
curl = curl -I http://localhost:${port}
browser = open http://localhost:${port}
EOF

# Start the web UI
./vp serve
# Open browser to http://localhost:8080
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
