# test-project

This is a Chisel configuration management project.

## Getting Started

1. Edit the inventory file to define your target hosts:
   `inventory/hosts.yaml`

2. Customize the example module or create new ones:
   `modules/webserver.yaml`

3. Coming soon - Plan and apply workflow:
   `chisel plan --inventory inventory/hosts.yaml --module webserver`
   `chisel apply plan.json`

## Current Status

Chisel is in active development. Currently implemented:
- ✅ Project initialization (`chisel init`)
- ✅ Resource type system with file provider
- ✅ SSH connection management
- 🚧 Plan/apply workflow (coming soon)
- 🚧 Module loading and execution (coming soon)

## Project Structure

- `chisel.yaml` - Project configuration
- `inventory/` - Target host definitions
- `modules/` - Reusable configuration modules
- `templates/` - Configuration templates
- `plans/` - Generated execution plans (coming soon)

## Documentation

Visit https://github.com/ataiva-software/chisel for more information.
