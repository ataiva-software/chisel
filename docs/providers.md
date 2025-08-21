# Provider Documentation

## File Provider

Manages files and directories on target systems.

### Properties

- `path` (required): Path to the file or directory
- `state`: present (default) or absent
- `content`: File content (for files)
- `source`: Source file to copy
- `template`: Template file to render
- `mode`: File permissions (e.g., "0644")
- `owner`: File owner
- `group`: File group
- `file_type`: file (default) or directory

### Examples

```yaml
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
```

## Package Provider

Manages system packages using native package managers.

### Properties

- `state`: present (default) or absent
- `version`: Specific version to install (optional)

### Supported Package Managers

- **apt** (Debian/Ubuntu)
- **yum** (RHEL/CentOS 7)
- **dnf** (RHEL/CentOS 8+, Fedora)
- **zypper** (openSUSE)
- **brew** (macOS)
- **choco** (Windows)

### Examples

```yaml
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
```

## Service Provider

Manages system services using systemd or init systems.

### Properties

- `state`: running, stopped, restarted, or reloaded
- `enabled`: true or false (start on boot)

### Examples

```yaml
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
```

## User Provider

Manages users and groups on target systems.

### Properties

- `state`: present (default) or absent
- `uid`: User ID
- `gid`: Primary group ID
- `home`: Home directory path
- `shell`: Login shell
- `groups`: List of additional groups

### Examples

```yaml
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
```

## Shell Provider

Executes shell commands on target systems.

### Properties

- `command` (required): Command to execute
- `creates`: Path that should exist after command runs
- `unless`: Command that prevents execution if it succeeds
- `only_if`: Command that must succeed for execution
- `user`: User to run command as
- `cwd`: Working directory
- `timeout`: Timeout in seconds

### Examples

```yaml
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
```

## Provider Development

### Creating Custom Providers

Providers implement the `Provider` interface:

```go
type Provider interface {
    Type() string
    Validate(*Resource) error
    Read(context.Context, *Resource) (map[string]interface{}, error)
    Diff(context.Context, *Resource, map[string]interface{}) (*ResourceDiff, error)
    Apply(context.Context, *Resource, *ResourceDiff) error
}
```

### Provider Registration

Register providers with the registry:

```go
registry := types.NewProviderRegistry()
err := registry.Register(providers.NewMyProvider(connection))
```
