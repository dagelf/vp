# Vibeprocess Manager - Implementation Plan

## Phase 1: Frontend with Mock Data

This document outlines the implementation plan for Phase 1 of Vibeprocess Manager, focusing on building a functional frontend interface with mock data.

---

## 1. Technology Stack Decisions

### Core Framework
- **Next.js 15+ (App Router)**: Modern React framework with built-in routing
- **React 19+**: Latest React features with improved performance
- **TypeScript**: Type safety for data models and components

### Styling
- **Tailwind CSS**: Utility-first CSS for rapid UI development
- **shadcn/ui**: Pre-built, accessible component library
- **Lucide React**: Icon library for consistent iconography

### State Management
- **React Context + Hooks**: Lightweight state management for Phase 1
- **localStorage**: Client-side persistence for templates and instances

### Development Tools
- **ESLint**: Code quality and consistency
- **Prettier**: Code formatting
- **TypeScript**: Type checking

### Why This Stack?
1. **Minimal Dependencies**: Core request from PRD
2. **Simple Architecture**: Easy to understand and maintain
3. **Fast Development**: Pre-built components accelerate UI creation
4. **Type Safety**: TypeScript prevents runtime errors
5. **Future-Ready**: Easy migration to backend in Phase 2

---

## 2. Project Structure

```
vp/
├── app/                          # Next.js app directory
│   ├── layout.tsx               # Root layout with theme provider
│   ├── page.tsx                 # Main dashboard page
│   └── globals.css              # Global styles and Tailwind imports
├── components/
│   ├── ui/                      # shadcn/ui base components
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── dialog.tsx
│   │   ├── tabs.tsx
│   │   ├── badge.tsx
│   │   └── ...
│   ├── dashboard/               # Main dashboard components
│   │   ├── Dashboard.tsx        # Main container component
│   │   ├── InstancesTab.tsx     # Process instances view
│   │   ├── TemplatesTab.tsx     # Templates management
│   │   ├── ResourcesTab.tsx     # Resource allocation view
│   │   └── LogsTab.tsx          # Logs and history view
│   ├── instances/               # Instance-related components
│   │   ├── InstanceCard.tsx     # Single instance display
│   │   ├── InstanceForm.tsx     # Create/edit instance form
│   │   ├── InstanceMetrics.tsx  # CPU/memory metrics display
│   │   └── StatusBadge.tsx      # Status indicator
│   ├── templates/               # Template-related components
│   │   ├── TemplateCard.tsx     # Single template display
│   │   ├── TemplateForm.tsx     # Create/edit template form
│   │   └── TemplatePreview.tsx  # Template details preview
│   ├── resources/               # Resource-related components
│   │   ├── PortAllocationTable.tsx
│   │   ├── ResourceOverview.tsx
│   │   └── ConflictIndicator.tsx
│   └── modals/                  # Modal dialogs
│       ├── CreateInstanceModal.tsx
│       ├── CreateTemplateModal.tsx
│       └── ConnectionModal.tsx
├── lib/
│   ├── types.ts                 # TypeScript interfaces
│   ├── mock-data.ts             # Mock templates and instances
│   ├── utils.ts                 # Utility functions
│   ├── resource-manager.ts      # Resource conflict detection
│   └── variable-interpolation.ts # Template variable handling
├── contexts/
│   ├── ProcessContext.tsx       # Process state management
│   └── ThemeProvider.tsx        # Dark mode support
├── hooks/
│   ├── useProcesses.ts          # Process management hooks
│   ├── useTemplates.ts          # Template management hooks
│   ├── useResources.ts          # Resource tracking hooks
│   └── useLocalStorage.ts       # Persistent storage hooks
├── public/
│   └── favicon.ico
├── PRD.md                        # Product requirements (this doc)
├── PLAN.md                       # Implementation plan
├── package.json
├── tsconfig.json
├── tailwind.config.ts
├── next.config.js
└── README.md
```

---

## 3. Implementation Phases

### Phase 1.1: Project Setup & Foundation (Day 1)

**Tasks:**
1. Initialize Next.js project with TypeScript
2. Install and configure Tailwind CSS
3. Set up shadcn/ui component library
4. Create project structure (directories)
5. Define TypeScript interfaces in `lib/types.ts`
6. Create mock data in `lib/mock-data.ts`
7. Set up ESLint and Prettier

**Deliverables:**
- Working Next.js application scaffold
- Type-safe data models
- Mock data for 3-5 templates and instances
- Development environment configured

### Phase 1.2: Core Data Layer (Day 1-2)

**Tasks:**
1. Implement ProcessContext for state management
2. Create useProcesses hook with CRUD operations
3. Create useTemplates hook with CRUD operations
4. Implement localStorage persistence
5. Build variable interpolation engine
6. Create resource conflict detection logic
7. Implement auto-increment port allocation

