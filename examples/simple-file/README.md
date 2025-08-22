# Simple File Management Example

The simplest possible Forge example - just create a file on a server.

## What it does
- Creates `/tmp/hello.txt` with some content
- Sets proper permissions (644)
- Owned by root

## Files
- `module.yaml` - Single file with everything needed

## Usage
```bash
# 1. Copy this example
cp -r examples/simple-file my-config
cd my-config

# 2. Edit the target server
vim module.yaml  # Change the host IP/hostname

# 3. Run it (coming soon)
forge apply module.yaml
```

This is the absolute minimum Forge configuration - perfect for getting started!
