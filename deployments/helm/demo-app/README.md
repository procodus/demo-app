# Demo App Helm Chart

Helm chart for deploying the IoT data pipeline demo application with generator, backend, and frontend microservices.

## Overview

This Helm chart deploys a complete IoT data pipeline consisting of:

- **Generator**: Generates synthetic IoT sensor data and publishes to RabbitMQ
- **Backend**: Consumes messages from RabbitMQ, persists to PostgreSQL, and provides gRPC API
- **Frontend**: Web application that queries backend via gRPC and displays data

## Prerequisites

- Kubernetes 1.21+
- Helm 3.8+
- RabbitMQ (in-cluster or external)
- PostgreSQL (in-cluster or external)
- (Optional) Prometheus Operator for ServiceMonitor resources

## Installation

### Quick Start

```bash
# Add the Helm repository (if published)
helm repo add demo-app https://your-helm-repo.example.com
helm repo update

# Install with default values
helm install my-demo-app demo-app/demo-app

# Or install from local chart
helm install my-demo-app ./deployments/helm/demo-app
```

### Install with Custom Values

```bash
# Create a custom values file
cat > custom-values.yaml <<EOF
rabbitmq:
  host: my-rabbitmq
  user: myuser
  password: mypassword

postgresql:
  host: my-postgresql
  database: my_iot_db
  user: myuser
  password: mypassword

metrics:
  enabled: true
  serviceMonitor:
    enabled: true

frontend:
  replicaCount: 3
  ingress:
    enabled: true
    className: nginx
    hosts:
      - host: demo-app.example.com
        paths:
          - path: /
            pathType: Prefix
EOF

# Install with custom values
helm install my-demo-app ./deployments/helm/demo-app -f custom-values.yaml
```

## Configuration

### Image Configuration

By default, the chart uses placeholder images from GitHub Container Registry:

```yaml
global:
  imageRegistry: ghcr.io

generator:
  image:
    repository: procodus/demo-app
    tag: latest

backend:
  image:
    repository: procodus/demo-app
    tag: latest

frontend:
  image:
    repository: procodus/demo-app
    tag: latest
```

**Note**: Update these values to point to your actual container registry and image tags.

### RabbitMQ Configuration

#### Using In-Cluster RabbitMQ

```yaml
rabbitmq:
  host: rabbitmq  # Service name in cluster
  port: 5672
  user: guest
  password: guest
  sensorQueue: sensor-data
  deviceQueue: device-data
```

#### Using External RabbitMQ

```yaml
rabbitmq:
  host: rabbitmq.example.com
  port: 5672
  user: myuser
  password: mypassword
  # Or use full URL
  url: "amqp://myuser:mypassword@rabbitmq.example.com:5672"
```

### PostgreSQL Configuration

#### Using In-Cluster PostgreSQL

```yaml
postgresql:
  host: postgresql  # Service name in cluster
  port: 5432
  database: iot_db
  user: postgres
  password: postgres
  sslmode: disable
```

#### Using External PostgreSQL

```yaml
postgresql:
  host: postgres.example.com
  port: 5432
  database: my_iot_db
  user: myuser
  password: mypassword
  sslmode: require
```

### Prometheus Metrics

#### Enable Metrics Collection

```yaml
metrics:
  enabled: true  # Enables metrics endpoints on all services
```

**Metrics Endpoints**:
- Generator: `http://demo-app-generator:9091/metrics`
- Backend: `http://demo-app-backend:9090/metrics`
- Frontend: `http://demo-app-frontend:8080/metrics`

#### Enable ServiceMonitor (Prometheus Operator)

```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    labels:
      prometheus: kube-prometheus  # Match your Prometheus selector
    interval: 30s
    scrapeTimeout: 10s
```

**ServiceMonitor Resources**: Creates Prometheus Operator ServiceMonitor resources for automatic scraping:
- `demo-app-generator` - Scrapes generator metrics
- `demo-app-backend` - Scrapes backend metrics
- `demo-app-frontend` - Scrapes frontend metrics

