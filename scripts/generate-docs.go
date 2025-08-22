package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

// PackageDoc represents documentation for a package
type PackageDoc struct {
	Name        string
	ImportPath  string
	Synopsis    string
	Doc         string
	Types       []TypeDoc
	Functions   []FunctionDoc
	Variables   []VariableDoc
	Constants   []ConstantDoc
}

// TypeDoc represents documentation for a type
type TypeDoc struct {
	Name    string
	Doc     string
	Methods []FunctionDoc
}

// FunctionDoc represents documentation for a function
type FunctionDoc struct {
	Name      string
	Doc       string
	Signature string
}

// VariableDoc represents documentation for a variable
type VariableDoc struct {
	Name string
	Doc  string
	Type string
}

// ConstantDoc represents documentation for a constant
type ConstantDoc struct {
	Name  string
	Doc   string
	Value string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <output-dir>\n", os.Args[0])
		os.Exit(1)
	}

	outputDir := os.Args[1]
	
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate documentation
	if err := generateDocs(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating documentation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Documentation generated successfully in %s\n", outputDir)
}

func generateDocs(outputDir string) error {
	// Parse packages
	packages, err := parsePackages("./pkg")
	if err != nil {
		return fmt.Errorf("failed to parse packages: %w", err)
	}

	// Generate API reference
	if err := generateAPIReference(outputDir, packages); err != nil {
		return fmt.Errorf("failed to generate API reference: %w", err)
	}

	// Generate getting started guide
	if err := generateGettingStarted(outputDir); err != nil {
		return fmt.Errorf("failed to generate getting started guide: %w", err)
	}

	// Generate user guide
	if err := generateUserGuide(outputDir); err != nil {
		return fmt.Errorf("failed to generate user guide: %w", err)
	}

	// Generate provider documentation
	if err := generateProviderDocs(outputDir); err != nil {
		return fmt.Errorf("failed to generate provider documentation: %w", err)
	}

	// Generate index
	if err := generateIndex(outputDir, packages); err != nil {
		return fmt.Errorf("failed to generate index: %w", err)
	}

	return nil
}

func parsePackages(rootDir string) ([]PackageDoc, error) {
	var packages []PackageDoc

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip hidden directories and testdata
		if strings.HasPrefix(info.Name(), ".") || info.Name() == "testdata" {
			return filepath.SkipDir
		}

		// Parse package
		pkg, err := parsePackage(path)
		if err != nil {
			// Skip directories without Go files
			return nil
		}

		if pkg != nil {
			packages = append(packages, *pkg)
		}

		return nil
	})

	return packages, err
}

func parsePackage(dir string) (*PackageDoc, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for name, pkg := range pkgs {
		// Skip test packages
		if strings.HasSuffix(name, "_test") {
			continue
		}

		docPkg := doc.New(pkg, dir, doc.AllDecls)
		
		packageDoc := &PackageDoc{
			Name:       name,
			ImportPath: strings.TrimPrefix(dir, "./"),
			Synopsis:   docPkg.Synopsis(docPkg.Doc),
			Doc:        docPkg.Doc,
		}

		// Parse types
		for _, t := range docPkg.Types {
			typeDoc := TypeDoc{
				Name: t.Name,
				Doc:  t.Doc,
			}

			// Parse methods
			for _, m := range t.Methods {
				typeDoc.Methods = append(typeDoc.Methods, FunctionDoc{
					Name:      m.Name,
					Doc:       m.Doc,
					Signature: formatSignature(m.Decl),
				})
			}

			packageDoc.Types = append(packageDoc.Types, typeDoc)
		}

		// Parse functions
		for _, f := range docPkg.Funcs {
			packageDoc.Functions = append(packageDoc.Functions, FunctionDoc{
				Name:      f.Name,
				Doc:       f.Doc,
				Signature: formatSignature(f.Decl),
			})
		}

		// Parse variables
		for _, v := range docPkg.Vars {
			for _, name := range v.Names {
				packageDoc.Variables = append(packageDoc.Variables, VariableDoc{
					Name: name,
					Doc:  v.Doc,
					Type: formatType(v.Decl),
				})
			}
		}

		// Parse constants
		for _, c := range docPkg.Consts {
			for _, name := range c.Names {
				packageDoc.Constants = append(packageDoc.Constants, ConstantDoc{
					Name:  name,
					Doc:   c.Doc,
					Value: formatValue(c.Decl),
				})
			}
		}

		return packageDoc, nil
	}

	return nil, nil
}

func formatSignature(decl *ast.FuncDecl) string {
	// Simple signature formatting - in a real implementation,
	// you'd want more sophisticated formatting
	return fmt.Sprintf("func %s(...)", decl.Name.Name)
}

