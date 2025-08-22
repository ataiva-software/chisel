package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ataiva-software/forge/pkg/core"
	"github.com/ataiva-software/forge/pkg/types"
)

func TestWebUIServer_New(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	if server == nil {
		t.Fatal("Expected non-nil web UI server")
	}
	
	if server.addr != ":8080" {
		t.Errorf("Expected address ':8080', got '%s'", server.addr)
	}
}

func TestWebUIServer_HealthEndpoint(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	
	server.handleHealth(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}
}

func TestWebUIServer_ModulesEndpoint(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	// Add test modules
	module1 := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name:    "web-server",
			Version: "1.0.0",
			Labels: map[string]string{
				"environment": "production",
			},
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{Type: "pkg", Name: "nginx"},
				{Type: "service", Name: "nginx"},
			},
		},
	}
	
	module2 := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name:    "database",
			Version: "1.0.0",
			Labels: map[string]string{
				"environment": "production",
			},
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{Type: "pkg", Name: "postgresql"},
				{Type: "service", Name: "postgresql"},
			},
		},
	}
	
	server.AddModule(module1)
	server.AddModule(module2)
	
	req := httptest.NewRequest("GET", "/api/modules", nil)
	w := httptest.NewRecorder()
	
	server.handleModules(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if len(response) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(response))
	}
	
	// Check first module (alphabetically sorted, so "database" comes first)
	if response[0]["name"] != "database" {
		t.Errorf("Expected module name 'database', got '%v'", response[0]["name"])
	}
}

func TestWebUIServer_ModuleDetailEndpoint(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name:    "test-module",
			Version: "1.0.0",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{Type: "file", Name: "config"},
			},
		},
	}
	
	server.AddModule(module)
	
	req := httptest.NewRequest("GET", "/api/modules/test-module", nil)
	w := httptest.NewRecorder()
	
	server.handleModuleDetail(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if metadata, ok := response["metadata"].(map[string]interface{}); ok {
		if metadata["name"] != "test-module" {
			t.Errorf("Expected module name 'test-module', got '%v'", metadata["name"])
		}
	} else {
		t.Fatal("Expected metadata to be a map")
	}
	
	// Check resources
	spec, ok := response["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected spec to be a map")
	}
	
	resources, ok := spec["resources"].([]interface{})
	if !ok {
		t.Fatal("Expected resources to be an array")
	}
	
	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}
}

func TestWebUIServer_ExecutionsEndpoint(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	// Add test executions
	execution1 := &ExecutionRecord{
		ID:         "exec-1",
		ModuleName: "web-server",
		Action:     "apply",
		Status:     "completed",
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(-30 * time.Minute),
		User:       "admin",
	}
	
	execution2 := &ExecutionRecord{
		ID:         "exec-2",
		ModuleName: "database",
		Action:     "plan",
		Status:     "running",
		StartTime:  time.Now().Add(-10 * time.Minute),
		User:       "operator",
	}
	
	server.AddExecution(execution1)
	server.AddExecution(execution2)
	
	req := httptest.NewRequest("GET", "/api/executions", nil)
	w := httptest.NewRecorder()
	
	server.handleExecutions(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if len(response) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(response))
	}
	
	// Check first execution (should be most recent)
	if response[0]["id"] != "exec-2" {
		t.Errorf("Expected execution ID 'exec-2', got '%v'", response[0]["id"])
	}
}

func TestWebUIServer_StatisticsEndpoint(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	// Add test data
	module := &core.Module{
		Metadata: core.ModuleMetadata{Name: "test"},
		Spec:     core.ModuleSpec{Resources: []types.Resource{{Type: "file", Name: "test"}}},
	}
	server.AddModule(module)
	
	execution := &ExecutionRecord{
		ID:         "exec-1",
		ModuleName: "test",
		Action:     "apply",
		Status:     "completed",
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(-30 * time.Minute),
	}
	server.AddExecution(execution)
	
	req := httptest.NewRequest("GET", "/api/statistics", nil)
	w := httptest.NewRecorder()
	
	server.handleStatistics(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if response["total_modules"] != float64(1) {
		t.Errorf("Expected total_modules 1, got %v", response["total_modules"])
	}
	
	if response["total_executions"] != float64(1) {
		t.Errorf("Expected total_executions 1, got %v", response["total_executions"])
	}
}

func TestWebUIServer_StaticFiles(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	server.handleIndex(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "Chisel Dashboard") {
		t.Error("Expected response to contain 'Chisel Dashboard'")
	}
	
	if !strings.Contains(body, "html") {
		t.Error("Expected response to contain HTML")
	}
}

func TestWebUIServer_CORSHeaders(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	req := httptest.NewRequest("OPTIONS", "/api/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	
	// Use the CORS-wrapped handler
	corsHandler := server.withCORS(server.handleHealth)
	corsHandler(w, req)
	
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers to be set")
	}
	
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected CORS methods to be set")
	}
}

func TestWebUIServer_JSONResponse(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	
	server.handleHealth(w, req)
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

func TestWebUIServer_NotFoundEndpoint(t *testing.T) {
	server := NewWebUIServer(":8080")
	
	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	w := httptest.NewRecorder()
	
	// This would be handled by the router's 404 handler
	server.handleNotFound(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}