**Deliverables:**
- Functional state management
- Working variable substitution (${port}, %tcpport)
- Resource conflict detection
- Data persistence across page refreshes

### Phase 1.3: Dashboard Layout & Navigation (Day 2)

**Tasks:**
1. Create main Dashboard component
2. Implement tabbed navigation (Instances, Templates, Resources, Logs)
3. Build responsive layout with proper spacing
4. Add theme provider for dark mode support
5. Create header with app branding
6. Add status indicators and global metrics

**Deliverables:**
- Complete dashboard layout
- Tabbed navigation working
- Responsive design
- Dark mode toggle

### Phase 1.4: Process Instances UI (Day 3)

**Tasks:**
1. Build InstanceCard component with status badges
2. Create InstanceForm for creating/editing
3. Implement start/stop button functionality
4. Add process metrics display (CPU, memory, uptime)
5. Build CreateInstanceModal
6. Add instance filtering and sorting
7. Implement connection button with ConnectionModal

**Deliverables:**
- Full instances management UI
- Start/stop simulation with status transitions
- Metrics visualization
- Connection commands execution

### Phase 1.5: Process Templates UI (Day 3-4)

**Tasks:**
1. Build TemplateCard component
2. Create TemplateForm for template creation
3. Add variable editor with syntax highlighting
4. Implement template preview functionality
5. Build CreateTemplateModal
6. Add template deletion with confirmation
7. Create default template library

**Deliverables:**
- Template management interface
- Template creation/editing forms
- Visual template preview
- Pre-configured example templates

### Phase 1.6: Resource Management UI (Day 4)

**Tasks:**
1. Build ResourcesTab component
2. Create PortAllocationTable showing allocations
3. Add resource filter (all, in use, available)
4. Implement ConflictIndicator component
5. Show resource allocation history
6. Display PID assignments
7. Add resource usage statistics

**Deliverables:**
- Resource overview dashboard
- Port allocation visualization
- Conflict detection display
- Resource filtering

### Phase 1.7: Logs & History (Day 5)

**Tasks:**
1. Create LogsTab component
2. Implement event logging system
3. Add log filtering (by instance, by type, by date)
4. Create log entry components
5. Add start/stop event logging
6. Implement error message display
7. Add log export functionality

**Deliverables:**
- Logs view with filtering
- Event history tracking
- Error message display

### Phase 1.8: Polish & UX Refinements (Day 5-6)

**Tasks:**
1. Add loading states and animations
2. Implement error handling and user feedback
3. Add confirmation dialogs for destructive actions
4. Improve keyboard navigation and accessibility
5. Add tooltips for complex features
6. Optimize performance (memoization, lazy loading)
7. Add empty states with helpful CTAs
8. Create onboarding hints for first-time users

**Deliverables:**
- Polished user experience
- Comprehensive error handling
- Accessibility improvements
- Performance optimizations

### Phase 1.9: Testing & Documentation (Day 6-7)

**Tasks:**
1. Manual testing of all features
2. Cross-browser testing
3. Responsive design testing (mobile, tablet, desktop)
4. Create README.md with setup instructions
5. Document component API and usage
6. Add inline code comments
7. Create user guide screenshots
8. Test data persistence and recovery

**Deliverables:**
- Fully tested application
- Comprehensive README
- User documentation
- Bug-free release

---

## 4. Data Models (TypeScript Interfaces)

```typescript
// lib/types.ts

export type ProcessStatus =
  | "stopped"
  | "starting"
  | "running"
  | "stopping"
  | "error";

export interface Template {
  id: string;
  label: string;
  command_template: string;
  defaults?: Record<string, string>;
  variables?: string[];
  resources?: {
    ports?: string[];
    files?: string[];
  };
  exposes?: Record<string, string>;
  connections?: Record<string, string>;
  notes?: string;
  ai_instructions?: string;
}

export interface ProcessInstance {
  id: string;
  name: string;
  status: ProcessStatus;
  pid: number | null;
  ports: number[];
  template_id: string;
  notes?: string;
  vars: Record<string, string>;
  command: string;
  error_message?: string;
  cpu_usage?: number;
  memory_usage?: number;
  uptime?: number;
  created_at: string;
  started_at?: string;
  stopped_at?: string;
}

export interface ResourceAllocation {
  type: "port" | "file";
  value: string | number;
  instance_id: string;
  instance_name: string;
  allocated_at: string;
}

export interface LogEntry {
  id: string;
  timestamp: string;
  type: "info" | "warn" | "error" | "success";
  instance_id?: string;
  instance_name?: string;
  message: string;
}

export interface PortCounter {
  name: string; // e.g., "tcpport", "vnc"
  current: number;
  min: number;
  max: number;
}
```

