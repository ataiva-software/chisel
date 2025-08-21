package templating

import (
	"strings"
	"testing"
)

func TestTemplateEngine_Render(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "simple variable substitution",
			template: "Hello {{.name}}!",
			vars:     map[string]interface{}{"name": "World"},
			want:     "Hello World!",
			wantErr:  false,
		},
		{
			name:     "multiple variables",
			template: "{{.greeting}} {{.name}}, you have {{.count}} messages",
			vars: map[string]interface{}{
				"greeting": "Hello",
				"name":     "Alice",
				"count":    5,
			},
			want:    "Hello Alice, you have 5 messages",
			wantErr: false,
		},
		{
			name:     "conditional rendering",
			template: "{{if .enabled}}Service is enabled{{else}}Service is disabled{{end}}",
			vars:     map[string]interface{}{"enabled": true},
			want:     "Service is enabled",
			wantErr:  false,
		},
		{
			name:     "loop rendering",
			template: "{{range .items}}{{.}} {{end}}",
			vars:     map[string]interface{}{"items": []string{"a", "b", "c"}},
			want:     "a b c ",
			wantErr:  false,
		},
		{
			name:     "missing variable",
			template: "Hello {{.missing}}!",
			vars:     map[string]interface{}{"name": "World"},
			want:     "Hello <no value>!",
			wantErr:  false,
		},
		{
			name:     "invalid template syntax",
			template: "Hello {{.name}!",
			vars:     map[string]interface{}{"name": "World"},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "nested object access",
			template: "User: {{.user.name}} ({{.user.email}})",
			vars: map[string]interface{}{
				"user": map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
				},
			},
			want:    "User: John Doe (john@example.com)",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTemplateEngine()
			got, err := engine.Render(tt.template, tt.vars)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("TemplateEngine.Render() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("TemplateEngine.Render() unexpected error = %v", err)
				return
			}
			
			if got != tt.want {
				t.Errorf("TemplateEngine.Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateEngine_RenderFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		vars     map[string]interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "render template file",
			filename: "testdata/simple.tmpl",
			vars:     map[string]interface{}{"name": "Chisel"},
			want:     "Welcome to Chisel!\n",
			wantErr:  false,
		},
		{
			name:     "file not found",
			filename: "testdata/nonexistent.tmpl",
			vars:     map[string]interface{}{},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTemplateEngine()
			got, err := engine.RenderFile(tt.filename, tt.vars)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("TemplateEngine.RenderFile() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("TemplateEngine.RenderFile() unexpected error = %v", err)
				return
			}
			
			if got != tt.want {
				t.Errorf("TemplateEngine.RenderFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateEngine_AddFunction(t *testing.T) {
	engine := NewTemplateEngine()
	
	// Add a custom function
	engine.AddFunction("upper", func(s string) string {
		return strings.ToUpper(s)
	})
	
	template := "{{upper .name}}"
	vars := map[string]interface{}{"name": "hello"}
	
	got, err := engine.Render(template, vars)
	if err != nil {
		t.Errorf("TemplateEngine.Render() unexpected error = %v", err)
		return
	}
	
	want := "HELLO"
	if got != want {
		t.Errorf("TemplateEngine.Render() = %v, want %v", got, want)
	}
}

func TestTemplateEngine_WithBuiltinFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "default function",
			template: "{{default \"unknown\" .name}}",
			vars:     map[string]interface{}{},
			want:     "unknown",
			wantErr:  false,
		},
		{
			name:     "default with value",
			template: "{{default \"unknown\" .name}}",
			vars:     map[string]interface{}{"name": "Alice"},
			want:     "Alice",
			wantErr:  false,
		},
		{
			name:     "join function",
			template: "{{join \",\" .items}}",
			vars:     map[string]interface{}{"items": []string{"a", "b", "c"}},
			want:     "a,b,c",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTemplateEngine()
			got, err := engine.Render(tt.template, tt.vars)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("TemplateEngine.Render() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("TemplateEngine.Render() unexpected error = %v", err)
				return
			}
			
			if got != tt.want {
				t.Errorf("TemplateEngine.Render() = %v, want %v", got, tt.want)
			}
		})
	}
}