func formatType(decl *ast.GenDecl) string {
	return "type"
}

func formatValue(decl *ast.GenDecl) string {
	return "value"
}

func generateAPIReference(outputDir string, packages []PackageDoc) error {
	tmpl := `# API Reference

Generated on {{ .Timestamp }}

## Packages

{{ range .Packages }}
### {{ .Name }}

**Import Path:** ` + "`{{ .ImportPath }}`" + `

{{ .Doc }}

{{ if .Types }}
#### Types

{{ range .Types }}
##### {{ .Name }}

{{ .Doc }}

{{ if .Methods }}
**Methods:**

{{ range .Methods }}
- ` + "`{{ .Name }}`" + `: {{ .Doc }}
{{ end }}
{{ end }}
{{ end }}
{{ end }}

{{ if .Functions }}
#### Functions

{{ range .Functions }}
##### {{ .Name }}

{{ .Doc }}

{{ end }}
{{ end }}

{{ if .Constants }}
#### Constants

{{ range .Constants }}
- ` + "`{{ .Name }}`" + `: {{ .Doc }}
{{ end }}
{{ end }}

{{ if .Variables }}
#### Variables

{{ range .Variables }}
- ` + "`{{ .Name }}`" + `: {{ .Doc }}
{{ end }}
{{ end }}

---

{{ end }}
`

	t, err := template.New("api").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(outputDir, "api-reference.md"))
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		Timestamp string
		Packages  []PackageDoc
	}{
		Timestamp: time.Now().Format("2006-01-02 15:04:05 UTC"),
		Packages:  packages,
	}

	return t.Execute(file, data)
}

func generateGettingStarted(outputDir string) error {
	content := `# Getting Started with Chisel

## Installation

### From Source

` + "```bash" + `
git clone https://github.com/ataiva-software/forge.git
cd chisel
make build
` + "```" + `

The binary will be available at ` + "`./bin/chisel`" + `.

### Verify Installation

` + "```bash" + `
./bin/chisel --version
` + "```" + `

## Quick Start

### 1. Create Your First Module

Create a simple file management module:

` + "```yaml" + `
# hello.yaml
apiVersion: ataiva.com/chisel/v1
kind: Module
metadata:
  name: hello-world
  version: 1.0.0
spec:
  resources:
    - type: file
      name: hello-file
      state: present
      path: /tmp/hello.txt
      content: "Hello from Chisel!"
      mode: "0644"
` + "```" + `

### 2. Plan Your Changes

See what Chisel will do before applying:

` + "```bash" + `
./bin/chisel plan --module hello.yaml
` + "```" + `

### 3. Apply Your Configuration

Apply the changes:

` + "```bash" + `
./bin/chisel apply --module hello.yaml
` + "```" + `

### 4. Verify the Result

` + "```bash" + `
cat /tmp/hello.txt
` + "```" + `

## Core Concepts

### Resources

Resources are the fundamental units in Chisel. Each resource has:
- **Type**: What kind of resource (file, pkg, service, user, shell)
- **Name**: Unique identifier within the module
- **State**: Desired state (present, absent, running, stopped)
- **Properties**: Resource-specific configuration

### Modules

Modules are YAML files that define a collection of resources and their desired state.

### Providers

Providers implement the logic for managing specific resource types:
- **file**: File and directory management
- **pkg**: Package installation and removal
- **service**: System service management
- **user**: User and group management
- **shell**: Command execution

### Plan/Apply Workflow

1. **Plan**: Analyze current state vs desired state
2. **Apply**: Execute changes to reach desired state

## Next Steps

- Read the [User Guide](user-guide.md) for detailed usage
- Check out [Provider Documentation](providers.md) for resource types
- See [Examples](../examples/) for real-world scenarios
- Review [API Reference](api-reference.md) for development
`

	return os.WriteFile(filepath.Join(outputDir, "getting-started.md"), []byte(content), 0644)
}