---

## 5. Mock Data Examples

```typescript
// lib/mock-data.ts

export const mockTemplates: Template[] = [
  {
    id: "node-express",
    label: "Node.js Express Server",
    command_template: "node server.js --port ${port} --env ${env}",
    defaults: { port: "%tcpport", env: "development" },
    variables: ["port", "env"],
    resources: {
      ports: ["${port}"],
      files: ["server.js", "package.json"]
    },
    exposes: { http: ":${port}" },
    connections: {
      curl: "curl -I http://localhost:${port}",
      browser: "open http://localhost:${port}"
    },
    notes: "Standard Node.js Express server configuration"
  },
  {
    id: "postgresql",
    label: "PostgreSQL Database",
    command_template: "postgres -D ${data_dir} -p ${port}",
    defaults: { data_dir: "/var/lib/postgresql/data", port: "5432" },
    variables: ["data_dir", "port"],
    resources: {
      ports: ["${port}"],
      files: ["${data_dir}"]
    },
    exposes: { psql: ":${port}" },
    connections: {
      psql: "psql -h localhost -p ${port} -U postgres"
    }
  },
  {
    id: "redis",
    label: "Redis Cache",
    command_template: "redis-server --port ${port} --dir ${data_dir}",
    defaults: { port: "6379", data_dir: "/var/lib/redis" },
    variables: ["port", "data_dir"],
    resources: {
      ports: ["${port}"],
      files: ["${data_dir}"]
    },
    connections: {
      cli: "redis-cli -p ${port}"
    }
  }
];

export const mockInstances: ProcessInstance[] = [
  {
    id: "inst-1",
    name: "Dev API Server",
    status: "running",
    pid: 12345,
    ports: [3000],
    template_id: "node-express",
    vars: { port: "3000", env: "development" },
    command: "node server.js --port 3000 --env development",
    cpu_usage: 12.5,
    memory_usage: 256,
    uptime: 3600,
    created_at: "2025-11-13T10:00:00Z",
    started_at: "2025-11-13T10:05:00Z"
  },
  {
    id: "inst-2",
    name: "Test Database",
    status: "stopped",
    pid: null,
    ports: [],
    template_id: "postgresql",
    vars: { port: "5433", data_dir: "/var/lib/postgresql/test" },
    command: "postgres -D /var/lib/postgresql/test -p 5433",
    created_at: "2025-11-13T09:00:00Z",
    stopped_at: "2025-11-13T11:30:00Z"
  }
];
```

---

## 6. Key Features Implementation Details

### 6.1 Variable Interpolation

```typescript
// lib/variable-interpolation.ts

export function interpolateVariables(
  template: string,
  vars: Record<string, string>,
  counters: Record<string, PortCounter>
): string {
  let result = template;

  // Replace ${variable} with values
  Object.entries(vars).forEach(([key, value]) => {
    result = result.replace(new RegExp(`\\$\\{${key}\\}`, 'g'), value);
  });

  // Replace %counter with auto-incremented values
  Object.entries(counters).forEach(([key, counter]) => {
    if (result.includes(`%${key}`)) {
      const nextValue = getNextAvailableValue(counter);
      result = result.replace(`%${key}`, String(nextValue));
    }
  });

  return result;
}

function getNextAvailableValue(counter: PortCounter): number {
  // Implementation for finding next available port
  // Checks system availability and increments
  return counter.current++;
}
```

### 6.2 Resource Conflict Detection

```typescript
// lib/resource-manager.ts

export function detectConflicts(
  instance: ProcessInstance,
  allInstances: ProcessInstance[]
): string[] {
  const conflicts: string[] = [];

  // Check port conflicts
  instance.ports.forEach(port => {
    const conflicting = allInstances.find(
      inst => inst.id !== instance.id &&
              inst.status === 'running' &&
              inst.ports.includes(port)
    );
    if (conflicting) {
      conflicts.push(`Port ${port} already in use by ${conflicting.name}`);
    }
  });

  // Check file conflicts (for Phase 1, simplified)
  // Future: Check actual file system locks

  return conflicts;
}
```

### 6.3 Process Status Simulation

```typescript
// Mock process lifecycle for Phase 1
export function simulateProcessStart(instanceId: string): Promise<void> {
  return new Promise((resolve, reject) => {
    // Simulate starting delay
    setTimeout(() => {
      // 90% success rate in mock
      if (Math.random() > 0.1) {
        resolve();
      } else {
        reject(new Error('Failed to start process'));
      }
    }, 2000);
  });
}

export function simulateProcessStop(instanceId: string): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, 1000);
  });
}

// Mock metrics generation
export function generateMockMetrics() {
  return {
    cpu_usage: Math.random() * 100,
    memory_usage: Math.random() * 1024
  };
}
```

