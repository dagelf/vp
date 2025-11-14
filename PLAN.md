# Vibeprocess Manager - Implementation Plan

## Ultra-Lean Firmware-Style Architecture

This document outlines the implementation plan for Vibeprocess Manager using a radically simplified, zero-assumption architecture. The design philosophy: **pure mechanism, no policy** - like firmware that provides primitives and lets users configure everything.

---

## 1. Architecture Philosophy

### Core Principles
1. **Zero Hardcoded Assumptions**: No opinions about what resources are or how they work
2. **Everything is a Command**: Resource validation via shell commands
3. **Six Files Total**: Brutally simple structure
4. **Generic Resource System**: Resources are just type:value pairs with check commands
5. **State as JSON**: Human-readable, debuggable persistence
6. **Firmware-Style**: Provide primitives (allocate, start, stop, persist), users configure behavior

### Why This Approach?
- **Maximum Flexibility**: Want GPU allocation? Add a resource type. Want license servers? Add a check command.
- **No Special Cases**: Ports, files, GPUs, databases - all just resources with check commands
- **Debuggable**: Everything is data (JSON state) + commands (shell)
- **Extensible**: Add resource types at runtime without code changes
- **Minimal Code**: ~500 lines total for core functionality
- **Zero Dependencies**: Go stdlib only

---

## 2. Project Structure

```
vp/
├── main.go          # CLI entry point (~80 lines)
├── state.go         # State persistence (~100 lines)
├── process.go       # Process lifecycle (~150 lines)
├── resource.go      # Generic resource system (~100 lines)
├── api.go           # HTTP server (~70 lines)
└── web.html         # Embedded UI (single page)
```

**6 files. ~500 lines total.**

---

## 3. Core Data Structures

### state.go
```go
type State struct {
    Instances  map[string]*Instance   `json:"instances"`   // name -> Instance
    Templates  map[string]*Template   `json:"templates"`   // id -> Template
    Resources  map[string]*Resource   `json:"resources"`   // type:value -> Resource
    Counters   map[string]int         `json:"counters"`    // counter_name -> current
    Types      map[string]*ResourceType `json:"types"`     // Resource type definitions
}

func (s *State) Load() error {
    // Read from ~/.vibeprocess/state.json
}

func (s *State) Save() error {
    // Write to ~/.vibeprocess/state.json
}
```

### process.go
```go
type Instance struct {
    Name       string                 `json:"name"`
    Template   string                 `json:"template"`
    Command    string                 `json:"command"`      // Final interpolated command
    PID        int                    `json:"pid"`
    Status     string                 `json:"status"`       // stopped|starting|running|stopping
    Resources  map[string]string      `json:"resources"`    // resource_type -> value
    Started    int64                  `json:"started"`      // Unix timestamp
}

type Template struct {
    ID         string                 `json:"id"`
    Label      string                 `json:"label,omitempty"`
    Command    string                 `json:"command"`      // Template with ${var} and %counter
    Resources  []string               `json:"resources"`    // Resource types this needs
    Vars       map[string]string      `json:"vars"`         // Default variables
}

func StartProcess(template *Template, name string, vars map[string]string) (*Instance, error)
func StopProcess(inst *Instance) error
```

### resource.go
```go
type Resource struct {
    Type       string                 `json:"type"`         // tcpport|vncport|dbfile|gpu|whatever
    Value      string                 `json:"value"`        // "3000" or "/var/db/mydb" or "0"
    Owner      string                 `json:"owner"`        // Instance name
}

type ResourceType struct {
    Name       string                 `json:"name"`         // tcpport, vncport, dbfile, gpu
    Check      string                 `json:"check"`        // Shell command to check availability
    Counter    bool                   `json:"counter"`      // Is this auto-incrementing?
    Start      int                    `json:"start"`        // Counter start value
    End        int                    `json:"end"`          // Counter end value
}

func AllocateResource(rtype string, requestedValue string) (string, error)
func CheckResource(rt *ResourceType, value string) bool
func (s *State) ClaimResource(rtype, value, owner string)
func (s *State) ReleaseResources(owner string)
```