func generateUserGuide(outputDir string) error {
	content := `# User Guide

## Module Structure

### Basic Module

` + "```yaml" + `
apiVersion: ataiva.com/chisel/v1
kind: Module
metadata:
  name: my-module
  version: 1.0.0
  description: Description of what this module does
spec:
  resources:
    - type: file
      name: config-file
      state: present
      path: /etc/myapp/config.yml
      content: |
        key: value
        debug: true
      mode: "0644"
      owner: root
      group: root
` + "```" + `

### Resource Properties

Each resource type supports different properties. Common properties:

- ` + "`state`" + `: Desired state (present, absent, running, stopped)
- ` + "`name`" + `: Unique identifier within the module

## Resource Types

### File Resources

Manage files and directories:

` + "```yaml" + `
- type: file
  name: my-file
  state: present
  path: /path/to/file
  content: "File content"
  mode: "0644"
  owner: user
  group: group
` + "```" + `

### Package Resources

Install and manage packages:

` + "```yaml" + `
- type: pkg
  name: nginx
  state: present
` + "```" + `

### Service Resources

Manage system services:

` + "```yaml" + `
- type: service
  name: nginx
  state: running
  enabled: true
` + "```" + `

### User Resources

Manage users and groups:

` + "```yaml" + `
- type: user
  name: webuser
  state: present
  uid: 1001
  gid: 1001
  home: /home/webuser
  shell: /bin/bash
  groups:
    - sudo
    - www-data
` + "```" + `

### Shell Resources

Execute commands:

` + "```yaml" + `
- type: shell
  name: setup-script
  command: /path/to/setup.sh
  creates: /path/to/marker/file
  user: root
` + "```" + `

## Inventory Management

### Static Inventory

` + "```yaml" + `
# inventory.yaml
apiVersion: ataiva.com/chisel/v1
kind: Inventory
metadata:
  name: my-inventory
spec:
  targets:
    - host: server1.example.com
      user: ubuntu
      key_file: ~/.ssh/id_rsa
      labels:
        role: web
        env: prod
    - host: server2.example.com
      user: ubuntu
      key_file: ~/.ssh/id_rsa
      labels:
        role: db
        env: prod
` + "```" + `

### Using Inventory

` + "```bash" + `
chisel plan --module module.yaml --inventory inventory.yaml
chisel apply --module module.yaml --inventory inventory.yaml
` + "```" + `

## Advanced Features

### Templating

Use Go templates in file content:

` + "```yaml" + `
- type: file
  name: templated-config
  state: present
  path: /etc/app/config.yml
  content: |
    server_name: {{ .hostname }}
    environment: {{ .env | default "production" }}
    debug: {{ if eq .env "development" }}true{{ else }}false{{ end }}
` + "```" + `

### Conditional Execution

Shell resources support conditional execution:

` + "```yaml" + `
- type: shell
  name: conditional-command
  command: echo "Running setup"
  creates: /var/lib/app/setup.done
  unless: test -f /var/lib/app/skip-setup
` + "```" + `

## Best Practices

### Module Organization

- Keep modules focused on a single purpose
- Use descriptive names for resources
- Group related resources together
- Use comments to explain complex logic

### Security

- Use SSH keys instead of passwords
- Set appropriate file permissions
- Run services as non-root users when possible
- Validate input in shell commands

### Performance

- Use ` + "`creates`" + ` and ` + "`unless`" + ` conditions to avoid unnecessary work
- Group independent resources to enable parallel execution
- Keep modules small and focused

## Troubleshooting

### Common Issues

1. **Permission Denied**: Check SSH keys and user permissions
2. **Package Not Found**: Verify package names for your distribution
3. **Service Won't Start**: Check service dependencies and configuration
4. **Template Errors**: Validate template syntax and variable names

### Debug Mode

Run with verbose output:

` + "```bash" + `
chisel plan --module module.yaml --verbose
chisel apply --module module.yaml --verbose
` + "```" + `

### Dry Run

Test changes without applying:

` + "```bash" + `
chisel apply --module module.yaml --dry-run
` + "```" + `
`

	return os.WriteFile(filepath.Join(outputDir, "user-guide.md"), []byte(content), 0644)
}

