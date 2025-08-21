# Chisel

**Carving infrastructure into shape**

[![Go Report Card](https://goreportcard.com/badge/github.com/ataiva-software/chisel)](https://goreportcard.com/report/github.com/ataiva-software/chisel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI/CD](https://github.com/ataiva-software/chisel/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/ataiva-software/chisel/actions/workflows/ci-cd.yml)

> **ALPHA SOFTWARE WARNING**
>
> Chisel is currently in **alpha stage**. While the architecture is solid and the CLI is functional, many providers use mock implementations for testing. This is perfect for:
>
> - **Testing the CLI and workflow**
> - **Evaluating the architecture**
> - **Providing feedback on the user experience**
>
> **Not yet ready for production use.** See [Implementation Status](#implementation-status) for details.

Chisel is a modern, agentless configuration management and infrastructure orchestration tool written in Go. It combines the best features of Terraform's plan/apply workflow, Ansible's agentless approach, and Puppet's resource model into a fast, typed, and secure platform.

## Quick Start

### Installation

#### Download Pre-built Binaries

```bash
# Linux (x86_64)
curl -L https://github.com/ataiva-software/chisel/releases/latest/download/chisel-linux-amd64 -o chisel
chmod +x chisel
sudo mv chisel /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/ataiva-software/chisel/releases/latest/download/chisel-darwin-amd64 -o chisel
chmod +x chisel
sudo mv chisel /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/ataiva-software/chisel/releases/latest/download/chisel-darwin-arm64 -o chisel
chmod +x chisel
sudo mv chisel /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/ataiva-software/chisel/releases/latest/download/chisel-windows-amd64.exe" -OutFile "chisel.exe"
```

#### Build from Source

```bash
git clone https://github.com/ataiva-software/chisel.git
cd chisel
make build
sudo cp bin/chisel /usr/local/bin/
```

### Try It Out

```bash
# Initialize a new project
chisel init my-infrastructure
cd my-infrastructure

# Create a plan to see what will change
chisel plan --module modules/webserver.yaml --inventory inventory/hosts.yaml

# Apply changes (dry run first)
chisel apply --module modules/webserver.yaml --inventory inventory/hosts.yaml --dry-run

# Apply for real (with confirmation)
chisel apply --module modules/webserver.yaml --inventory inventory/hosts.yaml
```

## CLI Commands

### Available Commands

```bash
# Initialize a new project
chisel init <project-name>

# Create an execution plan
chisel plan --module <module.yaml> [--inventory <inventory.yaml>]

# Apply changes to infrastructure
chisel apply --module <module.yaml> [--inventory <inventory.yaml>] [--dry-run] [--auto-approve]

# Get help
chisel --help
chisel <command> --help
```

### Examples

```bash
# Plan changes for a module
chisel plan --module examples/simple-file/module.yaml

# Apply changes with confirmation
chisel apply --module examples/simple-file/module.yaml

# Dry run (show what would happen)
chisel apply --module examples/simple-file/module.yaml --dry-run

# Auto-approve (skip confirmation)
chisel apply --module examples/simple-file/module.yaml --auto-approve

# Use with inventory
chisel plan --module webserver.yaml --inventory hosts.yaml
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Controller    │    │    Inventory    │    │     Targets     │
│                 │    │                 │    │                 │
│ • Compiler      │◄──►│ • Static YAML   │    │ • Linux/macOS   │
│ • Planner       │    │ • Cloud APIs    │    │ • Windows       │
│ • Executor      │    │ • Labels/Facts  │    │ • Containers    │
│ • Policy Engine │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                                              ▲
         │              SSH/WinRM/Cloud APIs            │
         └──────────────────────────────────────────────┘
```

## Core Concepts

### Resources

Typed units of infrastructure state (packages, files, services, users, etc.)

```yaml
apiVersion: ataiva.com/chisel/v1
kind: Module
metadata:
  name: webserver
  version: 1.0.0
spec:
  resources:
    - type: pkg
      name: nginx
      state: present
    - type: file
      name: nginx-config
      state: present
      path: /etc/nginx/nginx.conf
      content: |
        server {
          listen 80;
          root /var/www/html;
        }
      mode: "0644"
    - type: service
      name: nginx
      state: running
      enabled: true
```

### Inventory

Dynamic and static target discovery

```yaml
# inventory/hosts.yaml
targets:
  webservers:
    selector: "role=web,env=prod"
    connection:
      type: ssh
      user: ubuntu
      key: ~/.ssh/id_rsa
  databases:
    hosts:
      - db1.example.com
      - db2.example.com
    connection:
      type: ssh
      user: admin
      key: ~/.ssh/id_rsa
```

### Plans

Deterministic change previews

```bash
$ chisel plan --inventory inventory/hosts.yaml --module webserver.yaml
Plan: 3 to add, 1 to change, 0 to destroy

+ pkg.nginx
  (will be created)

~ file.nginx-config
  (will be updated)
  path: /etc/nginx/nginx.conf

+ service.nginx
  (will be created)
  state: stopped → running
  enabled: false → true
```

## Implementation Status

### Phase 1: Core Configuration Management - COMPLETE

- [x] **Basic CLI structure** - Full command-line interface
- [x] **Resource type system** - Strongly typed resource definitions
- [x] **SSH connection management** - Secure remote execution
- [x] **Core providers** - All 5 providers implemented and tested
  - [x] **File Provider** - Files, directories, templates, permissions
  - [x] **Package Provider** - Cross-platform package management
  - [x] **Service Provider** - System service lifecycle management
  - [x] **User Provider** - User and group management
  - [x] **Shell Provider** - Command execution with guardrails
- [x] **Module system and registry** - YAML-based configuration
- [x] **Plan/apply workflow** - Terraform-style preview and execution
- [x] **Static inventory support** - Host and group management
- [x] **Basic templating** - Go template engine integration

### Phase 2: Orchestration & Workflow - COMPLETE

- [x] **Dynamic inventory** - Pluggable inventory providers with AWS support
- [x] **Parallel execution engine** - Dependency-aware concurrent execution
- [x] **Dependency resolution** - Automatic dependency graph creation
- [x] **Error handling and rollback** - Automatic failure recovery with retry logic
- [x] **Real SSH integration** - Production-ready SSH connection management
- [x] **WinRM integration** - Windows remote management support
- [x] **Drift detection scheduling** - Continuous monitoring with configurable intervals
- [x] **Event system and notifications** - Real-time status updates with multiple channels
- [x] **Web UI dashboard** - Visual management interface

### Phase 3: Policy & Compliance - COMPLETE

- [x] **Policy engine** - Built-in policy evaluation with Rego-style rules
- [x] **Audit logging and trails** - Comprehensive audit logging with rotation
- [x] **RBAC and multi-tenancy** - Role-based access control with user management
- [x] **Secrets management integration** - Vault, AWS Secrets Manager integration
- [x] **Compliance modules** (CIS, NIST, STIG) - Pre-built compliance policies
- [x] **Approval workflows** - Multi-stage approval processes

### Phase 4: Advanced Features - MAJOR PROGRESS

- [x] **Kubernetes provider** - Container orchestration support
- [x] **Cloud provider integrations** - Azure VM discovery
- [x] **Monitoring and observability** - Prometheus metrics and monitoring
- [ ] **WASM provider SDK** - WebAssembly provider extensions
- [ ] **Supply chain security** (cosign, SLSA) - Signed modules and provenance
- [ ] **High availability controller** - Distributed controller architecture
- [ ] **Enterprise features** (SAML, SCIM) - Enterprise authentication

### Phase 5: Ecosystem Integration - PLANNED

- [ ] **Terraform integration**
- [ ] **GitOps workflows**
- [ ] **CI/CD pipeline integration**
- [ ] **Monitoring and observability**
- [ ] **SaaS offering** (Chisel Cloud)

## Development

### Prerequisites

- Go 1.21+
- Make
- Docker (for testing)

### Building

```bash
make build
```

### Testing

```bash
make test
make test-integration
```

### Contributing

We use Test-Driven Development (TDD):

1. Write failing tests first
2. Implement minimal code to pass
3. Refactor and improve
4. Repeat

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## Providers

### Core Providers - IMPLEMENTED

- **file**: File and directory management with templating
- **pkg**: Package management (apt, yum, dnf, zypper, brew, choco)
- **service**: System service management (systemd, init)
- **user**: User and group management with full configuration
- **shell**: Command execution with conditional logic and guardrails

### Cloud Providers - PLANNED

- **aws**: EC2, SSM, tags, metadata
- **azure**: VM metadata, tags
- **gcp**: Compute metadata, tags

### Planned Providers - FUTURE

- **kubernetes**: Deployments, ConfigMaps, Secrets
- **database**: PostgreSQL, MySQL user/schema management
- **network**: Routes, firewall rules, DNS

## Examples

### Simple File Management

```yaml
apiVersion: ataiva.com/chisel/v1
kind: Module
metadata:
  name: simple-file
spec:
  resources:
    - type: file
      name: hello-file
      state: present
      path: /tmp/hello.txt
      content: "Hello from Chisel!"
      mode: "0644"
```

### Complete User Environment

```yaml
apiVersion: ataiva.com/chisel/v1
kind: Module
metadata:
  name: user-management
spec:
  resources:
    # Install development tools
    - type: pkg
      name: git
      state: present
    
    # Create user
    - type: user
      name: devuser
      state: present
    
    # Create home directory
    - type: file
      name: devuser-home
      state: present
      path: /home/devuser
      file_type: directory
      mode: "0755"
      owner: devuser
    
    # Configure shell
    - type: file
      name: devuser-bashrc
      state: present
      path: /home/devuser/.bashrc
      content: |
        alias ll='ls -alF'
        export EDITOR=vim
      mode: "0644"
      owner: devuser
    
    # Setup workspace
    - type: shell
      name: create-workspace
      command: mkdir -p /home/devuser/workspace
      creates: /home/devuser/workspace
```

## Performance

- **Concurrent execution**: Parallel SSH connections with dependency resolution
- **Efficient planning**: Incremental graph computation
- **Minimal footprint**: Single binary, no runtime dependencies
- **Fast startup**: Sub-second initialization
- **Smart caching**: Resource state caching for performance

## Security

- **mTLS**: All communication encrypted with mutual TLS
- **RBAC**: Role-based access control *(planned)*
- **Audit**: Immutable audit trails *(planned)*
- **Secrets**: Integration with Vault, AWS Secrets Manager, etc. *(planned)*
- **Supply Chain**: Signed modules and provenance tracking *(planned)*

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Community

- **Website**: [ataiva.com](https://ataiva.com)
- **GitHub**: [github.com/ataiva-software/chisel](https://github.com/ataiva-software/chisel)
- **Discussions**: [GitHub Discussions](https://github.com/ataiva-software/chisel/discussions)
- **Issues**: [GitHub Issues](https://github.com/ataiva-software/chisel/issues)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Contribution Guide

1. **Check the [roadmap](#implementation-status)** - See what's planned
2. **Create an issue** - Discuss your idea first
3. **Follow TDD** - Write tests first, then implement
4. **Submit a PR** - Include tests and documentation

### Priority Areas

- **Real provider implementations** - Replace mock implementations with actual SSH/WinRM operations
- **Error handling** - Robust error handling and recovery
- **Cloud integrations** - Additional cloud providers (GCP, DigitalOcean, etc.)
- **Documentation** - More examples and tutorials

## Releases

### Alpha Releases

Chisel follows semantic versioning. Alpha releases are available for testing:

- **v0.1.0-alpha** - Initial alpha release with core functionality
- Check [releases page](https://github.com/ataiva-software/chisel/releases) for latest versions

### Release Schedule

- **Alpha** (Current) - Core functionality, mock implementations
- **Beta** (Q1 2024) - Real provider implementations, production testing
- **v1.0** (Q2 2024) - Production ready, full feature set

## Acknowledgments

Chisel learns from and builds upon the excellent work of:

- Terraform (HashiCorp)
- Ansible (Red Hat)
- Puppet (Puppet Inc.)
- Chef (Progress Software)
- SaltStack (VMware)

---

*"The best way to predict the future is to invent it."* - Alan Kay
