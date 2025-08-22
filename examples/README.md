# Forge Examples

Simple, practical examples to get you started with Forge configuration management.

## Quick Start Examples

### 1. Simple File (`simple-file/`)
The absolute simplest example - create a file on a server.
- **Perfect for**: First-time users, testing connectivity
- **What it does**: Creates `/tmp/hello.txt` with content
- **Files**: Just one `module.yaml` file

### 2. Basic Web Server (`basic-webserver/`)
Install and configure NGINX web server.
- **Perfect for**: Setting up a simple website
- **What it does**: Installs nginx, creates homepage, starts service
- **Files**: One `webserver.yaml` file

### 3. User Management (`user-management/`)
Create users and manage SSH access.
- **Perfect for**: Team onboarding, server access setup
- **What it does**: Creates users, sets up SSH keys, configures sudo
- **Files**: One `users.yaml` file

## Design Philosophy

These examples follow Forge's "simple by default" approach:

- **One file per example** - No complex directory structures
- **Self-contained** - Everything needed is in one place
- **Copy and edit** - Just change the server details and run
- **Real-world useful** - Solve actual problems you face daily

## 🔧 How to Use

1. **Copy an example**:
   ```bash
   cp -r examples/simple-file my-config
   cd my-config
   ```

2. **Edit the target server**:
   ```bash
   vim module.yaml  # Change host, user, SSH key path
   ```

3. **Apply the configuration** (coming soon):
   ```bash
   forge apply module.yaml
   ```

## 📚 Learning Path

1. Start with `simple-file` to test your setup
2. Try `basic-webserver` to see real infrastructure changes
3. Use `user-management` for team collaboration
4. Create your own modules based on these patterns

## 🆘 Need Help?

- Visit [ataiva.com](https://ataiva.com) for documentation
- Check the [GitHub repository](https://github.com/ataiva-software/forge)
- Look at the main README.md for more details

## Next Steps

Once you're comfortable with these examples:
- Combine multiple resources in one file
- Use variables and templates for reusability
- Set up inventory files for multiple servers
- Explore the plan/apply workflow (coming soon)

Remember: Forge is designed to be simple. If you find yourself creating complex directory structures, you might be overcomplicating it!
