#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="demo-app-local"
REGISTRY_NAME="kind-registry"
REGISTRY_PORT="5001"
NAMESPACE="demo-app"

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing_tools=()

    if ! command -v docker &> /dev/null; then
        missing_tools+=("docker")
    fi

    if ! command -v kind &> /dev/null; then
        missing_tools+=("kind")
    fi

    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi

    if ! command -v helm &> /dev/null; then
        missing_tools+=("helm")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_error "Please install them before running this script"
        exit 1
    fi

    log_info "All prerequisites satisfied"
}

create_local_registry() {
    log_info "Creating local Docker registry..."

    # Check if registry already exists
    if [ "$(docker ps -q -f name=${REGISTRY_NAME})" ]; then
        log_warn "Registry ${REGISTRY_NAME} already running"
        return 0
    fi

    if [ "$(docker ps -aq -f name=${REGISTRY_NAME})" ]; then
        log_info "Starting existing registry container"
        docker start ${REGISTRY_NAME}
    else
        log_info "Creating new registry container"
        docker run -d \
            --restart=always \
            -p "127.0.0.1:${REGISTRY_PORT}:5000" \
            --name ${REGISTRY_NAME} \
            registry:2
    fi

    log_info "Registry available at localhost:${REGISTRY_PORT}"
}

create_kind_cluster() {
    log_info "Creating Kind cluster: ${CLUSTER_NAME}..."

    # Check if cluster already exists
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        log_warn "Cluster ${CLUSTER_NAME} already exists"
        kind get kubeconfig --name="${CLUSTER_NAME}" > /dev/null
        return 0
    fi

    # Create cluster config with registry
    cat <<EOF | kind create cluster --name="${CLUSTER_NAME}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
  - containerPort: 30080
    hostPort: 30080
    protocol: TCP
  - containerPort: 30443
    hostPort: 30443
    protocol: TCP
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${REGISTRY_PORT}"]
    endpoint = ["http://${REGISTRY_NAME}:5000"]
EOF

    log_info "Cluster ${CLUSTER_NAME} created successfully"
}

connect_registry_to_cluster() {
    log_info "Connecting registry to cluster network..."

    # Connect registry to kind network if not already connected
    if [ ! "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${REGISTRY_NAME}")" = 'null' ]; then
        log_warn "Registry already connected to kind network"
        return 0
    fi

    docker network connect "kind" "${REGISTRY_NAME}" 2>/dev/null || true

    # Document the local registry
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${REGISTRY_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

    log_info "Registry connected to cluster"
}

create_namespace() {
    log_info "Creating namespace: ${NAMESPACE}..."

    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

    log_info "Namespace ${NAMESPACE} created"
}

deploy_rabbitmq() {
    log_info "Deploying RabbitMQ to cluster..."

    kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: rabbitmq
  namespace: ${NAMESPACE}
  labels:
    app: rabbitmq
spec:
  type: ClusterIP
  ports:
  - port: 5672
    targetPort: 5672
    name: amqp
  - port: 15672
    targetPort: 15672
    name: management
  selector:
    app: rabbitmq
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq
  namespace: ${NAMESPACE}
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
          name: amqp
        - containerPort: 15672
          name: management
        env:
        - name: RABBITMQ_DEFAULT_USER
          value: "guest"
        - name: RABBITMQ_DEFAULT_PASS
          value: "guest"
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        readinessProbe:
          exec:
            command: ["rabbitmq-diagnostics", "ping"]
          initialDelaySeconds: 20
          periodSeconds: 10
        livenessProbe:
          exec:
            command: ["rabbitmq-diagnostics", "ping"]
          initialDelaySeconds: 30
          periodSeconds: 30
EOF

    log_info "Waiting for RabbitMQ to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/rabbitmq -n ${NAMESPACE}

    log_info "RabbitMQ deployed successfully"
}

deploy_postgresql() {
    log_info "Deploying PostgreSQL to cluster..."

    kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: postgresql
  namespace: ${NAMESPACE}
  labels:
    app: postgresql
spec:
  type: ClusterIP
  ports:
  - port: 5432
    targetPort: 5432
    name: postgresql
  selector:
    app: postgresql
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgresql
  namespace: ${NAMESPACE}
spec:
  serviceName: postgresql
  replicas: 1
  selector:
    matchLabels:
      app: postgresql
  template:
    metadata:
      labels:
        app: postgresql
    spec:
      containers:
      - name: postgresql
        image: postgres:16-alpine
        ports:
        - containerPort: 5432
          name: postgresql
        env:
        - name: POSTGRES_DB
          value: "iot_db"
        - name: POSTGRES_USER
          value: "postgres"
        - name: POSTGRES_PASSWORD
          value: "postgres"
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        volumeMounts:
        - name: postgresql-data
          mountPath: /var/lib/postgresql/data
        readinessProbe:
          exec:
            command: ["pg_isready", "-U", "postgres"]
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          exec:
            command: ["pg_isready", "-U", "postgres"]
          initialDelaySeconds: 30
          periodSeconds: 10
  volumeClaimTemplates:
  - metadata:
      name: postgresql-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
EOF

    log_info "Waiting for PostgreSQL to be ready..."
    kubectl wait --for=condition=ready --timeout=300s pod -l app=postgresql -n ${NAMESPACE}

    log_info "PostgreSQL deployed successfully"
}

print_cluster_info() {
    log_info "========================================"
    log_info "Kind cluster setup complete!"
    log_info "========================================"
    echo ""
    log_info "Cluster name: ${CLUSTER_NAME}"
    log_info "Namespace: ${NAMESPACE}"
    log_info "Local registry: localhost:${REGISTRY_PORT}"
    echo ""
    log_info "RabbitMQ:"
    log_info "  - AMQP: rabbitmq.${NAMESPACE}.svc.cluster.local:5672"
    log_info "  - Management: http://localhost:15672 (guest/guest)"
    log_info "    Port-forward: kubectl port-forward -n ${NAMESPACE} svc/rabbitmq 15672:15672"
    echo ""
    log_info "PostgreSQL:"
    log_info "  - Host: postgresql.${NAMESPACE}.svc.cluster.local:5432"
    log_info "  - Database: iot_db"
    log_info "  - User: postgres"
    log_info "  - Password: postgres"
    echo ""
    log_info "Next steps:"
    log_info "  1. Run ./02-build-push-images.sh to build and push images"
    log_info "  2. Run ./03-deploy-helm.sh to deploy the application"
    echo ""
}

main() {
    log_info "Starting Kind cluster setup..."
    echo ""

    check_prerequisites
    create_local_registry
    create_kind_cluster
    connect_registry_to_cluster
    create_namespace
    deploy_rabbitmq
    deploy_postgresql

    echo ""
    print_cluster_info
}

main "$@"
