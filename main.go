package main

import (
	"encoding/json"
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
		handleStart(args)
	case "stop":
		handleStop(args)
	case "ps":
		listInstances()
	case "serve":
		handleServe(args)
	case "template":
		handleTemplate(args)
	case "resource-type":
		handleResourceType(args)
	case "discover":
		handleDiscoverCLI(args)
	case "discover-port":
		handleDiscoverPortCLI(args)
	case "inspect":
		handleInspect(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Fprintf(os.Stderr, "Commands: start, stop, ps, serve, template, resource-type, discover, discover-port, inspect\n")
		os.Exit(1)
	}
}

func handleStart(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: vp start <template> <name> [--key=value...]\n")
		os.Exit(1)
	}

	templateID := args[0]
	name := args[1]
	vars := parseVars(args[2:])

	template := state.Templates[templateID]
	if template == nil {
		fmt.Fprintf(os.Stderr, "Template not found: %s\n", templateID)
		fmt.Fprintf(os.Stderr, "Available templates:\n")
		for id, tmpl := range state.Templates {
			fmt.Fprintf(os.Stderr, "  %s - %s\n", id, tmpl.Label)
		}
		os.Exit(1)
	}

	inst, err := StartProcess(state, template, name, vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Started %s (PID %d)\n", inst.Name, inst.PID)
	fmt.Printf("Command: %s\n", inst.Command)
	fmt.Printf("Resources:\n")
	for k, v := range inst.Resources {
		fmt.Printf("  %s = %s\n", k, v)
	}
}

func handleStop(args []string) {
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

	if err := StopProcess(state, inst); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	state.ReleaseResources(name)
	delete(state.Instances, name)
	state.Save()

	fmt.Printf("Stopped %s\n", name)
}

func handleServe(args []string) {
	port := "8080"
	if len(args) > 0 {
		port = args[0]
	}

	fmt.Printf("Starting web UI on http://localhost:%s\n", port)
	if err := ServeHTTP(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}

func handleTemplate(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: vp template <list|add|show>\n")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		for id, tmpl := range state.Templates {
			fmt.Printf("%-20s %s\n", id, tmpl.Label)
		}
	case "add":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: vp template add <file.json>\n")
			os.Exit(1)
		}
		addTemplate(args[1])
	case "show":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: vp template show <id>\n")
			os.Exit(1)
		}
		showTemplate(args[1])
	default:
		fmt.Fprintf(os.Stderr, "Unknown template command: %s\n", args[0])
		os.Exit(1)
	}
}

func handleResourceType(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: vp resource-type <list|add>\n")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		for name, rt := range state.Types {
			fmt.Printf("%-15s counter=%-5v check=%s\n", name, rt.Counter, rt.Check)
		}
	case "add":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: vp resource-type add <name> --check=<cmd> [--counter] [--start=N] [--end=N]\n")
			os.Exit(1)
		}
		addResourceType(args[1], args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown resource-type command: %s\n", args[0])
		os.Exit(1)
	}
}

func addTemplate(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var tmpl Template
	if err := json.Unmarshal(data, &tmpl); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	state.Templates[tmpl.ID] = &tmpl
	state.Save()

	fmt.Printf("Added template: %s\n", tmpl.ID)
}

func showTemplate(id string) {
	tmpl := state.Templates[id]
	if tmpl == nil {
		fmt.Fprintf(os.Stderr, "Template not found: %s\n", id)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(tmpl, "", "  ")
	fmt.Println(string(data))
}

func addResourceType(name string, args []string) {
	vars := parseVars(args)

	rt := &ResourceType{
		Name:    name,
		Check:   vars["check"],
		Counter: vars["counter"] == "true",
		Start:   0,
		End:     0,
	}

	if vars["start"] != "" {
		fmt.Sscanf(vars["start"], "%d", &rt.Start)
	}
	if vars["end"] != "" {
		fmt.Sscanf(vars["end"], "%d", &rt.End)
	}

	state.Types[name] = rt
	state.Save()

	fmt.Printf("Added resource type: %s\n", name)
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

		// Check if process is still running
		if inst.Status == "running" && !IsProcessRunning(inst.PID) {
			inst.Status = "stopped"
			inst.PID = 0
		}

		fmt.Printf("%-20s %-10s %-8d %-40s %s\n",
			name, inst.Status, inst.PID, truncate(inst.Command, 40), resources)
	}
}

