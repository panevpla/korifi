#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/.."
API_DIR="$ROOT_DIR/api"
CTRL_DIR="$ROOT_DIR/controllers"
EIRINI_CONTROLLER_DIR="$ROOT_DIR/../eirini-controller"
export PATH="$PATH:$API_DIR/bin"

# undo *_IMG changes in config and reference
clean_up_img_refs() {
  cd "$ROOT_DIR"
  unset IMG_CONTROLLERS
  unset IMG_API
  make build-reference
}
trap clean_up_img_refs EXIT

ensure_kind_cluster() {
  if ! kind get clusters | grep -q "$cluster"; then
    current_cluster="$(kubectl config current-context)" || true
    cat <<EOF | kind create cluster --name "$cluster" --wait 5m --config=-
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
EOF
    if [[ -n "$current_cluster" ]]; then
      kubectl config use-context "$current_cluster"
    fi
  fi
  kind export kubeconfig --name "$cluster" --kubeconfig "$HOME/.kube/$cluster.yml"
}

deploy_cf_k8s_controllers() {
  pushd $ROOT_DIR >/dev/null
  {
    "$SCRIPT_DIR/install-dependencies.sh"
    export KUBEBUILDER_ASSETS=$ROOT_DIR/testbin/bin
    echo $PWD
    make generate-controllers
    IMG_CONTROLLERS=${IMG_CONTROLLERS:-"cf-k8s-controllers:$(uuidgen)"}
    export IMG_CONTROLLERS
    if [[ -z "${SKIP_DOCKER_BUILD:-}" ]]; then
      make docker-build-controllers
    fi
    kind load docker-image --name "$cluster" "$IMG_CONTROLLERS"
    make install-crds
    make deploy-controllers
  }
  popd >/dev/null
}

deploy_cf_k8s_api() {
  pushd $ROOT_DIR >/dev/null
  {
    IMG_API=${IMG_API:-"cf-k8s-api:$(uuidgen)"}
    export IMG_API
    if [[ -z "${SKIP_DOCKER_BUILD:-}" ]]; then
      make docker-build-api
    fi
    kind load docker-image --name "$cluster" "$IMG_API"
    make deploy-api-kind-auth
  }
  popd >/dev/null
}

cluster=${1:?specify cluster name}
ensure_kind_cluster "$cluster"
export KUBECONFIG="$HOME/.kube/$cluster.yml"
deploy_cf_k8s_controllers
deploy_cf_k8s_api
