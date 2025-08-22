# Basic Web Server Example

Install and configure NGINX in one simple file.

## What it does
- Installs nginx package
- Creates a custom index.html
- Starts and enables nginx service

## Usage
```bash
# 1. Copy this example
cp -r examples/basic-webserver my-webserver
cd my-webserver

# 2. Edit your server details
vim webserver.yaml  # Change host, user, etc.

# 3. Deploy (coming soon)
forge apply webserver.yaml
```

After running, visit http://your-server-ip to see your site!
