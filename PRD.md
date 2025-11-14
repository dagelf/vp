# Product Requirements Document (PRD)

## Vibeprocess Manager

Vibeprocess Manager is a firmware-style process orchestration tool that provides pure primitives for process and resource management. It makes **zero assumptions** about what resources are or how they work - everything is user-defined through simple check commands. Think the Unix `ps` tool, modernized, with generic resource allocation and zero opinions.

### Current Pain Points

Modern servers often experience process creep: some processes are started by systemd, some in screen, and many manually. When the server experiences a failure or reboot, a lot of state goes missing - and lore about how and what to start and what was running drifts.

1. **Manual Process Management**: Operators must manually start/stop processes in multiple terminal windows, making it difficult to track what's running and where
2. **Resource Conflicts**: Port collisions and file conflicts are common when running multiple instances or switching between projects in various states of development and production and update.
3. **Configuration Complexity**: Each process requires specific environment variables, ports, and file paths that must be remembered and typed manually
4. **Lack of Visibility**: No centralized view of running processes, resource usage, or allocated ports
5. **Onboarding Friction**: New team members struggle to understand which processes to run and how to configure them
6. **Connection Management**: Accessing services requires remembering connection strings and CLI commands

## Product Vision

Create a unified interface that transforms process management from a manual, error-prone task into an automated, visual experience - making orchestration as easy as clicking "Start" and as safe as guardrails that prevent resource conflicts.

## Target Users

### Primary Users
- **Managed Service Providers**: Rapidly assessing what's running, how it was started, and what to maintain
- **Full-stack Developers**: Managing web servers, databases, and APIs simultaneously
- **Backend Engineers**: Running multiple microservices and data stores
- **DevOps Engineers**: Testing infrastructure configurations locally

### Secondary Users
- **Frontend Developers**: Starting backend services for local development
- **Junior Developers**: Learning how services interconnect
- **Technical Leads**: Standardizing team development environments

---

## Core Features

### 1. Process Templates

**Description**: Reusable configurations defining how to start specific types of processes.

**Key Capabilities**:
- Define command templates with variable interpolation (eg., `node server.js --port ${port}`)
- Specify default values for variables
- Support auto-incrementing counters (eg. `%tcpport`, `%vnc`, `%serialport`)
- Test before run option
- Declare resource requirements and exposed endpoints (ports, files) - Purely declarative
- Define connection commands for quick access
- Add notes/documentation per template and what will become default for instance
- Integrated AI recommendations and tests (eg. `Try bun dev --port ${port}"`)
- Default AI recommendation instruction particular to this template

**User Value**: Write once, use many times and anywhere. Teams can share templates to ensure consistent configurations.

### 2. Process Instances

**Description**: Named, configured instances of templates that can be started/stopped independently.

**Key Capabilities**:
- Create instances from templates with custom variable values
- Match running processes on the host system to existing templates - or create new templates based on the suggestions
- Edit instance configuration before or after creation
- Start/stop processes with visual status indicators
- View real-time status (stopped, starting, running (eg busy checking resource availability), stopping (stopped but still shutting down, ability to force kill), error)
- Track process metadata (PID, ports, uptime, command, cgroups)
- Add notes/documentation per instance

**User Value**: Manage multiple configurations of the same service (e.g., "Dev DB" vs "Test DB") without manual reconfiguration.

### 3. Generic Resource Management

**Description**: Zero-assumption resource tracking where resources are just type:value pairs validated by user-defined shell commands.

**Core Philosophy**:
- **No hardcoded resource types** - Ports, files, GPUs, licenses, databases are all just resources
- **Validation via commands** - Check availability using any shell command (nc, test, nvidia-smi, lmutil, etc.)
- **Counter resources** - Auto-increment counters (just a flag on resource type)
- **User-extensible** - Add new resource types at runtime without code changes

