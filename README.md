# Chisel

**Carving infrastructure into shape**

[![Go Report Card](https://goreportcard.com/badge/github.com/ao/chisel)](https://goreportcard.com/report/github.com/ao/chisel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Chisel is a modern, agentless configuration management and infrastructure orchestration tool written in Go. It combines the best features of Terraform's plan/apply workflow, Ansible's agentless approach, and Puppet's resource model into a fast, typed, and secure platform.

## Why Chisel?

**Pain Points We Solve:**

- **Ansible**: Slow execution, weak typing, Python dependency hell, no plan/preview
- **Puppet/Chef**: Heavy agent overhead, complex setup, Ruby/DSL learning curve
- **Terraform**: Not designed for OS-level configuration, limited drift detection
- **Salt**: Complex architecture, steep learning curve, inconsistent execution models

**Chisel's Approach:**

- ✅ **Agentless by default** - SSH/WinRM execution, no persistent footprint
- ✅ **Plan before apply** - See exactly what will change before execution
- ✅ **Strongly typed** - Catch errors at compile time, not runtime
- ✅ **Fast & concurrent** - Go's concurrency model for parallel execution
- ✅ **Drift detection** - Continuous monitoring without agents
- ✅ **Policy-driven** - Built-in compliance and governance
- ✅ **Supply chain security** - Signed modules, provenance tracking

## Quick Start

```bash
# Install Chisel
go install github.com/ao/chisel/cmd/chisel@latest

# Initialize a new project
chisel init my-infrastructure

# Add a module
chisel module add std/nginx@1.0.0

# Plan changes
chisel plan --inventory hosts.yaml --module nginx

# Apply changes
chisel apply plan.json
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
apiVersion: chisel.dev/v1
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
      path: /etc/nginx/nginx.conf
      content: |
        server {
          listen 80;
          root /var/www/html;
        }
      mode: "0644"
      notify:
        - restart: nginx
    - type: service
      name: nginx
      state: running
      enabled: true
```

### Inventory
Dynamic and static target discovery

```yaml
# inventory.yaml
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
```

### Plans
Deterministic change previews

```bash
$ chisel plan --inventory inventory.yaml --module webserver
Plan: 3 to add, 1 to change, 0 to destroy

+ pkg.nginx
  state: absent → present

~ file./etc/nginx/nginx.conf
  content: (differs)
  mode: 0644

+ service.nginx
  state: stopped → running
  enabled: false → true
```

## Roadmap

### Phase 1: Core Configuration Management ✅
- [x] Basic CLI structure
- [x] Resource type system
- [x] SSH connection management
- [x] Core providers (file provider implemented)
- [ ] Module system and registry
- [ ] Plan/apply workflow
- [ ] Static inventory support
- [ ] Basic templating

### Phase 2: Orchestration & Workflow
- [ ] Dynamic inventory (cloud APIs)
- [ ] Parallel execution engine
- [ ] Dependency resolution
- [ ] Error handling and rollback
- [ ] Drift detection scheduling
- [ ] Event system and notifications
- [ ] Web UI dashboard

### Phase 3: Policy & Compliance
- [ ] Policy engine (OPA/CEL integration)
- [ ] Compliance modules (CIS, NIST, STIG)
- [ ] Audit logging and trails
- [ ] RBAC and multi-tenancy
- [ ] Approval workflows
- [ ] Secrets management integration

### Phase 4: Advanced Features
- [ ] Kubernetes provider
- [ ] Cloud provider integrations
- [ ] WASM provider SDK
- [ ] Supply chain security (cosign, SLSA)
- [ ] High availability controller
- [ ] Enterprise features (SAML, SCIM)

### Phase 5: Ecosystem Integration
- [ ] Terraform integration
- [ ] GitOps workflows
- [ ] CI/CD pipeline integration
- [ ] Monitoring and observability
- [ ] SaaS offering (Chisel Cloud)

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

### Core Providers
- **pkg**: Package management (apt, yum, dnf, zypper, brew, choco)
- **file**: File and directory management
- **service**: System service management
- **user**: User and group management
- **template**: Configuration templating
- **shell**: Command execution (with guardrails)

### Cloud Providers
- **aws**: EC2, SSM, tags, metadata
- **azure**: VM metadata, tags
- **gcp**: Compute metadata, tags

### Planned Providers
- **kubernetes**: Deployments, ConfigMaps, Secrets
- **database**: PostgreSQL, MySQL user/schema management
- **network**: Routes, firewall rules, DNS

## Security

- **mTLS**: All communication encrypted with mutual TLS
- **RBAC**: Role-based access control
- **Audit**: Immutable audit trails
- **Secrets**: Integration with Vault, AWS Secrets Manager, etc.
- **Supply Chain**: Signed modules and provenance tracking

## Performance

- **Concurrent execution**: Parallel SSH connections
- **Efficient planning**: Incremental graph computation
- **Minimal footprint**: Single binary, no runtime dependencies
- **Fast startup**: Sub-second initialization

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Community

- **GitHub**: [github.com/ao/chisel](https://github.com/ao/chisel)
- **Discussions**: [GitHub Discussions](https://github.com/ao/chisel/discussions)
- **Issues**: [GitHub Issues](https://github.com/ao/chisel/issues)

## Acknowledgments

Chisel learns from and builds upon the excellent work of:
- Terraform (HashiCorp)
- Ansible (Red Hat)
- Puppet (Puppet Inc.)
- Chef (Progress Software)
- SaltStack (VMware)

---

*"The best way to predict the future is to invent it."* - Alan Kay