### Application Configuration

#### Generator

```yaml
generator:
  enabled: true
  replicaCount: 1
  config:
    producerCount: 5      # Number of concurrent producers
    interval: "5s"        # Data generation interval
    logLevel: info
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi
```

#### Backend

```yaml
backend:
  enabled: true
  replicaCount: 1
  config:
    grpcPort: 50051
    logLevel: info
  livenessProbe:
    enabled: true
  readinessProbe:
    enabled: true
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 256Mi
```

#### Frontend

```yaml
frontend:
  enabled: true
  replicaCount: 2
  config:
    httpPort: 8080
    logLevel: info
  livenessProbe:
    enabled: true
    httpGet:
      path: /health
      port: 8080
  readinessProbe:
    enabled: true
    httpGet:
      path: /health
      port: 8080
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi
```

### Ingress Configuration

Enable ingress for external access to the frontend:

```yaml
frontend:
  ingress:
    enabled: true
    className: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
    hosts:
      - host: demo-app.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: demo-app-tls
        hosts:
          - demo-app.example.com
```

### Autoscaling

Enable Horizontal Pod Autoscaling (requires metrics-server):

```yaml
autoscaling:
  generator:
    enabled: true
    minReplicas: 1
    maxReplicas: 10
    targetCPUUtilizationPercentage: 80

  backend:
    enabled: true
    minReplicas: 1
    maxReplicas: 10
    targetCPUUtilizationPercentage: 80

  frontend:
    enabled: true
    minReplicas: 2
    maxReplicas: 20
    targetCPUUtilizationPercentage: 80
```

### Security

#### Pod Security Context

All pods run with restricted security context by default:

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL
```

#### Image Pull Secrets

For private container registries:

```yaml
global:
  imagePullSecrets:
    - name: ghcr-secret
```

Create the secret:

```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=your-username \
  --docker-password=your-token \
  --docker-email=your-email@example.com
```

## Usage Examples

### Development Environment

```bash
helm install dev-demo-app ./deployments/helm/demo-app \
  --set generator.config.producerCount=3 \
  --set generator.config.interval=1s \
  --set frontend.replicaCount=1 \
  --set backend.replicaCount=1 \
  --set metrics.enabled=false
```

### Production Environment

```bash
helm install prod-demo-app ./deployments/helm/demo-app \
  --namespace production \
  --create-namespace \
  -f production-values.yaml
```

Example `production-values.yaml`:

```yaml
global:
  imageRegistry: ghcr.io
  imagePullSecrets:
    - name: ghcr-secret

rabbitmq:
  host: rabbitmq-ha.infrastructure.svc.cluster.local
  user: iot-app
  password: ${RABBITMQ_PASSWORD}

postgresql:
  host: postgres-ha.infrastructure.svc.cluster.local
  database: iot_production
  user: iot-app
  password: ${POSTGRES_PASSWORD}
  sslmode: require

metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    labels:
      prometheus: kube-prometheus

generator:
  replicaCount: 5
  config:
    producerCount: 10
    interval: 5s
    logLevel: info
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 256Mi

backend:
  replicaCount: 3
  config:
    logLevel: info
  resources:
    limits:
      cpu: 2000m
      memory: 2Gi
    requests:
      cpu: 500m
      memory: 512Mi

frontend:
  replicaCount: 5
  config:
    logLevel: info
  ingress:
    enabled: true
    className: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      nginx.ingress.kubernetes.io/rate-limit: "100"
    hosts:
      - host: iot-dashboard.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: iot-dashboard-tls
        hosts:
          - iot-dashboard.example.com
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 256Mi

autoscaling:
  backend:
    enabled: true
    minReplicas: 3
    maxReplicas: 20
  frontend:
    enabled: true
    minReplicas: 5
    maxReplicas: 50
```

## Upgrading

### Upgrade Release

```bash
# Upgrade with new values
helm upgrade my-demo-app ./deployments/helm/demo-app -f custom-values.yaml