func parseVars(args []string) map[string]string {
	vars := make(map[string]string)
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			parts := strings.SplitN(arg[2:], "=", 2)
			if len(parts) == 2 {
				vars[parts[0]] = parts[1]
			} else if len(parts) == 1 {
				vars[parts[0]] = "true"
			}
		}
	}
	return vars
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func handleDiscoverCLI(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: vp discover <pid> <name>\n")
		fmt.Fprintf(os.Stderr, "  Discovers a process by PID and imports it as a managed instance\n")
		os.Exit(1)
	}

	var pid int
	if _, err := fmt.Sscanf(args[0], "%d", &pid); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid PID: %s\n", args[0])
		os.Exit(1)
	}

	name := args[1]

	inst, err := DiscoverAndImportProcess(state, pid, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering process: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Discovered and imported process: %s\n", inst.Name)
	fmt.Printf("  PID:     %d\n", inst.PID)
	fmt.Printf("  Command: %s\n", inst.Command)

	if inst.LaunchScript != nil {
		fmt.Printf("\nLaunch script (child of shell):\n")
		fmt.Printf("  Command: %s\n", inst.LaunchScript.Cmdline)
		fmt.Printf("  CWD:     %s\n", inst.LaunchScript.Cwd)
		fmt.Printf("  Exe:     %s\n", inst.LaunchScript.Exe)
	}

	if len(inst.ParentChain) > 0 {
		fmt.Printf("\nParent chain:\n")
		for i, parent := range inst.ParentChain {
			fmt.Printf("  [%d] PID %d: %s (cwd: %s)\n", i, parent.PID, parent.Name, parent.Cwd)
		}
	}
}

func handleDiscoverPortCLI(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: vp discover-port <port> <name>\n")
		fmt.Fprintf(os.Stderr, "  Discovers a process listening on a port and imports it\n")
		os.Exit(1)
	}

	var port int
	if _, err := fmt.Sscanf(args[0], "%d", &port); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid port: %s\n", args[0])
		os.Exit(1)
	}

	name := args[1]

	inst, err := DiscoverAndImportProcessOnPort(state, port, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering process: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Discovered and imported process on port %d: %s\n", port, inst.Name)
	fmt.Printf("  PID:     %d\n", inst.PID)
	fmt.Printf("  Command: %s\n", inst.Command)

	if inst.LaunchScript != nil {
		fmt.Printf("\nLaunch script (child of shell):\n")
		fmt.Printf("  Command: %s\n", inst.LaunchScript.Cmdline)
		fmt.Printf("  CWD:     %s\n", inst.LaunchScript.Cwd)
		fmt.Printf("  Exe:     %s\n", inst.LaunchScript.Exe)
	}

	if len(inst.ParentChain) > 0 {
		fmt.Printf("\nParent chain:\n")
		for i, parent := range inst.ParentChain {
			fmt.Printf("  [%d] PID %d: %s (cwd: %s)\n", i, parent.PID, parent.Name, parent.Cwd)
		}
	}
}

func handleInspect(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: vp inspect <name>\n")
		fmt.Fprintf(os.Stderr, "  Shows detailed information about an instance\n")
		os.Exit(1)
	}

	name := args[0]
	inst := state.Instances[name]
	if inst == nil {
		fmt.Fprintf(os.Stderr, "Instance not found: %s\n", name)
		os.Exit(1)
	}

	// Pretty print the instance details
	data, _ := json.MarshalIndent(inst, "", "  ")
	fmt.Println(string(data))

	// Additional formatted output for better readability
	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Name:     %s\n", inst.Name)
	fmt.Printf("Status:   %s\n", inst.Status)
	fmt.Printf("PID:      %d\n", inst.PID)
	fmt.Printf("Template: %s\n", inst.Template)
	fmt.Printf("Command:  %s\n", inst.Command)

	if inst.Discovered {
		fmt.Printf("\n--- Discovery Information ---\n")
		fmt.Printf("This process was discovered (not started by vp)\n")

		if inst.LaunchScript != nil {
			fmt.Printf("\nLaunch Script (child of shell):\n")
			fmt.Printf("  PID:     %d\n", inst.LaunchScript.PID)
			fmt.Printf("  Name:    %s\n", inst.LaunchScript.Name)
			fmt.Printf("  Command: %s\n", inst.LaunchScript.Cmdline)
			fmt.Printf("  Exe:     %s\n", inst.LaunchScript.Exe)
			fmt.Printf("  CWD:     %s\n", inst.LaunchScript.Cwd)

			if len(inst.LaunchScript.Environ) > 0 {
				fmt.Printf("  Key Environment Variables:\n")
				for _, key := range []string{"PATH", "HOME", "USER", "PWD", "NODE_ENV", "PYTHON_ENV"} {
					if val, ok := inst.LaunchScript.Environ[key]; ok {
						fmt.Printf("    %s=%s\n", key, truncate(val, 80))
					}
				}
			}
		}

		if len(inst.ParentChain) > 0 {
			fmt.Printf("\nParent Process Chain:\n")
			for i, parent := range inst.ParentChain {
				fmt.Printf("  [%d] PID %d: %s\n", i, parent.PID, parent.Name)
				fmt.Printf("      Command: %s\n", truncate(parent.Cmdline, 70))
				fmt.Printf("      CWD:     %s\n", parent.Cwd)
				if parent.Exe != "" {
					fmt.Printf("      Exe:     %s\n", parent.Exe)
				}
			}
		}
	}

	if len(inst.Resources) > 0 {
		fmt.Printf("\n--- Resources ---\n")
		for k, v := range inst.Resources {
			fmt.Printf("  %s = %s\n", k, v)
		}
	}
}
