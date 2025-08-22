# Enterprise Compliance Example

This example demonstrates Forge's enterprise-grade compliance and governance features, including:

- **CIS Ubuntu 20.04 compliance** - Automated compliance checking
- **Multi-stage approval workflows** - Production change management
- **Audit logging** - Complete operational audit trails
- **RBAC integration** - Role-based access control
- **Policy as code** - Automated compliance validation

## Features Demonstrated

### 1. Compliance Modules
- **CIS Ubuntu 20.04** - Center for Internet Security benchmarks
- **NIST 800-53** - National Institute of Standards and Technology controls
- **Custom policies** - Organization-specific compliance rules

### 2. Approval Workflows
- **Multi-stage approvals** - Security → Operations → Change Advisory Board
- **Conditional workflows** - Different approval paths based on criticality
- **Timeout management** - Automatic expiration of stale requests
- **Notification integration** - Email and Slack notifications

### 3. Audit & Governance
- **Complete audit trails** - All actions logged with user attribution
- **Policy violations** - Automatic detection and reporting
- **RBAC enforcement** - Role-based access control
- **Secrets management** - Secure credential handling

## Quick Start

### 1. Plan the Configuration
```bash
# Check what changes will be made
forge plan --module module.yaml

# Run compliance checks
forge compliance check --module module.yaml --framework cis-ubuntu-20.04
```

### 2. Submit for Approval (Production)
```bash
# Submit production change for approval
forge apply --module module.yaml --require-approval

# Check approval status
forge approval status <request-id>
```

### 3. Apply After Approval
```bash
# Apply changes after approval
forge apply --module module.yaml --approval-id <request-id>
```

## Configuration Files

### Module Configuration (`module.yaml`)
- CIS-compliant web server setup
- Secure nginx configuration
- Audit daemon configuration
- Proper file permissions and ownership

### Approval Workflow (`approval-workflow.yaml`)
- Multi-stage approval process
- Security and operations reviews
- Change Advisory Board for critical changes
- Notification integration

### Compliance Policy (`compliance-policy.rego`)
- CIS Ubuntu 20.04 controls
- NIST 800-53 requirements
- Custom security policies
- Automated violation detection

## Compliance Controls Implemented

### CIS Ubuntu 20.04
- **CIS-6.1.2**: Ensure permissions on /etc/passwd are configured (0644)
- **CIS-6.1.3**: Ensure permissions on /etc/shadow are configured (0640)
- **CIS-5.4.2**: Ensure system accounts are secured (no interactive shells)
- **CIS-2.2.1**: Ensure X Window System is not installed on servers

### NIST 800-53
- **AU-2**: Audit Events - Audit daemon enabled and configured
- **AU-4**: Audit Storage Capacity - Log rotation configured
- **AC-2**: Account Management - Secure user account configuration

### Custom Security Policies
- Nginx security headers configuration
- Server token hiding
- Secure web content permissions
- Audit log management

## Usage Examples

### Check Compliance
```bash
# Check CIS compliance
forge compliance check --module module.yaml --framework cis-ubuntu-20.04

# Check NIST compliance  
forge compliance check --module module.yaml --framework nist-800-53

# Check all loaded compliance frameworks
forge compliance check --module module.yaml --all
```

### Approval Workflow
```bash
# Create approval workflow
forge approval create-workflow --config approval-workflow.yaml

# Submit change for approval
forge apply --module module.yaml --environment production

# List pending approvals
forge approval list --status pending

# Approve a request
forge approval approve <request-id> --comment "Security review passed"

# Reject a request
forge approval reject <request-id> --comment "Security concerns identified"
```

### Audit and Monitoring
```bash
# View audit logs
forge audit logs --filter "action=apply" --since "24h"

# View policy violations
forge policy violations --module module.yaml

# View user activity
forge audit user-activity --user "admin@company.com" --since "7d"
```

## Integration with Enterprise Systems

### LDAP/Active Directory
```yaml
rbac:
  enabled: true
  ldap:
    server: ldap://company.com
    base_dn: "dc=company,dc=com"
    user_filter: "(&(objectClass=user)(sAMAccountName=%s))"
    group_filter: "(&(objectClass=group)(member=%s))"
```

### Vault Integration
```yaml
secrets:
  enabled: true
  vault:
    address: https://vault.company.com
    auth_method: ldap
    mount_path: secret/
```

### Monitoring Integration
```yaml
monitoring:
  enabled: true
  prometheus:
    endpoint: http://prometheus.company.com:9090
  grafana:
    dashboard_url: https://grafana.company.com/d/forge
```

## Benefits

### For Security Teams
- **Automated compliance** - Continuous compliance checking
- **Policy as code** - Version-controlled security policies
- **Audit trails** - Complete operational visibility
- **Approval workflows** - Controlled change management

### For Operations Teams
- **Reduced manual work** - Automated compliance validation
- **Faster deployments** - Streamlined approval processes
- **Better visibility** - Real-time monitoring and alerting
- **Risk reduction** - Automated policy enforcement

### For Compliance Officers
- **Continuous monitoring** - Real-time compliance status
- **Audit readiness** - Complete audit trails
- **Policy enforcement** - Automated violation detection
- **Reporting** - Compliance dashboards and reports

This example showcases how Forge enables enterprise-grade infrastructure management with built-in compliance, governance, and security features.
