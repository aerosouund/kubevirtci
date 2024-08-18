#!/usr/bin/env bash

set -e

DEFAULT_CLUSTER_NAME="kind-1.28"
DEFAULT_HOST_PORT=5000
ALTERNATE_HOST_PORT=5001
export CLUSTER_NAME=${CLUSTER_NAME:-$DEFAULT_CLUSTER_NAME}
KUBEVIRT_NUM_NODES=${KUBEVIRT_NUM_NODES:-1}
REGISTRY_PROXY=""

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

make -C cluster-provision/gocli cli

function up() {
    if [ "$CI" != "false" ]; then export REGISTRY_PROXY="docker-mirror-proxy.kubevirt-prow.svc"; fi
    ./cluster-provision/gocli/build/cli run-kind k8s-1.28 --with-extra-mounts=true --nodes=$KUBEVIRT_NUM_NODES --registry-proxy=$REGISTRY_PROXY
}