### api.go
```go
func ServeHTTP(addr string) error

// Endpoints:
// GET  /api/instances
// POST /api/instances/start
// POST /api/instances/stop
// GET  /api/templates
// POST /api/templates
// GET  /api/resources
// POST /api/resource-types
```

### main.go
```go
func main() {
    // Ultra-minimal CLI dispatcher
    // Commands: start, stop, ps, serve, template, resource-type
}
```

---

## 4. The Genius: Generic Resource System

### Built-in Resource Types (Defaults)

```go
var DefaultResourceTypes = map[string]*ResourceType{
    "tcpport": {
        Name:    "tcpport",
        Check:   "nc -z localhost ${value} && exit 1 || exit 0",  // Fail if in use
        Counter: true,
        Start:   3000,
        End:     9999,
    },
    "vncport": {
        Name:    "vncport",
        Check:   "nc -z localhost ${value} && exit 1 || exit 0",
        Counter: true,
        Start:   5900,
        End:     5999,
    },
    "serialport": {
        Name:    "serialport",
        Check:   "nc -z localhost ${value} && exit 1 || exit 0",
        Counter: true,
        Start:   9600,
        End:     9699,
    },
    "dbfile": {
        Name:    "dbfile",
        Check:   "test -f ${value} && exit 1 || exit 0",  // Fail if exists
        Counter: false,
    },
    "socket": {
        Name:    "socket",
        Check:   "test -S ${value} && exit 1 || exit 0",  // Fail if socket exists
        Counter: false,
    },
}
```

### Adding Custom Resource Types at Runtime

```bash
# Add GPU resource type
vp resource-type add gpu \
  --check='nvidia-smi -L | grep GPU-${value}' \
  --counter=false

# Add license server connection
vp resource-type add license \
  --check='lmutil lmstat -c ${value} | grep "UP"' \
  --counter=false
```

### Resource Allocation Algorithm

```go
func AllocateResource(rtype string, requestedValue string) (string, error) {
    rt := state.Types[rtype]
    if rt == nil {
        return "", fmt.Errorf("unknown resource type: %s", rtype)
    }

    var value string

    if rt.Counter && requestedValue == "" {
        // Auto-increment counter
        current := state.Counters[rtype]
        if current == 0 {
            current = rt.Start
        }

        for v := current; v <= rt.End; v++ {
            value = strconv.Itoa(v)
            if CheckResource(rt, value) {
                state.Counters[rtype] = v + 1
                break
            }
        }
    } else {
        // Explicit value requested
        value = requestedValue
        if !CheckResource(rt, value) {
            return "", fmt.Errorf("%s %s not available", rtype, value)
        }
    }

    return value, nil
}

func CheckResource(rt *ResourceType, value string) bool {
    if rt.Check == "" {
        return true  // No check command = always available
    }

    // Interpolate check command
    check := strings.ReplaceAll(rt.Check, "${value}", value)

    // Execute check
    cmd := exec.Command("sh", "-c", check)
    err := cmd.Run()
    return err == nil  // Check command should exit 0 if available
}
```

---

## 5. Implementation Phases

### Phase 1: State & Resource System

**Goal**: Core resource allocation without any processes

**Files**: `state.go`, `resource.go`

**Tasks**:
1. Implement State struct with JSON persistence
2. Implement ResourceType and Resource structs
3. Implement AllocateResource with counter support
4. Implement CheckResource with shell command execution
5. Add default resource types (tcpport, vncport, serialport, dbfile, socket)
6. Test resource allocation and conflict detection

**Deliverables**:
- Can allocate/deallocate resources
- Counter resources auto-increment
- Check commands validate availability
- State persists to ~/.vibeprocess/state.json

### Phase 2: Process Management

**Goal**: Start/stop processes with resource allocation

**Files**: `process.go`, updates to `state.go`

