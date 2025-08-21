package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ataiva-software/chisel/pkg/core"
)

// ExecutionRecord represents an execution record for the UI
type ExecutionRecord struct {
	ID         string    `json:"id"`
	ModuleName string    `json:"module_name"`
	Action     string    `json:"action"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time,omitempty"`
	User       string    `json:"user"`
	Error      string    `json:"error,omitempty"`
	Duration   string    `json:"duration,omitempty"`
}

// WebUIServer provides a web interface for Chisel
type WebUIServer struct {
	addr       string
	modules    map[string]*core.Module
	executions []*ExecutionRecord
	mu         sync.RWMutex
	server     *http.Server
}

// NewWebUIServer creates a new web UI server
func NewWebUIServer(addr string) *WebUIServer {
	return &WebUIServer{
		addr:       addr,
		modules:    make(map[string]*core.Module),
		executions: make([]*ExecutionRecord, 0),
	}
}

// Start starts the web UI server
func (s *WebUIServer) Start() error {
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/api/health", s.withCORS(s.handleHealth))
	mux.HandleFunc("/api/modules", s.withCORS(s.handleModules))
	mux.HandleFunc("/api/modules/", s.withCORS(s.handleModuleDetail))
	mux.HandleFunc("/api/executions", s.withCORS(s.handleExecutions))
	mux.HandleFunc("/api/statistics", s.withCORS(s.handleStatistics))
	
	// Static files
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/static/", s.handleStatic)
	
	// 404 handler
	mux.HandleFunc("/api/", s.handleNotFound)
	
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}
	
	return s.server.ListenAndServe()
}

// Stop stops the web UI server
func (s *WebUIServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// AddModule adds a module to the UI
func (s *WebUIServer) AddModule(module *core.Module) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modules[module.Metadata.Name] = module
}

// AddExecution adds an execution record to the UI
func (s *WebUIServer) AddExecution(execution *ExecutionRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Calculate duration if execution is complete
	if !execution.EndTime.IsZero() {
		duration := execution.EndTime.Sub(execution.StartTime)
		execution.Duration = duration.String()
	}
	
	s.executions = append(s.executions, execution)
	
	// Keep only the last 100 executions
	if len(s.executions) > 100 {
		s.executions = s.executions[1:]
	}
}

// withCORS adds CORS headers to API responses
func (s *WebUIServer) withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		handler(w, r)
	}
}

// handleHealth handles health check requests
func (s *WebUIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}
	
	s.writeJSON(w, response)
}

// handleModules handles module list requests
func (s *WebUIServer) handleModules(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	modules := make([]map[string]interface{}, 0, len(s.modules))
	
	for _, module := range s.modules {
		moduleData := map[string]interface{}{
			"name":        module.Metadata.Name,
			"version":     module.Metadata.Version,
			"description": module.Metadata.Description,
			"labels":      module.Metadata.Labels,
			"resources":   len(module.Spec.Resources),
		}
		modules = append(modules, moduleData)
	}
	
	// Sort by name for consistent ordering
	sort.Slice(modules, func(i, j int) bool {
		return modules[i]["name"].(string) < modules[j]["name"].(string)
	})
	
	s.writeJSON(w, modules)
}

// handleModuleDetail handles individual module detail requests
func (s *WebUIServer) handleModuleDetail(w http.ResponseWriter, r *http.Request) {
	// Extract module name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/modules/")
	moduleName := strings.Split(path, "/")[0]
	
	s.mu.RLock()
	module, exists := s.modules[moduleName]
	s.mu.RUnlock()
	
	if !exists {
		http.NotFound(w, r)
		return
	}
	
	// Convert module to JSON-serializable format
	moduleData := map[string]interface{}{
		"apiVersion": module.APIVersion,
		"kind":       module.Kind,
		"metadata": map[string]interface{}{
			"name":        module.Metadata.Name,
			"version":     module.Metadata.Version,
			"description": module.Metadata.Description,
			"labels":      module.Metadata.Labels,
		},
		"spec": map[string]interface{}{
			"resources": module.Spec.Resources,
		},
	}
	
	s.writeJSON(w, moduleData)
}

// handleExecutions handles execution list requests
func (s *WebUIServer) handleExecutions(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return executions in reverse chronological order (most recent first)
	executions := make([]*ExecutionRecord, len(s.executions))
	for i, exec := range s.executions {
		executions[len(s.executions)-1-i] = exec
	}
	
	s.writeJSON(w, executions)
}

