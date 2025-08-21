# CIS Ubuntu 20.04 Compliance Policy
package chisel.compliance.cis_ubuntu_2004

import future.keywords.if
import future.keywords.in

# CIS 6.1.2 - Ensure permissions on /etc/passwd are configured
deny[msg] {
    input.resource.type == "file"
    input.resource.properties.path == "/etc/passwd"
    input.resource.properties.mode != "0644"
    msg := sprintf("CIS-6.1.2: /etc/passwd must have 0644 permissions, found %s", [input.resource.properties.mode])
}

# CIS 6.1.3 - Ensure permissions on /etc/shadow are configured
deny[msg] {
    input.resource.type == "file"
    input.resource.properties.path == "/etc/shadow"
    not input.resource.properties.mode in ["0640", "0000"]
    msg := sprintf("CIS-6.1.3: /etc/shadow must have 0640 or 0000 permissions, found %s", [input.resource.properties.mode])
}

# CIS 5.4.2 - Ensure system accounts are secured
deny[msg] {
    input.resource.type == "user"
    input.resource.name == "root"
    input.resource.properties.shell in ["/bin/bash", "/bin/sh", "/bin/zsh"]
    msg := "CIS-5.4.2: Root user should not have an interactive shell"
}

# CIS 2.2.1 - Ensure X Window System is not installed (for servers)
deny[msg] {
    input.resource.type == "pkg"
    input.resource.name in ["xserver-xorg", "xorg", "x11-common"]
    input.resource.state == "present"
    input.module.metadata.labels.server_type == "headless"
    msg := sprintf("CIS-2.2.1: X Window System package %s should not be installed on headless servers", [input.resource.name])
}

# Custom rule: Ensure nginx is configured securely
deny[msg] {
    input.resource.type == "file"
    input.resource.properties.path == "/etc/nginx/sites-available/default"
    not contains(input.resource.properties.content, "server_tokens off")
    msg := "Security: nginx configuration must include 'server_tokens off' to hide version information"
}

# Custom rule: Ensure audit daemon is enabled
deny[msg] {
    input.resource.type == "service"
    input.resource.name == "auditd"
    input.resource.properties.enabled != true
    msg := "NIST AU-2: Audit daemon (auditd) must be enabled for compliance"
}

# Custom rule: Ensure secure file permissions for web content
deny[msg] {
    input.resource.type == "file"
    startswith(input.resource.properties.path, "/var/www/")
    input.resource.properties.mode == "0777"
    msg := sprintf("Security: Web content at %s has overly permissive permissions (0777)", [input.resource.properties.path])
}

# Helper function to check if string contains substring
contains(string, substr) if {
    indexof(string, substr) != -1
}
