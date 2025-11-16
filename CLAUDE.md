# Product Requirements Document (PRD)

## Vibeprocess Manager

Process orchestration tool that provides pure primitives for process and resource management. It makes **zero assumptions** about what resources are or how they work - everything is user-defined through simple check commands. Think the Unix `ps` tool, modernized, with generic resource allocation and zero opinions.

### Current Pain Points

Modern servers often experience process creep: some processes are started by systemd, some in screen, and many manually. When the server experiences a failure or reboot, a lot of state goes missing - and lore about how and what to start and what was running drifts.

1. **Manual Process Management**: Operators must manually start/stop processes in multiple terminal windows, making it difficult to track what's running and where
2. **Resource Conflicts**: Port collisions and file conflicts are common when running multiple instances or switching between projects in various states of development and production and update.
3. **Configuration Complexity**: Each process requires specific environment variables, ports, and file paths that must be remembered and typed manually
4. **Lack of Visibility**: No centralized view of running processes, resource usage, or allocated ports
5. **Onboarding Friction**: New team members struggle to understand which processes to run and how to configure them
6. **Connection Management**: Accessing services requires remembering connection strings and CLI commands

### Glossary

- **Template**: Reusable process configuration with resource requirements
- **Instance**: Named, running process created from a template
- **Resource**: Generic type:value pair (port, file, GPU, license, etc.)
- **ResourceType**: User-defined resource with check command
- **Counter**: Auto-incrementing resource type (tcpport, vncport, etc.)
- **Check Command**: Shell command to validate resource availability
- **Interpolation**: ${var} and %counter replacement in commands
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

**Target**: minimal number of files, minimal LoC while maintaining readability, shrewd and visionary design with great planning

### Phase 5: Polish & Testing
- Error handling
- Better configuration editing interface
- Edge cases
- API security
- Example templates / template library or marketplace

