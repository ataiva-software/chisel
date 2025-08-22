# Getting Started with Forge

## Installation

### From Source

```bash
git clone https://github.com/ataiva-software/forge.git
cd forge
make build
```

The binary will be available at `./bin/forge`.

### Verify Installation

```bash
./bin/forge --version
```

## Quick Start

### 1. Create Your First Module

Create a simple file management module:

```yaml
# hello.yaml
apiVersion: ataiva.com/forge/v1
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
      content: "Hello from Forge!"
      mode: "0644"
```

### 2. Plan Your Changes

See what Forge will do before applying:

```bash
./bin/forge plan --module hello.yaml
```

### 3. Apply Your Configuration

Apply the changes:

```bash
./bin/forge apply --module hello.yaml
```

### 4. Verify the Result

```bash
cat /tmp/hello.txt
```

## Core Concepts

### Resources

Resources are the fundamental units in Forge. Each resource has:
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
