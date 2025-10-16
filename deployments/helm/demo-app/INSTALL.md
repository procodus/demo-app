# Quick Installation Guide

## Prerequisites

Before installing the demo-app Helm chart, ensure you have:

1. **Kubernetes cluster** (1.21+)
2. **Helm** (3.8+)
3. **kubectl** configured to access your cluster
4. **RabbitMQ** (in-cluster or external)
5. **PostgreSQL** (in-cluster or external)

Optional:
- **Prometheus Operator** (for ServiceMonitor support)

## Quick Start

### 1. Install with Default Values

```bash
helm install my-demo-app ./deployments/helm/demo-app
```

This deploys all three services (generator, backend, frontend) with default settings.

### 2. Verify Installation

```bash
# Check pods
kubectl get pods -l app.kubernetes.io/instance=my-demo-app

# Check services
kubectl get svc -l app.kubernetes.io/instance=my-demo-app

# Check ServiceMonitors (if Prometheus Operator is installed)
kubectl get servicemonitor -l app.kubernetes.io/instance=my-demo-app
```

### 3. Access Frontend

```bash
# Port-forward to frontend
kubectl port-forward svc/my-demo-app-frontend 8080:8080

# Open http://localhost:8080 in browser
```

## Configuration Examples

### Use External RabbitMQ and PostgreSQL

Create `custom-values.yaml`:

```yaml
rabbitmq:
  host: my-rabbitmq.example.com
  user: myuser
  password: mypassword

postgresql:
  host: my-postgres.example.com
  database: iot_db
  user: myuser
  password: mypassword
```

Install with custom values:

```bash
helm install my-demo-app ./deployments/helm/demo-app -f custom-values.yaml
```

### Enable Ingress

```yaml
frontend:
  ingress:
    enabled: true
    className: nginx
    hosts:
      - host: demo-app.example.com
        paths:
          - path: /
            pathType: Prefix
```

### Disable Metrics

```yaml
metrics:
  enabled: false
  serviceMonitor:
    enabled: false
```

### Scale Services

```yaml
generator:
  replicaCount: 3

backend:
  replicaCount: 5

frontend:
  replicaCount: 10
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade my-demo-app ./deployments/helm/demo-app -f custom-values.yaml

# Rollback if needed
helm rollback my-demo-app
```

## Uninstalling

```bash
helm uninstall my-demo-app
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl describe pod <pod-name>

# View logs
kubectl logs <pod-name>
```

### Common Issues

1. **ImagePullBackOff**: Update `global.imageRegistry` and image tags in values.yaml
2. **RabbitMQ Connection Failed**: Verify `rabbitmq.host` and credentials
3. **PostgreSQL Connection Failed**: Verify `postgresql.host` and credentials
4. **ServiceMonitor Not Working**: Ensure Prometheus Operator is installed

## Next Steps

- Review [README.md](./README.md) for complete documentation
- Configure persistent storage for PostgreSQL
- Set up TLS certificates for ingress
- Configure autoscaling for production
- Set up monitoring and alerting

## Support

For issues and questions:
- **Documentation**: https://github.com/procodus/demo-app
- **Issues**: https://github.com/procodus/demo-app/issues