---

## 7. UI Component Specifications

### 7.1 InstanceCard Component

**Props:**
- `instance: ProcessInstance`
- `onStart: () => void`
- `onStop: () => void`
- `onEdit: () => void`
- `onConnect: () => void`

**Features:**
- Color-coded status badge (green=running, gray=stopped, yellow=starting, red=error)
- CPU/memory progress bars
- Uptime display
- Port allocations list
- Quick action buttons
- Expandable details section

### 7.2 TemplateCard Component

**Props:**
- `template: Template`
- `onUse: () => void`
- `onEdit: () => void`
- `onDelete: () => void`

**Features:**
- Template name and description
- Variable list display
- Command preview
- Usage count indicator
- Quick "Use Template" button

### 7.3 Dashboard Tabs Structure

**Instances Tab:**
- Grid/list view toggle
- Status filter (all, running, stopped, error)
- Sort options (name, status, CPU, memory)
- Search bar
- "Create Instance" button

**Templates Tab:**
- Grid view of template cards
- Search/filter by type
- "Create Template" button
- "Import Template" button (future)

**Resources Tab:**
- Port allocation table
- File resource list
- Conflict warnings
- Resource usage charts

**Logs Tab:**
- Chronological log entries
- Filter by instance
- Filter by log level
- Search functionality
- Clear logs button

---

## 8. Development Workflow

### Setup Commands
```bash
# Initialize project
npx create-next-app@latest vp --typescript --tailwind --app

# Install shadcn/ui
npx shadcn-ui@latest init

# Install components
npx shadcn-ui@latest add button card dialog tabs badge input label select

# Install additional dependencies
npm install lucide-react date-fns
```

### Development Commands
```bash
# Start dev server
npm run dev

# Type check
npm run type-check

# Lint
npm run lint

# Build
npm run build

# Start production server
npm start
```

---

## 9. Success Criteria for Phase 1

### Functional Requirements
- [ ] Create and manage process templates
- [ ] Create instances from templates
- [ ] Start/stop instances (simulated)
- [ ] View process metrics (mock data)
- [ ] Detect port conflicts before starting
- [ ] Variable interpolation working
- [ ] Connection commands work
- [ ] Data persists in localStorage
- [ ] All tabs functional

### Non-Functional Requirements
- [ ] Responsive design (mobile, tablet, desktop)
- [ ] Dark mode support
- [ ] < 2s page load time
- [ ] Accessible (keyboard navigation, ARIA labels)
- [ ] No console errors
- [ ] Type-safe (no TypeScript errors)

### User Experience
- [ ] Intuitive navigation
- [ ] Clear error messages
- [ ] Loading states for async operations
- [ ] Empty states with helpful guidance
- [ ] Confirmation for destructive actions

---

## 10. Phase 2 Preparation

While implementing Phase 1, keep in mind these architectural decisions for Phase 2:

1. **API Design**: Structure components to easily swap mock data with API calls
2. **State Management**: Use patterns that can scale to server state (React Query)
3. **Process Communication**: Design connection system to work with real processes
4. **Security**: Add input validation that will carry forward
5. **Error Handling**: Build robust error handling from the start

---

## 11. Timeline Estimate

**Total Estimated Time: 5-7 days**

- Day 1: Project setup, data layer, mock data
- Day 2: Dashboard layout, navigation, state management
- Day 3: Instances UI and functionality
- Day 4: Templates UI and resource management
- Day 5: Logs, polish, UX refinements
- Day 6-7: Testing, documentation, bug fixes

---

## 12. Next Steps

1. **Review and approve this plan**
2. **Set up development environment**
3. **Begin Phase 1.1: Project Setup**
4. **Daily progress check-ins**
5. **Demo after Phase 1.4 (instances working)**
6. **Final review before Phase 2 planning**

---

## Appendix: Technology Research

### Why Not Other Frameworks?

**Vue.js/Nuxt**: Smaller ecosystem, team familiarity with React
**Svelte/SvelteKit**: Less mature for production, smaller community
**Plain React**: No built-in routing, more setup overhead
**Angular**: Too heavy for this use case, steeper learning curve

### Why Tailwind over CSS Modules?

- Faster development
- Consistent design system
- No naming conflicts
- Easy responsive design
- Works well with shadcn/ui

### Why shadcn/ui over Material-UI/Chakra?

- Lighter weight (copy components, not full library)
- Full customization
- Better TypeScript support
- Accessible by default
- Modern design aesthetic
