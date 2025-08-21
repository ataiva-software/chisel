# User Guide

## Module Structure

### Basic Module

```yaml
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
```

### Resource Properties

Each resource type supports different properties. Common properties:

- `state`: Desired state (present, absent, running, stopped)
- `name`: Unique identifier within the module

## Resource Types

### File Resources

Manage files and directories:

```yaml
- type: file
  name: my-file
  state: present
  path: /path/to/file
  content: "File content"
  mode: "0644"
  owner: user
  group: group
```

### Package Resources

Install and manage packages:

```yaml
- type: pkg
  name: nginx
  state: present
```

### Service Resources

Manage system services:

```yaml
- type: service
  name: nginx
  state: running
  enabled: true
```

### User Resources

Manage users and groups:

```yaml
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
```

### Shell Resources

Execute commands:

```yaml
- type: shell
  name: setup-script
  command: /path/to/setup.sh
  creates: /path/to/marker/file
  user: root
```

## Inventory Management

### Static Inventory

```yaml
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
```

### Using Inventory

```bash
chisel plan --module module.yaml --inventory inventory.yaml
chisel apply --module module.yaml --inventory inventory.yaml
```

## Advanced Features

### Templating

Use Go templates in file content:

```yaml
- type: file
  name: templated-config
  state: present
  path: /etc/app/config.yml
  content: |
    server_name: {{ .hostname }}
    environment: {{ .env | default "production" }}
    debug: {{ if eq .env "development" }}true{{ else }}false{{ end }}
```

### Conditional Execution

Shell resources support conditional execution:

```yaml
- type: shell
  name: conditional-command
  command: echo "Running setup"
  creates: /var/lib/app/setup.done
  unless: test -f /var/lib/app/skip-setup
```

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

- Use `creates` and `unless` conditions to avoid unnecessary work
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

```bash
chisel plan --module module.yaml --verbose
chisel apply --module module.yaml --verbose
```

### Dry Run

Test changes without applying:

```bash
chisel apply --module module.yaml --dry-run
```
