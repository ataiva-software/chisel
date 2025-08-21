# User Management Example

This example demonstrates all of Chisel's core providers working together to set up a complete development user environment.

## What This Example Does

### User Provider
- Creates a user `devuser` with specific UID/GID
- Assigns the user to multiple groups (sudo, docker, developers)
- Sets up home directory and shell

### File Provider
- Creates user's home directory structure
- Sets up configuration files (.bashrc, scripts)
- Creates SSH directory with proper permissions
- Manages file ownership and permissions

### Package Provider
- Installs essential development tools (git, vim, curl, wget)
- Ensures packages are present and up-to-date

### Service Provider
- Ensures SSH service is running for remote access
- Enables SSH service to start on boot

### Shell Provider
- Creates workspace directories using mkdir
- Configures Git with user settings
- Sets up logging and monitoring scripts
- Uses conditional execution (creates, user context)

## Running the Example

```bash
# Navigate to the example directory
cd examples/user-management

# Create a plan to see what will change
../../bin/chisel plan --module module.yaml

# Apply the configuration (dry run first)
../../bin/chisel apply --module module.yaml --dry-run

# Apply for real
../../bin/chisel apply --module module.yaml
```

## Expected Output

The plan should show:
- 1 user to create
- 8 files to create/manage
- 4 packages to install
- 1 service to configure
- 3 shell commands to execute

## Resource Dependencies

This example demonstrates implicit dependencies:
1. User must be created before files can be owned by them
2. Directories must exist before files can be created in them
3. Packages should be installed before services are configured
4. Shell commands can depend on files and users existing

## Security Considerations

- SSH directory has restrictive permissions (0700)
- User scripts are executable only by the owner
- Log files are owned by root for security
- Groups provide appropriate access levels

## Customization

You can modify this example by:
- Changing the username and UID/GID
- Adding more packages or removing unnecessary ones
- Modifying file contents and permissions
- Adding more shell commands for setup tasks
- Adjusting service configurations

## Verification

After applying, you can verify the setup:

```bash
# Check user was created
id devuser

# Check user's groups
groups devuser

# Check home directory structure
ls -la /home/devuser/

# Check installed packages
dpkg -l | grep -E "(git|vim|curl|wget)"

# Check SSH service
systemctl status ssh

# Test user's script
sudo -u devuser /home/devuser/scripts/hello.sh
```

This example showcases Chisel's ability to manage complex, multi-resource configurations with proper ordering and dependencies.
