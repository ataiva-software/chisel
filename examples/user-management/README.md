# User Management Example

Create users and manage SSH access.

## What it does
- Creates a new user account
- Sets up SSH key access
- Configures sudo permissions

## Usage
```bash
# 1. Copy this example
cp -r examples/user-management setup-users
cd setup-users

# 2. Edit the configuration
vim users.yaml  # Add your servers and user details

# 3. Apply (coming soon)
chisel apply users.yaml
```

Perfect for setting up team access to servers!