// handleStatistics handles statistics requests
func (s *WebUIServer) handleStatistics(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Calculate statistics
	totalModules := len(s.modules)
	totalExecutions := len(s.executions)
	
	// Count executions by status
	statusCounts := make(map[string]int)
	for _, exec := range s.executions {
		statusCounts[exec.Status]++
	}
	
	// Count resources by type
	resourceCounts := make(map[string]int)
	totalResources := 0
	for _, module := range s.modules {
		for _, resource := range module.Spec.Resources {
			resourceCounts[resource.Type]++
			totalResources++
		}
	}
	
	statistics := map[string]interface{}{
		"total_modules":     totalModules,
		"total_executions":  totalExecutions,
		"total_resources":   totalResources,
		"execution_status":  statusCounts,
		"resource_types":    resourceCounts,
		"last_updated":      time.Now().Format(time.RFC3339),
	}
	
	s.writeJSON(w, statistics)
}

// handleIndex handles the main dashboard page
func (s *WebUIServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Chisel Dashboard</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .header h1 {
            margin: 0;
            font-size: 2em;
        }
        .header p {
            margin: 5px 0 0 0;
            opacity: 0.9;
        }
        .dashboard {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
        }
        .card {
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .card h2 {
            margin-top: 0;
            color: #333;
        }
        .stat {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
        }
        .api-link {
            color: #667eea;
            text-decoration: none;
            font-family: monospace;
        }
        .api-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔨 Chisel Dashboard</h1>
        <p>Configuration Management & Infrastructure Orchestration</p>
    </div>
    
    <div class="dashboard">
        <div class="card">
            <h2>📊 Quick Stats</h2>
            <div id="stats">Loading...</div>
        </div>
        
        <div class="card">
            <h2>📦 Modules</h2>
            <div id="modules">Loading...</div>
        </div>
        
        <div class="card">
            <h2>⚡ Recent Executions</h2>
            <div id="executions">Loading...</div>
        </div>
        
        <div class="card">
            <h2>🔗 API Endpoints</h2>
            <ul>
                <li><a href="/api/health" class="api-link">/api/health</a> - Health check</li>
                <li><a href="/api/modules" class="api-link">/api/modules</a> - List modules</li>
                <li><a href="/api/executions" class="api-link">/api/executions</a> - List executions</li>
                <li><a href="/api/statistics" class="api-link">/api/statistics</a> - Statistics</li>
            </ul>
        </div>
    </div>

    <script>
        // Load statistics
        fetch('/api/statistics')
            .then(response => response.json())
            .then(data => {
                document.getElementById('stats').innerHTML = 
                    '<div class="stat">' + data.total_modules + '</div>Modules<br><br>' +
                    '<div class="stat">' + data.total_executions + '</div>Executions<br><br>' +
                    '<div class="stat">' + data.total_resources + '</div>Resources';
            })
            .catch(error => {
                document.getElementById('stats').innerHTML = 'Error loading stats';
            });

        // Load modules
        fetch('/api/modules')
            .then(response => response.json())
            .then(data => {
                const modulesList = data.slice(0, 5).map(module => 
                    '<div><strong>' + module.name + '</strong> v' + module.version + 
                    ' (' + module.resources + ' resources)</div>'
                ).join('');
                document.getElementById('modules').innerHTML = modulesList || 'No modules';
            })
            .catch(error => {
                document.getElementById('modules').innerHTML = 'Error loading modules';
            });

        // Load executions
        fetch('/api/executions')
            .then(response => response.json())
            .then(data => {
                const executionsList = data.slice(0, 5).map(exec => 
                    '<div><strong>' + exec.module_name + '</strong> ' + exec.action + 
                    ' - ' + exec.status + '</div>'
                ).join('');
                document.getElementById('executions').innerHTML = executionsList || 'No executions';
            })
            .catch(error => {
                document.getElementById('executions').innerHTML = 'Error loading executions';
            });
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleStatic handles static file requests
func (s *WebUIServer) handleStatic(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would serve static files
	// For now, return a simple response
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Static files not implemented"))
}

// handleNotFound handles 404 requests
func (s *WebUIServer) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	
	response := map[string]interface{}{
		"error":   "Not Found",
		"message": "The requested endpoint does not exist",
		"path":    r.URL.Path,
	}
	
	json.NewEncoder(w).Encode(response)
}

// writeJSON writes a JSON response
func (s *WebUIServer) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("JSON encoding error: %v", err), http.StatusInternalServerError)
	}
}

// WebUIConfig represents web UI configuration
type WebUIConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Address string `yaml:"address" json:"address"`
	Port    int    `yaml:"port" json:"port"`
}

// DefaultWebUIConfig returns default web UI configuration
func DefaultWebUIConfig() *WebUIConfig {
	return &WebUIConfig{
		Enabled: false, // Disabled by default for security
		Address: "127.0.0.1",
		Port:    8080,
	}
}