**Tasks**:
1. Implement Template and Instance structs
2. Implement StartProcess with 3 phases:
   - Phase 1: Allocate resources
   - Phase 2: Interpolate command (${var} and %counter)
   - Phase 3: Start process
3. Implement StopProcess with resource cleanup
4. Add process tracking (PID, status)
5. Test process lifecycle

**Deliverables**:
- Can start processes from templates
- Resources allocated automatically
- Variables interpolated correctly
- Processes tracked with PIDs
- Resources released on stop

### Phase 3: CLI Interface

**Goal**: Minimal command-line interface

**Files**: `main.go`

**Tasks**:
1. Implement CLI dispatcher
2. Add commands:
   - `vp start <template> <name> [--key=value...]`
   - `vp stop <name>`
   - `vp ps`
   - `vp template add <file>`
   - `vp resource-type add <name> --check=<cmd> [--counter] [--start=N] [--end=N]`
3. Parse --key=value arguments
4. Pretty-print output

**CLI Examples**:
```bash
# Start with auto-allocated port
vp start postgres db1

# Start with explicit port
vp start postgres db2 --tcpport=5433 --datadir=/var/db2

# List instances
vp ps

# Add custom resource type
vp resource-type add gpu --check='nvidia-smi -L | grep GPU-${value}'

# Start with GPU
vp start ml-job training --gpu=0
```

**Deliverables**:
- Working CLI with all commands
- Argument parsing
- Human-readable output
- Error handling

### Phase 4: Web UI

**Goal**: Single-page web interface

**Files**: `api.go`, `web.html`

**Tasks**:
1. Implement HTTP server with embedded HTML
2. Add API endpoints:
   - GET /api/instances
   - POST /api/instances/start
   - POST /api/instances/stop
   - GET /api/templates
   - POST /api/templates
   - GET /api/resources
   - POST /api/resource-types
3. Create single-page HTML dashboard with:
   - Instance list with start/stop buttons
   - Template manager
   - Resource viewer
   - Resource type editor
4. Auto-refresh every 5 seconds
5. Embed web.html using go:embed

**Web UI Features**:
- Table of instances (name, status, PID, command, resources)
- Start/stop buttons
- Template list with "Create Instance" button
- Resource allocation table
- Add custom resource types via form

**Deliverables**:
- Working web UI on localhost:8080
- Can start/stop instances from browser
- Can view resource allocations
- Can add templates and resource types
- Single binary with embedded HTML

### Phase 5: Polish & Testing

**Goal**: Production-ready tool

**Tasks**:
1. Add comprehensive error handling
2. Add logging to ~/.vibeprocess/logs.json
3. Write unit tests for:
   - Resource allocation
   - Command interpolation
   - Process lifecycle
4. Add example templates
5. Write README with examples
6. Test on Linux/macOS
7. Create installation script

**Deliverables**:
- All tests passing
- Robust error handling
- Example templates included
- Documentation complete
- Installation script working

---

## 6. Template Examples

### PostgreSQL
```json
{
  "id": "postgres",
  "label": "PostgreSQL Database",
  "command": "postgres -D ${datadir} -p ${tcpport}",
  "resources": ["tcpport", "datadir"],
  "vars": {
    "datadir": "/tmp/pgdata"
  }
}
```

### QEMU Virtual Machine
```json
{
  "id": "qemu",
  "label": "QEMU Virtual Machine",
  "command": "qemu-system-x86_64 -vnc :${vncport} -serial tcp::${serialport},server,nowait ${args}",
  "resources": ["vncport", "serialport"],
  "vars": {
    "args": "-m 2G"
  }
}
```

### Node.js Server
```json
{
  "id": "node-express",
  "label": "Node.js Express Server",
  "command": "node server.js --port ${tcpport}",
  "resources": ["tcpport"],
  "vars": {}
}
```

### ML Training Job (Custom GPU Resource)
```json
{
  "id": "ml-training",
  "label": "ML Training Job",
  "command": "python train.py --gpu ${gpu} --data ${datadir}",
  "resources": ["gpu", "datadir"],
  "vars": {
    "datadir": "/data/training"
  }
}
```