**Key Capabilities**:
- Define custom resource types with check commands
- Auto-detect conflicts before starting processes (via check commands)
- Auto-increment counters for sequential allocation (ports, VNC displays, etc.)
- Track arbitrary resources (ports, files, GPUs, licenses, DB connections, etc.)
- Filter resources by status (all, in use, available)
- Show resource allocation history

**Built-in Resource Types** (examples, not limitations):
- `tcpport` - TCP ports (check via `nc`)
- `vncport` - VNC ports (check via `nc`)
- `serialport` - Serial ports (check via `nc`)
- `dbfile` - Database files (check via `test -f`)
- `socket` - Unix sockets (check via `test -S`)

**Custom Resource Examples**:
- GPU allocation: `nvidia-smi -L | grep GPU-${value}`
- License servers: `lmutil lmstat -c ${value} | grep "UP"`
- Database connections: `psql -h ${value} -c "SELECT 1"`

**User Value**: Manage ANY resource type without coding. Add GPU allocation? Just define a check command. Need license tracking? Add a resource type. Works for anything.

### 4. Performance Monitoring

**Description**: Real-time visibility into running process health.

**Key Capabilities**:
- Display CPU usage, CPU history per process
- Show memory consumption, memory history
- Track process uptime, stop and start logs
- Auto-refresh metrics every n seconds
- Visual progress bars for resource utilization
- Sort by any metric
- Compact overview that can expand for details

**User Value**: Quickly identify resource-intensive processes and potential issues.

### 5. Connection Management

**Description**: One-click access to running services.

**Key Capabilities**:
- Click running instances to execute connection commands (eg. launch vncviewer $ip:$port)
- Support multiple connection types per template (CLI, browser, curl, ab, etc.)
- Variable interpolation in connection commands
- Connection selection dialog for multi-connection services

**User Value**: Access services instantly without remembering connection strings or commands.

### 6. Visual Dashboard

**Description**: Modern, intuitive interface for all process management tasks.

**Key Capabilities**:
- Tabbed interface (Instances, Templates, Resources, Logs)
- Color-coded status badges
- Real-time status updates
- Responsive design for various screen sizes
- Dark mode support (via theme provider)
- Modal dialogs for creation/editing workflows

**User Value**: See everything at a glance. No terminal juggling required.

---

## Technical Architecture

### Technology Stack

**Ultra-Lean Firmware-Style**:
- **Language**: Go (stdlib only, zero dependencies)
- **Structure**: 6 files, ~500 lines total
- **Storage**: Single JSON state file (~/.vibeprocess/state.json)
- **Philosophy**: Pure mechanism, no policy

### Project Structure
```
vp/
├── main.go          # CLI entry point (~80 lines)
├── state.go         # State persistence (~100 lines)
├── process.go       # Process lifecycle (~150 lines)
├── resource.go      # Generic resource system (~100 lines)
├── api.go           # HTTP server (~70 lines)
└── web.html         # Embedded UI (single page)
```

### Core Data Structures

**State** (state.go):
```go
type State struct {
    Instances  map[string]*Instance       // name -> Instance
    Templates  map[string]*Template       // id -> Template
    Resources  map[string]*Resource       // type:value -> Resource
    Counters   map[string]int             // counter_name -> current
    Types      map[string]*ResourceType   // Resource type definitions
}
```

**Instance** (process.go):
```go
type Instance struct {
    Name       string                 // User-provided name
    Template   string                 // Template ID
    Command    string                 // Final interpolated command
    PID        int                    // Process ID
    Status     string                 // stopped|starting|running|stopping
    Resources  map[string]string      // resource_type -> value
    Started    int64                  // Unix timestamp
}
```

**Template** (process.go):
```go
type Template struct {
    ID         string                 // Unique template ID
    Label      string                 // Human-readable label
    Command    string                 // Template with ${var} and %counter
    Resources  []string               // Resource types this needs
    Vars       map[string]string      // Default variables
}
```

**Resource** (resource.go):
```go
type Resource struct {
    Type       string                 // tcpport|vncport|gpu|license|whatever
    Value      string                 // "3000" or "/path" or "0"
    Owner      string                 // Instance name
}
```

