# Deployment Guide

This guide covers deploying the Demo App IoT Data Pipeline to various environments.

## Table of Contents

- [Deployment Options](#deployment-options)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Helm Deployment](#helm-deployment)
- [Production Considerations](#production-considerations)

## Deployment Options

| Method | Use Case | Complexity | Scalability |
|--------|----------|------------|-------------|
| **Local Binary** | Development | Low | None |
| **Docker Compose** | Testing | Low | Limited |
| **Docker Containers** | Small deployments | Medium | Limited |
| **Kubernetes** | Production | High | Excellent |
| **Helm Chart** | Production (K8s) | Medium | Excellent |

## Docker Deployment

### Build Multi-Arch Images

```bash
# Build for local architecture
docker build -t demo-app:latest -f deployments/Dockerfile .

# Build for multiple architectures
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ghcr.io/procodus/demo-app:latest \
  --push \
  -f deployments/Dockerfile .
```

### Run with Docker

**1. Start Infrastructure**:
```bash
# Create network
docker network create demo-network

# Start PostgreSQL
docker run -d \
  --name postgres \
  --network demo-network \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=iot_db \
  -p 5432:5432 \
  postgres:16-alpine

# Start RabbitMQ
docker run -d \
  --name rabbitmq \
  --network demo-network \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management-alpine
```

**2. Run Backend**:
```bash
docker run -d \
  --name demo-backend \
  --network demo-network \
  -p 50051:50051 \
  -p 9090:9090 \
  -e APP_BACKEND_DB_HOST=postgres \
  -e APP_BACKEND_DB_PASSWORD=postgres \
  -e APP_BACKEND_RABBITMQ_URL=amqp://rabbitmq:5672 \
  ghcr.io/procodus/demo-app:latest backend
```

**3. Run Generator**:
```bash
docker run -d \
  --name demo-generator \
  --network demo-network \
  -p 9091:9091 \
  -e APP_GENERATOR_RABBITMQ_URL=amqp://rabbitmq:5672 \
  -e APP_GENERATOR_NUM_DEVICES=20 \
  ghcr.io/procodus/demo-app:latest generator
```

**4. Run Frontend**:
```bash
docker run -d \
  --name demo-frontend \
  --network demo-network \
  -p 8080:8080 \
  -e APP_FRONTEND_BACKEND_URL=demo-backend:50051 \
  ghcr.io/procodus/demo-app:latest frontend
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: iot_db
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    image: ghcr.io/procodus/demo-app:latest
    command: backend
    ports:
      - "50051:50051"
      - "9090:9090"
    environment:
      APP_BACKEND_DB_HOST: postgres
      APP_BACKEND_DB_PASSWORD: postgres
      APP_BACKEND_RABBITMQ_URL: amqp://rabbitmq:5672
    depends_on:
      postgres:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
    restart: unless-stopped

  generator:
    image: ghcr.io/procodus/demo-app:latest
    command: generator
    ports:
      - "9091:9091"
    environment:
      APP_GENERATOR_RABBITMQ_URL: amqp://rabbitmq:5672
      APP_GENERATOR_NUM_DEVICES: 20
      APP_GENERATOR_INTERVAL: 10s
    depends_on:
      rabbitmq:
        condition: service_healthy
    restart: unless-stopped

  frontend:
    image: ghcr.io/procodus/demo-app:latest
    command: frontend
    ports:
      - "8080:8080"
    environment:
      APP_FRONTEND_BACKEND_URL: backend:50051
    depends_on:
      - backend
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9000:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

volumes:
  postgres-data:
  prometheus-data:
```

Start services:
```bash
docker-compose up -d
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (Kind, Minikube, GKE, EKS, AKS)
- kubectl configured
- Helm 3+ (for Helm deployment)

### Manual Kubernetes Deployment

**1. Create Namespace**:
```bash
kubectl create namespace demo-app
```

**2. Deploy PostgreSQL**:
```yaml
# postgres.yaml
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: demo-app
type: Opaque
stringData:
  password: postgres
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: demo-app
spec:
  ports:
    - port: 5432
  selector:
    app: postgres
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: demo-app
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:16-alpine
          env:
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: password
            - name: POSTGRES_DB
              value: iot_db
          ports:
            - containerPort: 5432
          volumeMounts:
            - name: postgres-storage
              mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
    - metadata:
        name: postgres-storage
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
```

Apply:
```bash
kubectl apply -f postgres.yaml
```

**3. Deploy RabbitMQ**:
```yaml
# rabbitmq.yaml
apiVersion: v1
kind: Service
metadata:
  name: rabbitmq
  namespace: demo-app
spec:
  ports:
    - name: amqp
      port: 5672
    - name: management
      port: 15672
  selector:
    app: rabbitmq
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq
  namespace: demo-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rabbitmq
  template:
    metadata:
      labels:
        app: rabbitmq
    spec:
      containers:
        - name: rabbitmq
          image: rabbitmq:3-management-alpine
          ports:
            - containerPort: 5672
            - containerPort: 15672
```

Apply:
```bash
kubectl apply -f rabbitmq.yaml
```

**4. Deploy Backend**:
```yaml
# backend.yaml
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: demo-app
spec:
  ports:
    - name: grpc
      port: 50051
    - name: metrics
      port: 9090
  selector:
    app: backend
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: demo-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
        - name: backend
          image: ghcr.io/procodus/demo-app:latest
          command: ["/app/demo-app"]
          args: ["backend"]
          env:
            - name: APP_BACKEND_DB_HOST
              value: postgres
            - name: APP_BACKEND_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: password
            - name: APP_BACKEND_RABBITMQ_URL
              value: amqp://rabbitmq:5672
          ports:
            - containerPort: 50051
            - containerPort: 9090
          livenessProbe:
            httpGet:
              path: /metrics
              port: 9090
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /metrics
              port: 9090
            initialDelaySeconds: 10
            periodSeconds: 5
```

Apply:
```bash
kubectl apply -f backend.yaml
```

**5. Deploy Generator and Frontend** (similar structure)

### Verify Deployment

```bash
# Check pods
kubectl get pods -n demo-app

# Check services
kubectl get svc -n demo-app

# View logs
kubectl logs -n demo-app -l app=backend -f

# Port forward to access frontend
kubectl port-forward -n demo-app svc/frontend 8080:8080
```

## Helm Deployment

### Install from OCI Registry

```bash
# Add repository
helm pull oci://ghcr.io/procodus/charts/demo-app --version 1.0.0

# Install with default values
helm install demo-app oci://ghcr.io/procodus/charts/demo-app \
  --version 1.0.0 \
  --namespace demo-app \
  --create-namespace

# Install with custom values
helm install demo-app oci://ghcr.io/procodus/charts/demo-app \
  --version 1.0.0 \
  --namespace demo-app \
  --create-namespace \
  --set generator.replicaCount=2 \
  --set backend.database.host=external-postgres.example.com \
  --set backend.database.password=secret
```

### Custom Values File

Create `values.yaml`:

```yaml
# Generator configuration
generator:
  replicaCount: 2
  image:
    tag: "1.0.0"
  config:
    numDevices: 100
    interval: "10s"
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 250m
      memory: 256Mi

# Backend configuration
backend:
  replicaCount: 3
  image:
    tag: "1.0.0"
  database:
    host: postgres.default.svc.cluster.local
    port: 5432
    name: iot_db
    user: postgres
    password: secret  # Use secrets in production
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 500m
      memory: 512Mi

# Frontend configuration
frontend:
  replicaCount: 2
  image:
    tag: "1.0.0"
  ingress:
    enabled: true
    className: nginx
    hosts:
      - host: demo-app.example.com
        paths:
          - path: /
            pathType: Prefix
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 250m
      memory: 256Mi

# External services
postgresql:
  enabled: true
  auth:
    password: postgres
    database: iot_db
  primary:
    persistence:
      size: 10Gi

rabbitmq:
  enabled: true
  auth:
    username: guest
    password: guest
  persistence:
    size: 8Gi

# Monitoring
prometheus:
  enabled: true
  serviceMonitor:
    enabled: true
```

Install with values file:
```bash
helm install demo-app oci://ghcr.io/procodus/charts/demo-app \
  --version 1.0.0 \
  --namespace demo-app \
  --create-namespace \
  -f values.yaml
```

### Helm Operations

```bash
# Upgrade release
helm upgrade demo-app oci://ghcr.io/procodus/charts/demo-app \
  --version 1.1.0 \
  --namespace demo-app \
  -f values.yaml

# Rollback
helm rollback demo-app 1 --namespace demo-app

# Uninstall
helm uninstall demo-app --namespace demo-app

# List releases
helm list --namespace demo-app

# Get values
helm get values demo-app --namespace demo-app
```

## Production Considerations

### High Availability

**Backend**:
- Run 3+ replicas for HA
- Use PodDisruptionBudget
- Deploy across multiple availability zones

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: backend-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: backend
```

**Database**:
- Use managed PostgreSQL (RDS, Cloud SQL, Azure Database)
- Configure read replicas
- Enable automatic backups
- Set up point-in-time recovery

**Message Queue**:
- Use RabbitMQ cluster (3+ nodes)
- Or managed service (Amazon MQ, CloudAMQP)
- Configure durable queues
- Enable mirroring

### Security

**1. Use Secrets**:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: demo-app-secrets
type: Opaque
stringData:
  db-password: ${DB_PASSWORD}
  rabbitmq-password: ${RABBITMQ_PASSWORD}
```

**2. Network Policies**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-network-policy
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: frontend
      ports:
        - protocol: TCP
          port: 50051
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
```

**3. Enable TLS**:
- Use cert-manager for certificates
- Configure TLS for gRPC
- Use AMQPS for RabbitMQ

### Monitoring

**Prometheus ServiceMonitor**:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: demo-app-metrics
spec:
  selector:
    matchLabels:
      app: backend
  endpoints:
    - port: metrics
      interval: 30s
```

See [Monitoring Guide](./monitoring.md) for complete setup.

### Resource Limits

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

### Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### Backup Strategy

**Database Backups**:
- Daily automated backups
- 30-day retention
- Cross-region replication

**Configuration Backups**:
```bash
# Backup Helm values
helm get values demo-app -n demo-app > backup-values.yaml

# Backup all resources
kubectl get all -n demo-app -o yaml > backup-resources.yaml
```

## Next Steps

- [Monitoring Guide](./monitoring.md) - Set up monitoring
- [Configuration Guide](./configuration.md) - Tune settings
- [Troubleshooting](./troubleshooting.md) - Debug issues