---

## 7. Usage Examples

### Basic Usage
```bash
# Start instance with auto-allocated resources
vp start postgres mydb

# Start with explicit resource values
vp start postgres mydb --tcpport=5432 --datadir=/var/db/mydb

# Mix explicit and auto
vp start qemu vm1 --vncport=5901  # serialport auto-allocated

# List instances
vp ps

# Stop instance
vp stop mydb
```

### Adding Custom Resource Types
```bash
# Add GPU resource type
vp resource-type add gpu \
  --check='nvidia-smi -L | grep GPU-${value}' \
  --counter=false

# Add license server connection
vp resource-type add flexlm \
  --check='lmutil lmstat -c ${value} 2>&1 | grep "UP"' \
  --counter=false

# Add database connection
vp resource-type add dbconn \
  --check='psql -h ${value} -c "SELECT 1" 2>&1 | grep "1 row"' \
  --counter=false
```

### Using Custom Resources
```bash
# Now templates can use GPU resources
vp start ml-job training --gpu=0

# Use license server
vp start matlab session1 --flexlm=27000@licserver

# Use database connection
vp start webapp api --dbconn=postgres://localhost:5432/mydb
```

### Web UI
```bash
# Start web server
vp serve

# Open browser to http://localhost:8080
# - View all instances
# - Start/stop with buttons
# - Add templates via form
# - Add resource types via form
# - View resource allocations
```

---

## 8. Key Implementation Details

### Resource Allocation Flow

```go
func StartProcess(template *Template, name string, vars map[string]string) (*Instance, error) {
    inst := &Instance{
        Name:      name,
        Template:  template.ID,
        Status:    "starting",
        Resources: make(map[string]string),
    }

    // Phase 1: Allocate resources declared in template
    for _, rtype := range template.Resources {
        value, err := AllocateResource(rtype, vars[rtype])
        if err != nil {
            // Rollback all allocated resources
            state.ReleaseResources(name)
            return nil, err
        }
        inst.Resources[rtype] = value
        state.ClaimResource(rtype, value, name)
        vars[rtype] = value  // Make available for interpolation
    }

    // Phase 2: Interpolate command
    cmd := template.Command

    // Replace ${var} syntax
    for key, val := range vars {
        cmd = strings.ReplaceAll(cmd, "${"+key+"}", val)
    }

    // Handle %counter syntax (auto-allocate if not already allocated)
    for {
        match := regexp.MustCompile(`%(\w+)`).FindStringSubmatch(cmd)
        if match == nil {
            break
        }
        counter := match[1]
        value, _ := AllocateResource(counter, "")
        cmd = strings.ReplaceAll(cmd, "%"+counter, value)
        inst.Resources[counter] = value
        state.ClaimResource(counter, value, name)
    }

    inst.Command = cmd

    // Phase 3: Start process
    parts := strings.Fields(cmd)
    proc := exec.Command(parts[0], parts[1:]...)
    if err := proc.Start(); err != nil {
        state.ReleaseResources(name)
        return nil, err
    }

    inst.PID = proc.Process.Pid
    inst.Status = "running"
    inst.Started = time.Now().Unix()

    state.Instances[name] = inst
    state.Save()

    return inst, nil
}
```

### State Persistence

