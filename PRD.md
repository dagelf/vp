# Product Requirements Document (PRD)

## Vibeprocess Manager

Vibeprocess Manager is a web-based process orchestration tool that simplifies the management of multiple interconnected processes. It provides an intuitive interface for creating reusable process templates, instantiating configured processes, monitoring resource allocation, and preventing port/resource conflicts. Think the Unix `ps` tool, modernized, with memory and intelligence.

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

### 3. Resource Management

**Description**: Intelligent tracking and allocation of system resources to prevent conflicts.

**Key Capabilities**:
- Track port allocations across all instances
- Monitor file/directory usage
- Display PID assignments
- Auto-detect resource conflicts before starting processes
- Auto-increment port counters to find available ports
- Filter resources by status (all, in use, available)
- Show resource allocation history

**User Value**: Eliminate "address already in use" errors and port collision headaches.

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

TBD. Minimal Dependencies. Simple architecture.

### Data Models

NB Treat as a guideline

**Template Interface**:
```typescript
{
  id: string
  label: string
  command_template: string
  defaults?: Record<string, string>
  variables?: Record<string, string>
  resources?: Record<string, string>
  exposes?: Record<string, string>
  connections?: Record<string, string>
}
```

**ProcessInstance Interface**:
```typescript
{
  id: string
  name: string
  status: "stopped" | "starting" | "running" | "stopping" | "error"
  pid: number | null
  ports: number[]
  template_id: string
  notes?: string
  vars: Record<string, string>
  command: string
  error_message?: string
  cpu_usage?: number
  memory_usage?: number
}
```
---

## Future Roadmap

### Phase 1: Frontend with mock data

### Phase 2: Backend Integration
- Actual process spawning
- WebSocket for real-time process output streaming?
- File system integration for template persistence
- Process lifecycle management (spawn, kill, restart)

### Phase 3: Enhanced Observability
- Real-time log streaming and filtering
- Historical metrics and graphs
- Process dependency visualization
- Alert configuration (CPU/memory thresholds)

### Phase 4: Collaboration & Sharing
- Import/export templates as JSON/YAML
- Team template libraries
- Workspace persistence (save/load configurations)
- Configuration versioning

### Phase 5: Advanced Features
- Process dependency management (start order)
- Health check configuration
- Auto-restart on failure
- Environment variable management
- Docker container integration
- SSH remote process management

### Phase 6: Developer Experience
- CLI companion tool
- VS Code extension
- Global hotkeys for common actions
- Process grouping and bulk operations
- Quick start from project detection

---

## Non-Goals (Out of Scope)

- **Container Orchestration**: Not replacing Docker Compose or Kubernetes
- **Production Deployment**: Tool is for local development only
- **CI/CD Integration**: Not a build or deployment pipeline
- **Cloud Process Management**: Local machine only
- **Programming Language Execution**: Not a code editor or runtime manager

---

## Dependencies & Prerequisites

### User Prerequisites
- Modern web browser (Chrome, Firefox, Safari, Edge)
- Node.js installed (for process execution in Phase 2)
- Operating system: Linux, macOS, or Windows

### Technical Dependencies
- Next.js 16+
- React 19+
- Node.js 20+ (for development)
- Modern CSS support (Grid, Flexbox)

---

## Security Considerations

### Current Security Model
- **No Authentication**: Single-user local application
- **No Network Exposure**: Runs on localhost only
- **Process Isolation**: Will use OS-level process isolation (Phase 2)
- **Input Validation**: Template commands executed as shell commands (requires sanitization)

### Future Security Enhancements (Phase 2+)
- Command injection prevention
- Resource limit enforcement (CPU, memory, file descriptors)
- Template signing/verification for shared libraries
- Sandboxed process execution options
- Audit logging for security events

---

## Open Questions

1. **Process Execution Model**: Should we use child_process, PM2, or another process manager?
2. **Persistence Strategy**: SQLite, JSON files, or localStorage?
3. **Multi-Project Support**: Should users be able to switch between different project workspaces?
4. **Template Marketplace**: Should there be a community-shared template repository?
5. **Cross-Platform Support**: How do we handle OS-specific commands (bash vs PowerShell)?

---

## Appendix

### Example Templates

**Node.js Express Server**:
```javascript
{
  id: "node-express",
  label: "Node.js Express Server",
  command_template: "node server.js --port ${port} --env ${env}",
  defaults: { port: "%tcpport", env: "development" },
  variables: ["port", "env"],
  resources: { ports: ["${port}"], files: ["server.js", "package.json"] },
  exposes: { http: ":${port}" },
  connections: {
    curl: "curl -I http://localhost:${port}",
    browser: "open http://localhost:${port}"
  }
}
```

**PostgreSQL Database**:
```javascript
{
  id: "postgresql",
  label: "PostgreSQL Database",
  command_template: "postgres -D ${data_dir} -p ${port}",
  defaults: { data_dir: "/var/lib/postgresql/data", port: "5432" },
  variables: ["data_dir", "port"],
  resources: { ports: ["${port}"], files: ["${data_dir}"] },
  exposes: { psql: ":${port}" },
  connections: { psql: "psql -h localhost -p ${port} -U postgres" }
}
```

### Glossary

- **Template**: A reusable process configuration blueprint
- **Instance**: A named, configured instantiation of a template
- **Resource**: A system resource (port, file) that can be allocated to processes, can be auto incrementing or have a range
- **Connection**: A predefined command for accessing a running service
- **Interpolation**: The process of replacing variables with their values in templates
