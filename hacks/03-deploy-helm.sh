#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="demo-app-local"
RELEASE_NAME="demo-app"
NAMESPACE="demo-app"
REGISTRY_PORT="5001"
IMAGE_NAME="procodus/demo-app"
IMAGE_TAG="latest"

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHART_PATH="${PROJECT_ROOT}/deployments/helm/demo-app"

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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing_tools=()

    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi

    if ! command -v helm &> /dev/null; then
        missing_tools+=("helm")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi

    # Check if cluster exists
    if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        log_error "Kind cluster ${CLUSTER_NAME} not found"
        log_error "Please run ./01-setup-kind-cluster.sh first"
        exit 1
    fi

    # Set kubectl context
    kubectl config use-context "kind-${CLUSTER_NAME}" > /dev/null

    # Check if namespace exists
    if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
        log_error "Namespace ${NAMESPACE} not found"
        log_error "Please run ./01-setup-kind-cluster.sh first"
        exit 1
    fi

    log_info "All prerequisites satisfied"
}

validate_chart() {
    log_step "Validating Helm chart..."

    cd "${PROJECT_ROOT}"

    log_info "Linting chart: ${CHART_PATH}"
    helm lint "${CHART_PATH}"

    log_info "Chart validation passed"
}

deploy_application() {
    log_step "Deploying application with Helm..."
    echo ""

    cd "${PROJECT_ROOT}"

    # Check if release already exists
    if helm status ${RELEASE_NAME} -n ${NAMESPACE} &> /dev/null; then
        log_warn "Release ${RELEASE_NAME} already exists, upgrading..."

        helm upgrade ${RELEASE_NAME} "${CHART_PATH}" \
            --namespace ${NAMESPACE} \
            --set global.imageRegistry="localhost:${REGISTRY_PORT}" \
            --set global.imagePullPolicy=Always \
            --set generator.image.repository="${IMAGE_NAME}" \
            --set generator.image.tag="${IMAGE_TAG}" \
            --set backend.image.repository="${IMAGE_NAME}" \
            --set backend.image.tag="${IMAGE_TAG}" \
            --set frontend.image.repository="${IMAGE_NAME}" \
            --set frontend.image.tag="${IMAGE_TAG}" \
            --set rabbitmq.host=rabbitmq.${NAMESPACE}.svc.cluster.local \
            --set rabbitmq.user=guest \
            --set rabbitmq.password=guest \
            --set postgresql.host=postgresql.${NAMESPACE}.svc.cluster.local \
            --set postgresql.database=iot_db \
            --set postgresql.user=postgres \
            --set postgresql.password=postgres \
            --set metrics.enabled=true \
            --set metrics.serviceMonitor.enabled=false \
            --wait \
            --timeout=5m
    else
        log_info "Installing release ${RELEASE_NAME}..."

        helm install ${RELEASE_NAME} "${CHART_PATH}" \
            --namespace ${NAMESPACE} \
            --set global.imageRegistry="localhost:${REGISTRY_PORT}" \
            --set global.imagePullPolicy=Always \
            --set generator.image.repository="${IMAGE_NAME}" \
            --set generator.image.tag="${IMAGE_TAG}" \
            --set backend.image.repository="${IMAGE_NAME}" \
            --set backend.image.tag="${IMAGE_TAG}" \
            --set frontend.image.repository="${IMAGE_NAME}" \
            --set frontend.image.tag="${IMAGE_TAG}" \
            --set rabbitmq.host=rabbitmq.${NAMESPACE}.svc.cluster.local \
            --set rabbitmq.user=guest \
            --set rabbitmq.password=guest \
            --set postgresql.host=postgresql.${NAMESPACE}.svc.cluster.local \
            --set postgresql.database=iot_db \
            --set postgresql.user=postgres \
            --set postgresql.password=postgres \
            --set metrics.enabled=true \
            --set metrics.serviceMonitor.enabled=false \
            --wait \
            --timeout=5m
    fi

    echo ""
    log_info "Application deployed successfully"
}

wait_for_pods() {
    log_step "Waiting for pods to be ready..."

    log_info "Waiting for generator..."
    kubectl wait --for=condition=ready pod \
        -l app.kubernetes.io/component=generator \
        -n ${NAMESPACE} \
        --timeout=300s || true

    log_info "Waiting for backend..."
    kubectl wait --for=condition=ready pod \
        -l app.kubernetes.io/component=backend \
        -n ${NAMESPACE} \
        --timeout=300s || true

    log_info "Waiting for frontend..."
    kubectl wait --for=condition=ready pod \
        -l app.kubernetes.io/component=frontend \
        -n ${NAMESPACE} \
        --timeout=300s || true

    log_info "All pods are ready"
}

print_deployment_info() {
    log_info "========================================"
    log_info "Deployment complete!"
    log_info "========================================"
    echo ""

    # Get pod status
    log_info "Pod Status:"
    kubectl get pods -n ${NAMESPACE} -l app.kubernetes.io/instance=${RELEASE_NAME}
    echo ""

    # Get service status
    log_info "Services:"
    kubectl get svc -n ${NAMESPACE} -l app.kubernetes.io/instance=${RELEASE_NAME}
    echo ""

    log_info "Access the application:"
    echo ""
    log_info "Frontend (Web UI):"
    log_info "  kubectl port-forward -n ${NAMESPACE} svc/${RELEASE_NAME}-frontend 8080:8080"
    log_info "  Then open: http://localhost:8080"
    echo ""
    log_info "Backend gRPC API:"
    log_info "  kubectl port-forward -n ${NAMESPACE} svc/${RELEASE_NAME}-backend 50051:50051"
    echo ""
    log_info "RabbitMQ Management UI:"
    log_info "  kubectl port-forward -n ${NAMESPACE} svc/rabbitmq 15672:15672"
    log_info "  Then open: http://localhost:15672 (guest/guest)"
    echo ""
    log_info "Prometheus Metrics:"
    log_info "  Generator: kubectl port-forward -n ${NAMESPACE} svc/${RELEASE_NAME}-generator 9091:9091"
    log_info "             curl http://localhost:9091/metrics"
    log_info "  Backend:   kubectl port-forward -n ${NAMESPACE} svc/${RELEASE_NAME}-backend 9090:9090"
    log_info "             curl http://localhost:9090/metrics"
    log_info "  Frontend:  kubectl port-forward -n ${NAMESPACE} svc/${RELEASE_NAME}-frontend 8080:8080"
    log_info "             curl http://localhost:8080/metrics"
    echo ""
    log_info "View logs:"
    log_info "  kubectl logs -f -n ${NAMESPACE} -l app.kubernetes.io/component=generator"
    log_info "  kubectl logs -f -n ${NAMESPACE} -l app.kubernetes.io/component=backend"
    log_info "  kubectl logs -f -n ${NAMESPACE} -l app.kubernetes.io/component=frontend"
    echo ""
    log_info "Cleanup:"
    log_info "  Run ./04-cleanup.sh to remove everything"
    echo ""
}

main() {
    log_info "Starting Helm deployment..."
    echo ""

    check_prerequisites
    validate_chart
    deploy_application
    wait_for_pods

    echo ""
    print_deployment_info
}

main "$@"
