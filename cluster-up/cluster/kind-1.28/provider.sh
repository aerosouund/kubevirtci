#!/usr/bin/env bash

set -e

DEFAULT_CLUSTER_NAME="kind-1.28"
DEFAULT_HOST_PORT=5000
ALTERNATE_HOST_PORT=5001
export CLUSTER_NAME=${CLUSTER_NAME:-$DEFAULT_CLUSTER_NAME}
export KUBEVIRT_NUM_NODES=${$KUBEVIRT_NUM_NODES:-1}

if [ $CLUSTER_NAME == $DEFAULT_CLUSTER_NAME ]; then
    export HOST_PORT=$DEFAULT_HOST_PORT
else
    export HOST_PORT=$ALTERNATE_HOST_PORT
fi

function set_kind_params() {
    version=$(cat cluster-up/cluster/$KUBEVIRT_PROVIDER/version)
    export KIND_VERSION="${KIND_VERSION:-$version}"

    image=$(cat cluster-up/cluster/$KUBEVIRT_PROVIDER/image)
    export KIND_NODE_IMAGE="${KIND_NODE_IMAGE:-$image}"
}

function configure_registry_proxy() {
    [ "$CI" != "true" ] && return

    echo "Configuring cluster nodes to work with CI mirror-proxy..."

    local -r ci_proxy_hostname="docker-mirror-proxy.kubevirt-prow.svc"
    local -r kind_binary_path="${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kind"
    local -r configure_registry_proxy_script="${KUBEVIRTCI_PATH}/cluster/kind/configure-registry-proxy.sh"

    KIND_BIN="$kind_binary_path" PROXY_HOSTNAME="$ci_proxy_hostname" $configure_registry_proxy_script
}

make -C cluster-provision/gocli cli

./cluster-provision/gocli/build/cli run-kind --with-extra-mounts=true --nodes=$KUBEVIRT_NUM_NODES