func generateProviderDocs(outputDir string) error {
	content := `# Provider Documentation

## File Provider

Manages files and directories on target systems.

### Properties

- ` + "`path`" + ` (required): Path to the file or directory
- ` + "`state`" + `: present (default) or absent
- ` + "`content`" + `: File content (for files)
- ` + "`source`" + `: Source file to copy
- ` + "`template`" + `: Template file to render
- ` + "`mode`" + `: File permissions (e.g., "0644")
- ` + "`owner`" + `: File owner
- ` + "`group`" + `: File group
- ` + "`file_type`" + `: file (default) or directory

### Examples

` + "```yaml" + `
# Create a file with content
- type: file
  name: config-file
  state: present
  path: /etc/app/config.yml
  content: |
    debug: true
    port: 8080
  mode: "0644"
  owner: app
  group: app

# Create a directory
- type: file
  name: app-dir
  state: present
  path: /opt/app
  file_type: directory
  mode: "0755"
  owner: app
  group: app

# Copy a file
- type: file
  name: binary-file
  state: present
  path: /usr/local/bin/app
  source: ./dist/app
  mode: "0755"
  owner: root
  group: root
` + "```" + `

## Package Provider

Manages system packages using native package managers.

### Properties

- ` + "`state`" + `: present (default) or absent
- ` + "`version`" + `: Specific version to install (optional)

### Supported Package Managers

- **apt** (Debian/Ubuntu)
- **yum** (RHEL/CentOS 7)
- **dnf** (RHEL/CentOS 8+, Fedora)
- **zypper** (openSUSE)
- **brew** (macOS)
- **choco** (Windows)

### Examples

` + "```yaml" + `
# Install a package
- type: pkg
  name: nginx
  state: present

# Install specific version
- type: pkg
  name: docker-ce
  state: present
  version: "20.10.7"

# Remove a package
- type: pkg
  name: apache2
  state: absent
` + "```" + `

## Service Provider

Manages system services using systemd or init systems.

### Properties

- ` + "`state`" + `: running, stopped, restarted, or reloaded
- ` + "`enabled`" + `: true or false (start on boot)

### Examples

` + "```yaml" + `
# Start and enable a service
- type: service
  name: nginx
  state: running
  enabled: true

# Stop a service
- type: service
  name: apache2
  state: stopped

# Restart a service
- type: service
  name: mysql
  state: restarted
` + "```" + `

## User Provider

Manages users and groups on target systems.

### Properties

- ` + "`state`" + `: present (default) or absent
- ` + "`uid`" + `: User ID
- ` + "`gid`" + `: Primary group ID
- ` + "`home`" + `: Home directory path
- ` + "`shell`" + `: Login shell
- ` + "`groups`" + `: List of additional groups

### Examples

` + "```yaml" + `
# Create a user
- type: user
  name: appuser
  state: present
  uid: 1001
  gid: 1001
  home: /home/appuser
  shell: /bin/bash
  groups:
    - sudo
    - docker

# Remove a user
- type: user
  name: olduser
  state: absent
` + "```" + `

## Shell Provider

Executes shell commands on target systems.

### Properties

- ` + "`command`" + ` (required): Command to execute
- ` + "`creates`" + `: Path that should exist after command runs
- ` + "`unless`" + `: Command that prevents execution if it succeeds
- ` + "`only_if`" + `: Command that must succeed for execution
- ` + "`user`" + `: User to run command as
- ` + "`cwd`" + `: Working directory
- ` + "`timeout`" + `: Timeout in seconds

### Examples

` + "```yaml" + `
# Run a setup script
- type: shell
  name: app-setup
  command: /opt/app/setup.sh
  creates: /var/lib/app/setup.done
  user: app

# Conditional command
- type: shell
  name: backup-db
  command: mysqldump mydb > /backup/mydb.sql
  unless: test -f /backup/mydb.sql
  user: backup

# Command with timeout
- type: shell
  name: long-task
  command: /usr/bin/long-running-task
  timeout: 3600
  user: worker
` + "```" + `

## Provider Development

### Creating Custom Providers

Providers implement the ` + "`Provider`" + ` interface:

` + "```go" + `
type Provider interface {
    Type() string
    Validate(*Resource) error
    Read(context.Context, *Resource) (map[string]interface{}, error)
    Diff(context.Context, *Resource, map[string]interface{}) (*ResourceDiff, error)
    Apply(context.Context, *Resource, *ResourceDiff) error
}
` + "```" + `

### Provider Registration

Register providers with the registry:

` + "```go" + `
registry := types.NewProviderRegistry()
err := registry.Register(providers.NewMyProvider(connection))
` + "```" + `
`

	return os.WriteFile(filepath.Join(outputDir, "providers.md"), []byte(content), 0644)
}

func generateIndex(outputDir string, packages []PackageDoc) error {
	content := `# Chisel Documentation

Welcome to the Chisel documentation! Chisel is a modern, agentless configuration management tool written in Go.

## Documentation Sections

### [Getting Started](getting-started.md)
Quick start guide to get you up and running with Chisel in minutes.

### [User Guide](user-guide.md)
Comprehensive guide covering all aspects of using Chisel for configuration management.

### [Provider Documentation](providers.md)
Detailed documentation for all built-in providers and their properties.

### [API Reference](api-reference.md)
Complete API documentation for developers and advanced users.

## Quick Links

- [Installation](getting-started.md#installation)
- [Core Concepts](getting-started.md#core-concepts)
- [Resource Types](user-guide.md#resource-types)
- [Examples](../examples/)

## Package Overview

` + fmt.Sprintf("Documentation generated for %d packages:", len(packages)) + `

` + func() string {
		var lines []string
		for _, pkg := range packages {
			lines = append(lines, fmt.Sprintf("- **%s**: %s", pkg.Name, pkg.Synopsis))
		}
		sort.Strings(lines)
		return strings.Join(lines, "\n")
	}() + `

## Contributing

See the main [README](../README.md) for contribution guidelines.

## License

MIT License - see [LICENSE](../LICENSE) for details.
`

	return os.WriteFile(filepath.Join(outputDir, "README.md"), []byte(content), 0644)
}
