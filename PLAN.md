# Vibeprocess Manager - Implementation Plan

## Architecture Philosophy

**Firmware-Style Design**: Pure mechanism, no policy. Assume nothing, enable everything.

### Core Principles
1. **Zero Hardcoded Assumptions** - No opinions about what resources are
2. **Everything is a Command** - Validation via shell commands
3. **Six Files Total** - Brutally simple structure
4. **Generic Resources** - Type:value pairs with check commands
5. **Single State File** - Human-readable JSON
6. **Pure Primitives** - Memory allocation, process control, state persistence

### Why This Matters
- Maximum flexibility without code changes
- Add any resource type at runtime (GPU, license, DB, whatever)
- Validate using any installed tool (nc, test, nvidia-smi, lmutil)
- Debuggable: all state in one JSON file
- Extensible: shell commands = infinite possibilities

---

## Project Structure

```
vp/
├── main.go          # CLI dispatcher
├── state.go         # State persistence
├── process.go       # Process lifecycle
├── resource.go      # Generic resource system
├── api.go           # HTTP server
└── web.html         # Embedded UI
```

**Target**: 6 files, ~500 lines total

---

## Core Concepts

### Resources
- Generic type:value pairs (e.g., "tcpport:3000", "gpu:0")
- Validated by shell commands
- Counter resources auto-increment
- No special cases

### Templates
- Define how to start processes
- Declare resource requirements
- Support variable interpolation (${var} and %counter)

### Instances
- Running processes created from templates
- Track PID, status, allocated resources
- Resource cleanup on stop

### State
- Everything persists to ~/.vibeprocess/state.json
- Includes: instances, templates, resources, counters, types
- Load on start, save on changes

---

## Implementation Phases

### Phase 1: State & Resource System
**Goal**: Core allocation without processes

**Focus**:
- State persistence (load/save JSON)
- Resource types with check commands
- AllocateResource with counter support
- CheckResource via shell execution

**Milestone**: Can allocate/deallocate resources, counters work, check commands execute

### Phase 2: Process Management
**Goal**: Start/stop processes with resource allocation

**Focus**:
- Template and Instance structures
- StartProcess: allocate resources → interpolate → exec
- StopProcess: kill process → release resources
- Variable interpolation (${var} and %counter)

**Milestone**: Can start/stop processes from templates, resources allocated/released correctly

### Phase 3: CLI Interface
**Goal**: Minimal command-line tool

**Focus**:
- Command dispatcher (start, stop, ps, serve, template, resource-type)
- Argument parsing (--key=value)
- Human-readable output

**Milestone**: All CLI commands work, argument parsing correct, output clear

### Phase 4: Web UI
**Goal**: Single-page interface

**Focus**:
- HTTP server with API endpoints
- Single-page HTML with inline CSS/JS
- Embed web.html using go:embed
- Auto-refresh instances

**Milestone**: Web UI accessible, can start/stop from browser, view resources

### Phase 5: Polish & Testing
**Goal**: Production-ready

**Focus**:
- Error handling
- Edge cases
- Example templates
- Documentation
- Build/install process

**Milestone**: Ready for daily use, all tests pass, docs complete

---

## Default Resource Types

Provided as examples, not limitations:
- **tcpport**: Auto-increment TCP ports (3000-9999)
- **vncport**: Auto-increment VNC ports (5900-5999)
- **serialport**: Auto-increment serial ports (9600-9699)
- **dbfile**: Database files (check if exists)
- **socket**: Unix sockets (check if exists)
- **datadir**: Data directories (no check)

Users can add any resource type at runtime.

---

## Default Templates

Provided as examples:
- **postgres**: PostgreSQL with tcpport + datadir
- **node-express**: Node.js server with tcpport
- **qemu**: QEMU VM with vncport + serialport

Users can add templates via CLI or API.

---

## Success Criteria

### Phase 1
- [ ] Resources allocate/deallocate correctly
- [ ] Counter resources auto-increment
- [ ] Check commands execute and validate
- [ ] State persists to JSON

### Phase 2
- [ ] Processes start from templates
- [ ] Resources allocated before start
- [ ] Variables interpolated (${var} and %counter)
- [ ] PIDs tracked
- [ ] Resources released on stop

### Phase 3
- [ ] All CLI commands work
- [ ] Arguments parsed correctly
- [ ] Error messages clear
- [ ] Output human-readable

### Phase 4
- [ ] Web UI serves on localhost:8080
- [ ] Can start/stop from browser
- [ ] Resource allocations visible
- [ ] Auto-refresh works

### Phase 5
- [ ] Error handling comprehensive
- [ ] Edge cases handled
- [ ] Example templates included
- [ ] README clear
- [ ] Builds cleanly

---

## Design Decisions

### Why Go?
- Single binary deployment
- Excellent stdlib (http, json, exec)
- No runtime dependencies
- Cross-platform

### Why Single JSON State File?
- Human-readable
- Easy to debug
- Can manually edit if needed
- Simple backup/restore

### Why Shell Commands for Validation?
- Maximum flexibility
- Use any installed tool
- No custom validation logic
- Easy to test independently

### Why No Hardcoded Resource Types?
- Future-proof
- User-extensible
- Works for any use case
- No assumptions about environment

---

## The Point

This design makes **zero assumptions** about:
- What resources are (just strings)
- How to validate them (just commands)
- What templates do (just start processes)
- What users need (they define everything)

It provides pure primitives:
- **Memory allocation** (resources)
- **Process control** (start/stop)
- **State persistence** (JSON)
- **Variable interpolation** (string replacement)
- **Validation** (shell commands)

Everything else is user configuration.

**Design for Mars: assume nothing, enable everything.**
