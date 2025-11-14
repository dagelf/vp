package main

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed web.html
var webHTML string

// ServeHTTP starts the HTTP server
func ServeHTTP(addr string) error {
	// Web UI
	http.HandleFunc("/", serveWeb)

	// API endpoints
	http.HandleFunc("/api/instances", handleInstances)
	http.HandleFunc("/api/templates", handleTemplates)
	http.HandleFunc("/api/resources", handleResources)
	http.HandleFunc("/api/resource-types", handleResourceTypes)

	return http.ListenAndServe(addr, nil)
}

func serveWeb(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(webHTML))
}

func handleInstances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Update status for all instances
		for _, inst := range state.Instances {
			if inst.Status == "running" && !IsProcessRunning(inst.PID) {
				inst.Status = "stopped"
				inst.PID = 0
			}
		}
		json.NewEncoder(w).Encode(state.Instances)

	case "POST":
		var req struct {
			Action     string            `json:"action"` // "start" or "stop"
			Template   string            `json:"template"`
			Name       string            `json:"name"`
			Vars       map[string]string `json:"vars"`
			InstanceID string            `json:"instance_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Action == "start" {
			tmpl := state.Templates[req.Template]
			if tmpl == nil {
				http.Error(w, "template not found", http.StatusNotFound)
				return
			}

			inst, err := StartProcess(state, tmpl, req.Name, req.Vars)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			json.NewEncoder(w).Encode(inst)
		} else if req.Action == "stop" {
			inst := state.Instances[req.InstanceID]
			if inst == nil {
				http.Error(w, "instance not found", http.StatusNotFound)
				return
			}

			if err := StopProcess(state, inst); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			state.ReleaseResources(req.InstanceID)
			delete(state.Instances, req.InstanceID)
			state.Save()

			json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
		} else {
			http.Error(w, "invalid action", http.StatusBadRequest)
		}

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(state.Templates)

	case "POST":
		var tmpl Template
		if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		state.Templates[tmpl.ID] = &tmpl
		state.Save()

		json.NewEncoder(w).Encode(tmpl)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleResources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		// Group resources by type for better display
		grouped := make(map[string][]Resource)
		for _, res := range state.Resources {
			grouped[res.Type] = append(grouped[res.Type], *res)
		}
		json.NewEncoder(w).Encode(grouped)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleResourceTypes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(state.Types)

	case "POST":
		var rt ResourceType
		if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate required fields
		if rt.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		// Convert name to lowercase for consistency
		rt.Name = strings.ToLower(rt.Name)

		state.Types[rt.Name] = &rt
		state.Save()

		json.NewEncoder(w).Encode(rt)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper function to get path parameter
func getPathParam(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}
