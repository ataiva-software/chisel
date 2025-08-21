package templating

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

// TemplateEngine provides template rendering capabilities
type TemplateEngine struct {
	functions template.FuncMap
}

// NewTemplateEngine creates a new template engine with built-in functions
func NewTemplateEngine() *TemplateEngine {
	engine := &TemplateEngine{
		functions: make(template.FuncMap),
	}
	
	// Add built-in functions
	engine.addBuiltinFunctions()
	
	return engine
}

// addBuiltinFunctions adds commonly used template functions
func (te *TemplateEngine) addBuiltinFunctions() {
	te.functions["default"] = func(defaultValue interface{}, value interface{}) interface{} {
		if value == nil || value == "" {
			return defaultValue
		}
		return value
	}
	
	te.functions["join"] = func(sep string, items []string) string {
		return strings.Join(items, sep)
	}
	
	te.functions["upper"] = strings.ToUpper
	te.functions["lower"] = strings.ToLower
	te.functions["title"] = strings.Title
	
	te.functions["contains"] = strings.Contains
	te.functions["hasPrefix"] = strings.HasPrefix
	te.functions["hasSuffix"] = strings.HasSuffix
	
	te.functions["replace"] = func(old, new, s string) string {
		return strings.ReplaceAll(s, old, new)
	}
}

// AddFunction adds a custom function to the template engine
func (te *TemplateEngine) AddFunction(name string, fn interface{}) {
	te.functions[name] = fn
}

// Render renders a template string with the given variables
func (te *TemplateEngine) Render(templateStr string, vars map[string]interface{}) (string, error) {
	tmpl, err := template.New("template").Funcs(te.functions).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	var buf strings.Builder
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// RenderFile renders a template file with the given variables
func (te *TemplateEngine) RenderFile(filename string, vars map[string]interface{}) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", filename, err)
	}
	
	return te.Render(string(content), vars)
}

// RenderToFile renders a template and writes the result to a file
func (te *TemplateEngine) RenderToFile(templateStr string, vars map[string]interface{}, outputFile string) error {
	result, err := te.Render(templateStr, vars)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
		return fmt.Errorf("failed to write rendered template to %s: %w", outputFile, err)
	}
	
	return nil
}

// RenderFileToFile renders a template file and writes the result to another file
func (te *TemplateEngine) RenderFileToFile(templateFile string, vars map[string]interface{}, outputFile string) error {
	result, err := te.RenderFile(templateFile, vars)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
		return fmt.Errorf("failed to write rendered template to %s: %w", outputFile, err)
	}
	
	return nil
}
