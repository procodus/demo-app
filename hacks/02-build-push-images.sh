#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REGISTRY_NAME="kind-registry"
REGISTRY_PORT="5001"
IMAGE_NAME="procodus/demo-app"
IMAGE_TAG="latest"
LOCAL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"
REGISTRY_IMAGE="localhost:${REGISTRY_PORT}/${IMAGE_NAME}:${IMAGE_TAG}"

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

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

    if ! command -v docker &> /dev/null; then
        log_error "docker is not installed"
        exit 1
    fi

    # Check if registry is running
    if [ ! "$(docker ps -q -f name=${REGISTRY_NAME})" ]; then
        log_error "Local registry ${REGISTRY_NAME} is not running"
        log_error "Please run ./01-setup-kind-cluster.sh first"
        exit 1
    fi

    log_info "All prerequisites satisfied"
}

build_image() {
    log_step "Building Docker image..."
    echo ""

    cd "${PROJECT_ROOT}"

    log_info "Building image: ${LOCAL_IMAGE}"
    log_info "Context: ${PROJECT_ROOT}"
    log_info "Dockerfile: Dockerfile"
    echo ""

    # Build the image with BuildKit
    DOCKER_BUILDKIT=1 docker build \
        --file Dockerfile \
        --tag "${LOCAL_IMAGE}" \
        --tag "${REGISTRY_IMAGE}" \
        --progress=plain \
        .

    echo ""
    log_info "Image built successfully: ${LOCAL_IMAGE}"
}

push_image() {
    log_step "Pushing image to local registry..."
    echo ""

    log_info "Pushing ${REGISTRY_IMAGE}"

    docker push "${REGISTRY_IMAGE}"

    echo ""
    log_info "Image pushed successfully"
}

verify_image() {
    log_step "Verifying image in registry..."

    # Check if image exists in local registry
    if curl -s "http://localhost:${REGISTRY_PORT}/v2/${IMAGE_NAME}/tags/list" | grep -q "${IMAGE_TAG}"; then
        log_info "Image verified in registry"
        log_info "Available at: localhost:${REGISTRY_PORT}/${IMAGE_NAME}:${IMAGE_TAG}"
    else
        log_error "Failed to verify image in registry"
        exit 1
    fi
}

print_image_info() {
    log_info "========================================"
    log_info "Image build and push complete!"
    log_info "========================================"
    echo ""
    log_info "Local image: ${LOCAL_IMAGE}"
    log_info "Registry image: ${REGISTRY_IMAGE}"
    echo ""
    log_info "Image size:"
    docker images "${LOCAL_IMAGE}" --format "  {{.Repository}}:{{.Tag}} - {{.Size}}"
    echo ""
    log_info "Next steps:"
    log_info "  Run ./03-deploy-helm.sh to deploy the application"
    echo ""
}

main() {
    log_info "Starting image build and push..."
    echo ""

    check_prerequisites
    build_image
    push_image
    verify_image

    echo ""
    print_image_info
}

main "$@"