# Upgrade to new chart version
helm upgrade my-demo-app demo-app/demo-app --version 0.2.0
```

### Rollback Release

```bash
# View release history
helm history my-demo-app

# Rollback to previous version
helm rollback my-demo-app

# Rollback to specific revision
helm rollback my-demo-app 2
```

## Uninstallation

```bash
# Uninstall release
helm uninstall my-demo-app

# Uninstall from specific namespace
helm uninstall my-demo-app --namespace production
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/instance=my-demo-app
```

### View Pod Logs

```bash
# Generator logs
kubectl logs -l app.kubernetes.io/component=generator -f

# Backend logs
kubectl logs -l app.kubernetes.io/component=backend -f

# Frontend logs
kubectl logs -l app.kubernetes.io/component=frontend -f
```

### Check Service Endpoints

```bash
kubectl get svc -l app.kubernetes.io/instance=my-demo-app
```

### Verify Metrics

```bash
# Port-forward to metrics endpoints
kubectl port-forward svc/my-demo-app-generator 9091:9091
kubectl port-forward svc/my-demo-app-backend 9090:9090
kubectl port-forward svc/my-demo-app-frontend 8080:8080

# Curl metrics
curl http://localhost:9091/metrics
curl http://localhost:9090/metrics
curl http://localhost:8080/metrics
```

### Debug ServiceMonitor

```bash
# Check ServiceMonitor resources
kubectl get servicemonitor -l app.kubernetes.io/instance=my-demo-app

# Verify Prometheus targets
kubectl port-forward -n monitoring svc/prometheus-k8s 9090:9090
# Open http://localhost:9090/targets
```

## Parameters

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.imageRegistry` | Global container image registry | `ghcr.io` |
| `global.imagePullPolicy` | Global image pull policy | `IfNotPresent` |
| `global.imagePullSecrets` | Global image pull secrets | `[]` |

### RabbitMQ Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `rabbitmq.host` | RabbitMQ host | `rabbitmq` |
| `rabbitmq.port` | RabbitMQ port | `5672` |
| `rabbitmq.user` | RabbitMQ user | `guest` |
| `rabbitmq.password` | RabbitMQ password | `guest` |
| `rabbitmq.sensorQueue` | Sensor data queue name | `sensor-data` |
| `rabbitmq.deviceQueue` | Device data queue name | `device-data` |

### PostgreSQL Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `postgresql.host` | PostgreSQL host | `postgresql` |
| `postgresql.port` | PostgreSQL port | `5432` |
| `postgresql.database` | Database name | `iot_db` |
| `postgresql.user` | Database user | `postgres` |
| `postgresql.password` | Database password | `postgres` |
| `postgresql.sslmode` | SSL mode | `disable` |

### Metrics Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `metrics.enabled` | Enable metrics collection | `true` |
| `metrics.serviceMonitor.enabled` | Create ServiceMonitor resources | `true` |
| `metrics.serviceMonitor.interval` | Scrape interval | `30s` |

### Generator Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `generator.enabled` | Enable generator deployment | `true` |
| `generator.replicaCount` | Number of replicas | `1` |
| `generator.config.producerCount` | Number of producers | `5` |
| `generator.config.interval` | Generation interval | `5s` |

### Backend Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `backend.enabled` | Enable backend deployment | `true` |
| `backend.replicaCount` | Number of replicas | `1` |
| `backend.config.grpcPort` | gRPC port | `50051` |

### Frontend Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `frontend.enabled` | Enable frontend deployment | `true` |
| `frontend.replicaCount` | Number of replicas | `2` |
| `frontend.config.httpPort` | HTTP port | `8080` |
| `frontend.ingress.enabled` | Enable ingress | `false` |

See `values.yaml` for complete parameter documentation.

## License

Copyright © 2025 Demo App Team

## Support

- **Documentation**: https://github.com/procodus/demo-app
- **Issues**: https://github.com/procodus/demo-app/issues
