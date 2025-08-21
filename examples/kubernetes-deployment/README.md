# Kubernetes Deployment Example

This example demonstrates Chisel's Kubernetes provider capabilities for deploying a complete web application stack to Kubernetes.

## Features Demonstrated

### Kubernetes Resources
- **Namespace** - Isolated environment for the application
- **ConfigMap** - Application configuration and nginx config
- **Secret** - Secure storage for database credentials and API keys
- **Deployment** - Scalable web application deployment
- **Service** - Internal load balancing and service discovery
- **Ingress** - External access with TLS termination
- **HorizontalPodAutoscaler** - Automatic scaling based on CPU/memory
- **PersistentVolumeClaim** - Persistent storage for application data

### Best Practices
- **Resource limits and requests** - Proper resource management
- **Health checks** - Liveness and readiness probes
- **Configuration management** - Externalized configuration
- **Secret management** - Secure credential handling
- **Auto-scaling** - Horizontal pod autoscaling
- **TLS termination** - Secure HTTPS access
- **Persistent storage** - Data persistence across pod restarts

## Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured to access the cluster
- Ingress controller (nginx-ingress recommended)
- cert-manager for TLS certificates (optional)

## Quick Start

### 1. Plan the Deployment
```bash
# Check what Kubernetes resources will be created
chisel plan --module module.yaml

# Validate Kubernetes manifests
chisel validate --module module.yaml --provider kubernetes
```

### 2. Apply the Configuration
```bash
# Apply to Kubernetes cluster
chisel apply --module module.yaml

# Check deployment status
kubectl get all -n web-app
```

### 3. Verify the Deployment
```bash
# Check pods are running
kubectl get pods -n web-app

# Check service endpoints
kubectl get svc -n web-app

# Check ingress
kubectl get ingress -n web-app

# View application logs
kubectl logs -n web-app -l app=web-app
```

## Configuration Details

### Namespace Isolation
```yaml
- type: kubernetes
  name: web-app-namespace
  properties:
    kind: Namespace
    namespace: web-app
```

### Configuration Management
```yaml
- type: kubernetes
  name: web-app-config
  properties:
    kind: ConfigMap
    data:
      app.properties: |
        server.port=8080
        logging.level.root=INFO
      nginx.conf: |
        server {
            listen 80;
            location / {
                proxy_pass http://localhost:8080;
            }
        }
```

### Secret Management
```yaml
- type: kubernetes
  name: web-app-secrets
  properties:
    kind: Secret
    type: Opaque
    data:
      database-url: <base64-encoded-value>
      api-key: <base64-encoded-value>
```

### Scalable Deployment
```yaml
- type: kubernetes
  name: web-app-deployment
  properties:
    kind: Deployment
    spec:
      replicas: 3
      template:
        spec:
          containers:
            - name: web-app
              image: nginx:1.21-alpine
              resources:
                requests:
                  memory: "128Mi"
                  cpu: "100m"
                limits:
                  memory: "256Mi"
                  cpu: "200m"
```

### Auto-scaling Configuration
```yaml
- type: kubernetes
  name: web-app-hpa
  properties:
    kind: HorizontalPodAutoscaler
    spec:
      minReplicas: 3
      maxReplicas: 10
      metrics:
        - type: Resource
          resource:
            name: cpu
            target:
              averageUtilization: 70
```

## Advanced Features

### Rolling Updates
```bash
# Update the application image
chisel apply --module module.yaml --var image_tag=v1.1.0

# Monitor rollout status
kubectl rollout status deployment/web-app -n web-app

# Rollback if needed
kubectl rollout undo deployment/web-app -n web-app
```

### Scaling Operations
```bash
# Manual scaling
kubectl scale deployment web-app --replicas=5 -n web-app

# Check HPA status
kubectl get hpa -n web-app

# View HPA events
kubectl describe hpa web-app-hpa -n web-app
```

### Monitoring and Debugging
```bash
# View resource usage
kubectl top pods -n web-app

# Check events
kubectl get events -n web-app --sort-by='.lastTimestamp'

# Debug pod issues
kubectl describe pod <pod-name> -n web-app

# Access pod shell
kubectl exec -it <pod-name> -n web-app -- /bin/sh
```

## Integration with Chisel Features

### Compliance Checking
```bash
# Check Kubernetes security policies
chisel compliance check --module module.yaml --framework k8s-security

# Validate resource configurations
chisel policy validate --module module.yaml --policy k8s-best-practices
```

### Drift Detection
```bash
# Enable drift detection for Kubernetes resources
chisel drift enable --module module.yaml --interval 5m

# Check for configuration drift
chisel drift check --module module.yaml
```

### Approval Workflows
```bash
# Production deployments require approval
chisel apply --module module.yaml --environment production --require-approval

# Check approval status
chisel approval status <request-id>
```

## Troubleshooting

### Common Issues

1. **ImagePullBackOff**
   ```bash
   kubectl describe pod <pod-name> -n web-app
   # Check image name and registry access
   ```

2. **CrashLoopBackOff**
   ```bash
   kubectl logs <pod-name> -n web-app --previous
   # Check application logs and health checks
   ```

3. **Service Not Accessible**
   ```bash
   kubectl get svc -n web-app
   kubectl get endpoints -n web-app
   # Verify service selector matches pod labels
   ```

4. **Ingress Issues**
   ```bash
   kubectl describe ingress web-app-ingress -n web-app
   # Check ingress controller and DNS configuration
   ```

### Resource Cleanup
```bash
# Remove all resources
chisel destroy --module module.yaml

# Or manually clean up
kubectl delete namespace web-app
```

## Benefits

### For Development Teams
- **Simplified deployment** - Declarative Kubernetes configuration
- **Version control** - Infrastructure as code
- **Consistent environments** - Same configuration across dev/staging/prod
- **Easy rollbacks** - Built-in rollback capabilities

### For Operations Teams
- **Automated scaling** - HPA based on metrics
- **Health monitoring** - Built-in health checks
- **Resource management** - Proper resource limits and requests
- **Security** - Secret management and network policies

### For Platform Teams
- **Standardization** - Consistent deployment patterns
- **Compliance** - Built-in policy checking
- **Observability** - Integrated monitoring and logging
- **Governance** - Approval workflows for production changes

This example showcases how Chisel enables modern Kubernetes deployments with enterprise-grade features like compliance checking, approval workflows, and drift detection.