```go
// state.go
var state *State

func LoadState() *State {
    homeDir, _ := os.UserHomeDir()
    stateFile := filepath.Join(homeDir, ".vibeprocess", "state.json")

    data, err := os.ReadFile(stateFile)
    if err != nil {
        // Initialize with defaults
        return &State{
            Instances: make(map[string]*Instance),
            Templates: make(map[string]*Template),
            Resources: make(map[string]*Resource),
            Counters:  make(map[string]int),
            Types:     DefaultResourceTypes,
        }
    }

    var s State
    json.Unmarshal(data, &s)

    // Merge with default types
    if s.Types == nil {
        s.Types = make(map[string]*ResourceType)
    }
    for name, rt := range DefaultResourceTypes {
        if s.Types[name] == nil {
            s.Types[name] = rt
        }
    }

    return &s
}

func (s *State) Save() error {
    homeDir, _ := os.UserHomeDir()
    stateDir := filepath.Join(homeDir, ".vibeprocess")
    os.MkdirAll(stateDir, 0755)

    stateFile := filepath.Join(stateDir, "state.json")
    data, _ := json.MarshalIndent(s, "", "  ")
    return os.WriteFile(stateFile, data, 0644)
}
```

### Main CLI

```go
// main.go
package main

import (
    "fmt"
    "os"
    "strings"
)

var state *State

func main() {
    state = LoadState()
    defer state.Save()

    if len(os.Args) < 2 {
        listInstances()
        return
    }

    cmd := os.Args[1]
    args := os.Args[2:]

    switch cmd {
    case "start":
        // vp start <template> <name> [--key=value...]
        if len(args) < 2 {
            fmt.Fprintf(os.Stderr, "Usage: vp start <template> <name> [--key=value...]\n")
            os.Exit(1)
        }

        template := state.Templates[args[0]]
        if template == nil {
            fmt.Fprintf(os.Stderr, "Template not found: %s\n", args[0])
            os.Exit(1)
        }

        name := args[1]
        vars := parseVars(args[2:])

        // Merge template defaults
        for k, v := range template.Vars {
            if vars[k] == "" {
                vars[k] = v
            }
        }

        inst, err := StartProcess(template, name, vars)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("Started %s (PID %d)\n", inst.Name, inst.PID)
        fmt.Printf("Resources: %v\n", inst.Resources)

    case "stop":
        if len(args) < 1 {
            fmt.Fprintf(os.Stderr, "Usage: vp stop <name>\n")
            os.Exit(1)
        }

        name := args[0]
        inst := state.Instances[name]
        if inst == nil {
            fmt.Fprintf(os.Stderr, "Instance not found: %s\n", name)
            os.Exit(1)
        }

        if err := StopProcess(inst); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        state.ReleaseResources(name)
        delete(state.Instances, name)
        fmt.Printf("Stopped %s\n", name)

    case "ps":
        listInstances()

    case "serve":
        port := "8080"
        if len(args) > 0 {
            port = args[0]
        }
        fmt.Printf("Starting web UI on http://localhost:%s\n", port)
        ServeHTTP(":" + port)

    case "template":
        // vp template add <file>
        // vp template list
        handleTemplateCommand(args)

    case "resource-type":
        // vp resource-type add <name> --check=<cmd> [--counter] [--start=N] [--end=N]
        // vp resource-type list
        handleResourceTypeCommand(args)

    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
        fmt.Fprintf(os.Stderr, "Commands: start, stop, ps, serve, template, resource-type\n")
        os.Exit(1)
    }
}

func parseVars(args []string) map[string]string {
    vars := make(map[string]string)
    for _, arg := range args {
        if strings.HasPrefix(arg, "--") {
            parts := strings.SplitN(arg[2:], "=", 2)
            if len(parts) == 2 {
                vars[parts[0]] = parts[1]
            }
        }
    }
    return vars
}

func listInstances() {
    if len(state.Instances) == 0 {
        fmt.Println("No instances running")
        return
    }

    fmt.Printf("%-20s %-10s %-8s %-40s %s\n", "NAME", "STATUS", "PID", "COMMAND", "RESOURCES")
    for name, inst := range state.Instances {
        resources := ""
        for k, v := range inst.Resources {
            resources += fmt.Sprintf("%s=%s ", k, v)
        }
        fmt.Printf("%-20s %-10s %-8d %-40s %s\n",
            name, inst.Status, inst.PID, truncate(inst.Command, 40), resources)
    }
}

func truncate(s string, n int) string {
    if len(s) <= n {
        return s
    }
    return s[:n-3] + "..."
}
```