**ResourceType** (resource.go):
```go
type ResourceType struct {
    Name       string                 // Resource type name
    Check      string                 // Shell command to check availability
    Counter    bool                   // Is this auto-incrementing?
    Start      int                    // Counter start value
    End        int                    // Counter end value
}
```
---

## Implementation Phases

### Phase 1: State & Resource System
- Core resource allocation without processes
- ResourceType and Resource structs
- AllocateResource with counter support
- CheckResource with shell command execution
- State persistence to JSON

### Phase 2: Process Management
- StartProcess with resource allocation
- StopProcess with cleanup
- Template and Instance structs
- Variable interpolation (${var} and %counter)
- Process tracking (PID, status)

### Phase 3: CLI Interface
- Minimal command dispatcher
- Commands: start, stop, ps, template, resource-type
- Argument parsing (--key=value)
- Pretty output

### Phase 4: Web UI
- Single-page HTML dashboard
- API endpoints (instances, templates, resources, types)
- Start/stop buttons
- Resource type editor
- Auto-refresh

### Phase 5: Polish & Testing
- Error handling
- Unit tests
- Example templates
- Documentation
- Installation script

---

## Future Enhancements (Post-Core)

### Enhanced Observability
- Real-time CPU/memory metrics from /proc
- Historical metrics graphs
- Process output streaming
- Log filtering and search

### Collaboration Features
- Import/export templates
- Template marketplace/library
- Shared state across team

### Advanced Resource Management
- Process dependency management (start order)
- Health check configuration (custom commands)
- Auto-restart on failure
- Resource quotas

---

## Non-Goals (Out of Scope)

- **Container Orchestration**: Not replacing Docker Compose or Kubernetes
- **Production Deployment**: Tool is for local/development use
- **CI/CD Integration**: Not a build or deployment pipeline
- **Cloud Process Management**: Local machine focus
- **Hardcoded Resource Types**: Everything user-definable

---

## Dependencies & Prerequisites

### User Prerequisites
- **Operating System**: Linux or macOS (Windows with WSL)
- **Shell**: bash/sh (for check commands)
- **Browser**: Modern browser for web UI (Chrome, Firefox, Safari, Edge)
- **Optional Tools**: nc (netcat), nvidia-smi, lmutil, etc. depending on resource types used

### Technical Dependencies
- **Go 1.21+**: stdlib only, zero external dependencies
- **No Runtime Dependencies**: Single static binary

---

## Security Considerations

### Security Model
- **No Authentication**: Single-user local application
- **Localhost Only**: Web UI binds to 127.0.0.1 by default
- **Process Isolation**: Uses OS-level process isolation (standard Unix fork/exec)
- **Shell Command Execution**: Check commands run via `sh -c` (user responsibility to avoid injection)
- **State File Permissions**: ~/.vibeprocess/state.json is user-readable only (0600)

### Security Best Practices
- Review check commands before adding resource types
- Don't run as root (processes inherit permissions)
- Use explicit resource values when possible (avoid auto-allocation for sensitive resources)
- State file contains PIDs and commands - don't commit to git

---

## Appendix

### Example Templates

**PostgreSQL Database**:
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

**QEMU Virtual Machine**:
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

**Node.js Server**:
```json
{
  "id": "node-express",
  "label": "Node.js Express Server",
  "command": "node server.js --port ${tcpport}",
  "resources": ["tcpport"],
  "vars": {}
}
```

**ML Training (Custom GPU Resource)**:
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

### Glossary

- **Template**: Reusable process configuration with resource requirements
- **Instance**: Named, running process created from a template
- **Resource**: Generic type:value pair (port, file, GPU, license, etc.)
- **ResourceType**: User-defined resource with check command
- **Counter**: Auto-incrementing resource type (tcpport, vncport, etc.)
- **Check Command**: Shell command to validate resource availability
- **Interpolation**: ${var} and %counter replacement in commands
