package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Forge project",
	Long: `Initialize a new Forge project with the basic directory structure
and configuration files needed to get started.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := "chisel-project"
	if len(args) > 0 {
		projectName = args[0]
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		"modules",
		"inventory",
		"plans",
		"templates",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(projectName, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	// Create chisel.yaml config file
	config := ProjectConfig{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Project",
		Metadata: Metadata{
			Name:        projectName,
			Description: "A Forge configuration management project",
		},
		Spec: ProjectSpec{
			ModulePath: "./modules",
			Inventory:  "./inventory",
			Templates:  "./templates",
		},
	}

	configPath := filepath.Join(projectName, "chisel.yaml")
	if err := writeYAMLFile(configPath, config); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Create example inventory file
	inventory := InventoryConfig{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Inventory",
		Targets: map[string]TargetGroup{
			"webservers": {
				Selector: "role=web,env=prod",
				Connection: ConnectionConfig{
					Host:           "web1.example.com",
					User:           "ubuntu",
					PrivateKeyPath: "~/.ssh/id_rsa",
					Port:           22,
				},
			},
			"databases": {
				Hosts: []string{
					"db1.example.com",
					"db2.example.com",
				},
				Connection: ConnectionConfig{
					Host:           "db1.example.com",
					User:           "admin",
					PrivateKeyPath: "~/.ssh/id_rsa",
					Port:           22,
				},
			},
		},
	}

	inventoryPath := filepath.Join(projectName, "inventory", "hosts.yaml")
	if err := writeYAMLFile(inventoryPath, inventory); err != nil {
		return fmt.Errorf("failed to create inventory file: %w", err)
	}

	// Create example module
	module := ModuleConfig{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: Metadata{
			Name:        "webserver",
			Version:     "1.0.0",
			Description: "Basic web server configuration",
		},
		Spec: ModuleSpec{
			Resources: []ResourceConfig{
				{
					Type:  "pkg",
					Name:  "nginx",
					State: "present",
				},
				{
					Type: "file",
					Name: "nginx-config",
					Properties: map[string]interface{}{
						"path": "/etc/nginx/sites-enabled/default",
						"content": `server {
    listen 80 default_server;
    listen [::]:80 default_server;
    
    root /var/www/html;
    index index.html index.htm index.nginx-debian.html;
    
    server_name _;
    
    location / {
        try_files $uri $uri/ =404;
    }
}`,
						"mode":  "0644",
						"owner": "root",
						"group": "root",
					},
					Notify: []string{"restart:nginx"},
				},
				{
					Type:    "service",
					Name:    "nginx",
					State:   "running",
					Enabled: true,
				},
			},
		},
	}

	modulePath := filepath.Join(projectName, "modules", "webserver.yaml")
	if err := writeYAMLFile(modulePath, module); err != nil {
		return fmt.Errorf("failed to create example module: %w", err)
	}

	// Create README
	readme := fmt.Sprintf(`# %s

This is a Forge configuration management project.

## Getting Started

1. Edit the inventory file to define your target hosts:
   ` + "`inventory/hosts.yaml`" + `

2. Customize the example module or create new ones:
   ` + "`modules/webserver.yaml`" + `

3. Plan and apply workflow:
   ` + "`chisel plan --module modules/webserver.yaml --inventory inventory/hosts.yaml`" + `
   ` + "`chisel apply --module modules/webserver.yaml --inventory inventory/hosts.yaml`" + `

## Current Status

Forge is in active development. Currently implemented:
- ✅ Project initialization (` + "`chisel init`" + `)
- ✅ Resource type system with file provider
- ✅ SSH connection management
- ✅ Plan/apply workflow
- ✅ Module loading and execution
- ✅ Static inventory support
- 🚧 Additional providers (pkg, service, user)
- 🚧 Dynamic inventory (cloud APIs)
- 🚧 Templating system

## Project Structure

- ` + "`chisel.yaml`" + ` - Project configuration
- ` + "`inventory/`" + ` - Target host definitions
- ` + "`modules/`" + ` - Reusable configuration modules
- ` + "`templates/`" + ` - Configuration templates
- ` + "`plans/`" + ` - Generated execution plans (coming soon)

## Documentation

Visit https://github.com/ataiva-software/forge for more information.
`, projectName)

	readmePath := filepath.Join(projectName, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	fmt.Printf("✅ Initialized Forge project '%s'\n", projectName)
	fmt.Printf("📁 Project structure:\n")
	fmt.Printf("   %s/\n", projectName)
	fmt.Printf("   ├── chisel.yaml\n")
	fmt.Printf("   ├── README.md\n")
	fmt.Printf("   ├── inventory/\n")
	fmt.Printf("   │   └── hosts.yaml\n")
	fmt.Printf("   ├── modules/\n")
	fmt.Printf("   │   └── webserver.yaml\n")
	fmt.Printf("   ├── templates/\n")
	fmt.Printf("   └── plans/\n")
	fmt.Printf("\n")
	fmt.Printf("🚀 Next steps:\n")
	fmt.Printf("   cd %s\n", projectName)
	fmt.Printf("   # Edit inventory/hosts.yaml with your target hosts\n")
	fmt.Printf("   # Customize modules/webserver.yaml for your needs\n")
	fmt.Printf("   # Create a plan: chisel plan --module modules/webserver.yaml --inventory inventory/hosts.yaml\n")
	fmt.Printf("   # Apply changes: chisel apply --module modules/webserver.yaml --inventory inventory/hosts.yaml\n")

	return nil
}

func writeYAMLFile(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	defer encoder.Close()

	return encoder.Encode(data)
}

// Configuration structures
type ProjectConfig struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       ProjectSpec `yaml:"spec"`
}

type Metadata struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type ProjectSpec struct {
	ModulePath string `yaml:"module_path"`
	Inventory  string `yaml:"inventory"`
	Templates  string `yaml:"templates"`
}

type InventoryConfig struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Targets    map[string]TargetGroup `yaml:"targets"`
}

type TargetGroup struct {
	Selector   string           `yaml:"selector,omitempty"`
	Hosts      []string         `yaml:"hosts,omitempty"`
	Connection ConnectionConfig `yaml:"connection"`
}

type ConnectionConfig struct {
	Host           string `yaml:"host"`
	User           string `yaml:"user"`
	PrivateKeyPath string `yaml:"private_key_path,omitempty"`
	Port           int    `yaml:"port,omitempty"`
}

type ModuleConfig struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Spec       ModuleSpec `yaml:"spec"`
}

type ModuleSpec struct {
	Resources []ResourceConfig `yaml:"resources"`
}

type ResourceConfig struct {
	Type       string                 `yaml:"type"`
	Name       string                 `yaml:"name"`
	State      string                 `yaml:"state,omitempty"`
	Enabled    bool                   `yaml:"enabled,omitempty"`
	Properties map[string]interface{} `yaml:",inline,omitempty"`
	DependsOn  []string               `yaml:"depends_on,omitempty"`
	Notify     []string               `yaml:"notify,omitempty"`
}
