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
REGISTRY_NAME="kind-registry"
RELEASE_NAME="demo-app"
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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

confirm_cleanup() {
    log_warn "This will delete:"
    echo "  - Helm release: ${RELEASE_NAME}"
    echo "  - Kind cluster: ${CLUSTER_NAME}"
    echo "  - Local registry: ${REGISTRY_NAME}"
    echo ""

    read -p "Are you sure you want to continue? (yes/no): " -r
    echo ""

    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log_info "Cleanup cancelled"
        exit 0
    fi
}

uninstall_helm_release() {
    log_step "Uninstalling Helm release..."

    if ! command -v helm &> /dev/null; then
        log_warn "helm not found, skipping Helm release cleanup"
        return 0
    fi

    # Check if cluster exists
    if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        log_warn "Cluster ${CLUSTER_NAME} not found, skipping Helm uninstall"
        return 0
    fi

    # Set kubectl context
    kubectl config use-context "kind-${CLUSTER_NAME}" > /dev/null 2>&1 || true

    # Uninstall release if it exists
    if helm status ${RELEASE_NAME} -n ${NAMESPACE} &> /dev/null; then
        log_info "Uninstalling release ${RELEASE_NAME}..."
        helm uninstall ${RELEASE_NAME} -n ${NAMESPACE} --wait || true
        log_info "Release uninstalled"
    else
        log_warn "Release ${RELEASE_NAME} not found"
    fi
}

delete_kind_cluster() {
    log_step "Deleting Kind cluster..."

    if ! command -v kind &> /dev/null; then
        log_warn "kind not found, skipping cluster deletion"
        return 0
    fi

    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        log_info "Deleting cluster ${CLUSTER_NAME}..."
        kind delete cluster --name="${CLUSTER_NAME}"
        log_info "Cluster deleted"
    else
        log_warn "Cluster ${CLUSTER_NAME} not found"
    fi
}

delete_registry() {
    log_step "Deleting local registry..."

    if ! command -v docker &> /dev/null; then
        log_warn "docker not found, skipping registry deletion"
        return 0
    fi

    # Stop and remove registry container
    if [ "$(docker ps -aq -f name=${REGISTRY_NAME})" ]; then
        log_info "Removing registry container ${REGISTRY_NAME}..."
        docker rm -f ${REGISTRY_NAME} > /dev/null 2>&1 || true
        log_info "Registry deleted"
    else
        log_warn "Registry ${REGISTRY_NAME} not found"
    fi
}

cleanup_docker_images() {
    log_step "Cleaning up local Docker images..."

    if ! command -v docker &> /dev/null; then
        log_warn "docker not found, skipping image cleanup"
        return 0
    fi

    log_info "Removing demo-app images..."
    docker rmi -f $(docker images 'procodus/demo-app' -q) 2>/dev/null || true
    docker rmi -f $(docker images 'localhost:5001/procodus/demo-app' -q) 2>/dev/null || true

    log_info "Docker images cleaned up"
}

print_cleanup_summary() {
    log_info "========================================"
    log_info "Cleanup complete!"
    log_info "========================================"
    echo ""
    log_info "Removed:"
    echo "  ✓ Helm release: ${RELEASE_NAME}"
    echo "  ✓ Kind cluster: ${CLUSTER_NAME}"
    echo "  ✓ Local registry: ${REGISTRY_NAME}"
    echo "  ✓ Docker images"
    echo ""
    log_info "To recreate the environment:"
    log_info "  1. Run ./01-setup-kind-cluster.sh"
    log_info "  2. Run ./02-build-push-images.sh"
    log_info "  3. Run ./03-deploy-helm.sh"
    echo ""
}

main() {
    log_info "Starting cleanup..."
    echo ""

    confirm_cleanup
    uninstall_helm_release
    delete_kind_cluster
    delete_registry
    cleanup_docker_images

    echo ""
    print_cleanup_summary
}

main "$@"