---

## 9. Why This Design is Genius

### 1. Zero Hardcoded Assumptions
- Resources are not hardcoded (ports, files, etc.)
- Just type:value pairs with optional check commands
- Add any resource type at runtime

### 2. Maximum Flexibility
- Want GPU allocation? Add resource type with nvidia-smi check
- Want license servers? Add resource type with lmutil check
- Want database connections? Add resource type with psql check
- Want anything? Just define a check command

### 3. Validation via Shell Commands
- No special-purpose validation logic
- Check commands can be arbitrarily complex
- Use any installed tool (nc, test, nvidia-smi, lmutil, etc.)
- Full shell syntax available

### 4. Counter Resources Not Special-Cased
- Counter is just a boolean flag on resource type
- Same allocation mechanism as non-counter resources
- Counter state persisted like everything else

### 5. Brutally Simple
- 6 files total
- ~500 lines of code
- Zero dependencies
- Single JSON state file
- All behavior configurable via data

### 6. Firmware-Style Design
Provides primitives:
- **Memory allocation** (resources)
- **Process control** (start/stop)
- **State persistence** (JSON)
- **Variable interpolation** (string replacement)
- **Validation** (shell commands)

Everything else is user configuration.

### 7. Debuggable
- All state in human-readable JSON
- Check commands are just shell scripts
- Can manually edit state file
- Can test check commands independently

### 8. Extensible Without Code Changes
- Add resource types via CLI
- Add templates via CLI
- Modify check commands in state.json
- No recompilation needed

---

## 10. Implementation Sequence

Execute these phases in order:

1. **State & Resource System** - Core allocation without processes
2. **Process Management** - Start/stop with resource allocation
3. **CLI Interface** - Minimal command-line tool
4. **Web UI** - Single-page interface
5. **Polish & Testing** - Production-ready

Each phase builds on the previous. Total implementation: ~500 lines of Go.

---

## 11. Success Criteria

### Phase 1 Complete
- [ ] Can allocate/deallocate resources
- [ ] Counter resources auto-increment
- [ ] Check commands execute correctly
- [ ] State persists to JSON

### Phase 2 Complete
- [ ] Can start processes from templates
- [ ] Resources allocated automatically
- [ ] Variables interpolated (${var} and %counter)
- [ ] Processes tracked with PIDs
- [ ] Resources released on stop

### Phase 3 Complete
- [ ] CLI commands work (start, stop, ps, template, resource-type)
- [ ] Argument parsing correct
- [ ] Error handling robust
- [ ] Output human-readable

### Phase 4 Complete
- [ ] Web UI accessible on localhost:8080
- [ ] Can start/stop instances from browser
- [ ] Can view resource allocations
- [ ] Can add templates and resource types
- [ ] Single binary with embedded HTML

### Phase 5 Complete
- [ ] All tests passing
- [ ] Error handling comprehensive
- [ ] Example templates included
- [ ] Documentation complete
- [ ] Ready for daily use

---

## 12. Comparison to Traditional Approach

| Aspect | Traditional | Ultra-Lean |
|--------|------------|------------|
| **Lines of Code** | ~5000 | ~500 |
| **Files** | 30+ | 6 |
| **Resource Types** | Hardcoded | User-defined |
| **Validation** | Custom Go code | Shell commands |
| **Extensibility** | Requires coding | Just add data |
| **State Storage** | Multiple files | Single JSON |
| **Dependencies** | Several | Zero |
| **Assumptions** | Many | None |

---

## 13. The Point

This design has **zero opinions** about:
- What a resource is (just a string)
- How to check availability (just a command)
- What counters count (just incrementing numbers)
- What templates do (just commands with variables)

It's pure mechanism, no policy. Like firmware that just provides primitives.

Want GPU allocation? Add a resource type.
Want license servers? Add a resource type.
Want database connections? Add a resource type.
Want anything? Add a check command.

**This is how you design for Mars - assume nothing, enable everything.**
