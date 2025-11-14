# Vibeprocess Manager

**Firmware-style process orchestration with zero assumptions**

Vibeprocess Manager is an ultra-lean process manager built on a radical philosophy: make **zero assumptions** about what resources are or how they work. Everything is user-defined through simple shell commands.

## Philosophy

- **6 files, ~500 lines of Go** - Brutally simple
- **Zero dependencies** - stdlib only
- **Generic resources** - Not opinionated about ports, files, GPUs, etc.
- **Validation via shell commands** - Use any tool (nc, test, nvidia-smi, lmutil)
- **Pure mechanism, no policy** - Like firmware that provides primitives

## Quick Start

```bash
# Build
go build -o vp

# Start web UI
./vp serve

# Or use CLI
./vp ps
./vp start postgres mydb
./vp stop mydb
```

## Architecture

```
vp/
├── main.go          # CLI entry point (~80 lines)
├── state.go         # State persistence (~100 lines)
├── process.go       # Process lifecycle (~150 lines)
├── resource.go      # Generic resource system (~100 lines)
├── api.go           # HTTP server (~70 lines)
└── web.html         # Embedded UI (single page)
```

## Resource System

Resources are just **type:value pairs** validated by **shell commands**:

```bash
# Built-in resources (defaults)
tcpport   -> nc -z localhost ${value}
vncport   -> nc -z localhost ${value}
dbfile    -> test -f ${value}
socket    -> test -S ${value}

# Add custom resources
vp resource-type add gpu --check='nvidia-smi -L | grep GPU-${value}'
vp resource-type add license --check='lmutil lmstat -c ${value} | grep "UP"'
```

## Templates

Define how to start processes with resource requirements:

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

## Usage

```bash
# Start with auto-allocated resources
vp start postgres mydb

# Start with explicit resource values
vp start postgres mydb --tcpport=5432 --datadir=/var/db

# Mix explicit and auto
vp start qemu vm1 --vncport=5901  # serialport auto-allocated

# List instances
vp ps

# Stop instance
vp stop mydb

# Manage templates
vp template list
vp template add template.json

# Manage resource types
vp resource-type list
vp resource-type add gpu --check='nvidia-smi -L | grep GPU-${value}'
```

## Web UI

```bash
vp serve
# Open http://localhost:8080
```

Features:
- View all instances
- Start/stop with buttons
- Add templates via form
- Add resource types via form
- View resource allocations
- Auto-refresh every 5 seconds

## Why This Design is Genius

1. **Zero Hardcoded Assumptions** - Resources aren't hardcoded
2. **Maximum Flexibility** - Add ANY resource type at runtime
3. **Validation via Shell** - Use any installed tool
4. **Counter Resources Not Special-Cased** - Just a boolean flag
5. **Brutally Simple** - 6 files, ~500 lines
6. **Firmware-Style** - Pure primitives, users configure behavior
7. **Debuggable** - Human-readable JSON state
8. **Extensible Without Code Changes** - Add types via CLI

## State Storage

Everything persists to `~/.vibeprocess/state.json`:

```json
{
  "instances": {...},
  "templates": {...},
  "resources": {...},
  "counters": {...},
  "types": {...}
}
```

## Examples

### Custom GPU Resource
```bash
vp resource-type add gpu \
  --check='nvidia-smi -L | grep GPU-${value}' \
  --counter=false

# Now use in templates
vp start ml-training job1 --gpu=0
```

### License Server
```bash
vp resource-type add flexlm \
  --check='lmutil lmstat -c ${value} | grep "UP"'

vp start matlab session1 --flexlm=27000@licserver
```

### Database Connection
```bash
vp resource-type add dbconn \
  --check='psql -h ${value} -c "SELECT 1"'

vp start webapp api --dbconn=localhost:5432/mydb
```

## Design for Mars

**This is how you design for Mars - assume nothing, enable everything.**

Want GPU allocation? Add a resource type.
Want license servers? Add a resource type.
Want database connections? Add a resource type.
Want anything? Just define a check command.

---

See [PLAN.md](PLAN.md) for implementation details.
See [PRD.md](PRD.md) for product requirements.
