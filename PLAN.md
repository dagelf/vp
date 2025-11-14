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
- Manage processess no matter how they were started
- Validate using any installed tool (nc, test, nvidia-smi, lmutil)
- Debuggable: all state in one JSON file
- Extensible: shell commands = infinite possibilities

---

**Target**: minimal number of files, minimal LoC while maintaining readability, shrewd and visionary design with great planning

---

## Core Concepts

### Resources
- Generic type:value pairs (e.g., "tcpport:3000", "gpu:0")
- Each type has a validation shell command
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

### Phase 2: Process Management ✅ COMPLETE
**Goal**: Start/stop processes with resource allocation

**Focus**:
- Template and Instance structures
- StartProcess: allocate resources → interpolate → exec
- StopProcess: kill process → release resources
- RestartProcess: restart stopped instances with same resources
- Variable interpolation (${var} and %counter)
- Proper zombie reaping
- Process group management

**Milestone**: Can start/stop/restart processes, resources allocated/released correctly, no zombies

**Critical Fix**: Template-independent restart - instances now contain all data needed to restart (command + resources), no longer depend on template existing

### Phase 3: CLI Interface
**Goal**: Minimal command-line tool

**Focus**:
- Command dispatcher (start, stop, ps, serve, template, resource-type)
- Argument parsing (--key=value)
- Human-readable output

**Milestone**: All CLI commands work, argument parsing correct, output clear

### Phase 4: Web UI ✅ COMPLETE
**Goal**: Single-page interface

**Focus**:
- HTTP server with API endpoints
- Single-page HTML with inline CSS/JS
- Embed web.html using go:embed
- Auto-refresh instances
- Configuration editor

**Milestone**: Web UI accessible, can start/stop/restart from browser, view resources, edit config

**Features Implemented**:
- Tabbed interface (Instances, Templates, Resources, Types, Configuration)
- Configurable auto-refresh (1s default, adjustable)
- Clean refresh without visual flashing
- Restart stopped instances (no template required)
- Copy template command to clipboard
- Direct JSON configuration editing
- Real-time status updates

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

---

## Current Status (2025-11-15)

### What Works
- ✅ All 6 files implemented (~500 lines each)
- ✅ Full CLI with start/stop/restart/ps/serve/template/resource-type
- ✅ Web UI with 5 tabs (Instances, Templates, Resources, Types, Config)
- ✅ Generic resource system with shell command validation
- ✅ Proper process lifecycle (no zombies, graceful shutdown)
- ✅ Template-independent restart (uses stored command + resources)
- ✅ Configurable auto-refresh without flashing
- ✅ JSON configuration editor in web UI
- ✅ Default templates (postgres, node-express, qemu)
- ✅ Default resource types (tcpport, vncport, serialport, dbfile, socket, datadir)

### Recent Enhancements
1. **Restart Functionality** (2025-11-15)
   - Fixed design flaw: restart no longer requires original template
   - Instance struct contains all data needed (command + resources)
   - Added `RestartProcess()` in process.go:199-260
   - Added CLI command and API endpoint
   - Web UI restart button for stopped instances

2. **Configuration Editor** (2025-11-15)
   - New Configuration tab in web UI
   - Direct JSON editing of entire state
   - Real-time validation
   - Save/reload functionality
   - API endpoint: GET/POST /api/config

3. **Web UI Polish** (2025-11-15)
   - Configurable auto-refresh (1s default, 0-60s range)
   - Clean refresh without DOM flashing
   - Copy template command button
   - Status-based action buttons (Stop for running, Restart for stopped)

### File Breakdown
- **main.go** (267 lines) - CLI dispatcher, all commands
- **state.go** (~150 lines) - State persistence, default templates/types
- **process.go** (272 lines) - Process lifecycle, zombie reaping, restart
- **resource.go** (127 lines) - Generic resource allocation, validation
- **api.go** (250 lines) - HTTP server, all API endpoints including config
- **web.html** (600 lines) - Complete web UI with all features

**Total**: ~1,666 lines (within target, very lean)

### API Endpoints
- GET/POST `/api/instances` - List/control instances
- GET/POST `/api/templates` - Manage templates
- GET `/api/resources` - View resource allocations
- GET/POST `/api/resource-types` - Manage resource types
- GET/POST `/api/config` - View/edit entire state

### Next Steps
- [ ] Iterate on functionality and co-create with the user
- [ ] Remove clutter from .md files
- [ ] Example use cases
- [ ] Installation instructions
- [ ] Performance testing with many instances
