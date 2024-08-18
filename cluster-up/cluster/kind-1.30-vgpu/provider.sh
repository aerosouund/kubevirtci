#!/usr/bin/env bash

set -e

DEFAULT_CLUSTER_NAME="vgpu"
DEFAULT_HOST_PORT=5000
ALTERNATE_HOST_PORT=5001
export CLUSTER_NAME=${CLUSTER_NAME:-$DEFAULT_CLUSTER_NAME}

if [ $CLUSTER_NAME == $DEFAULT_CLUSTER_NAME ]; then
    export HOST_PORT=$DEFAULT_HOST_PORT
else
    export HOST_PORT=$ALTERNATE_HOST_PORT
fi

make -C cluster-provision/gocli cli

function up() {
    # print hardware info for easier debugging based on logs
    echo 'Available cards'
    ${CRI_BIN} run --rm --cap-add=SYS_RAWIO quay.io/phoracek/lspci@sha256:0f3cacf7098202ef284308c64e3fc0ba441871a846022bb87d65ff130c79adb1 sh -c "lspci -k | grep -EA2 'VGA|3D'"
    echo ""

    if [ "$CI" == "true" ]; then export REGISTRY_PROXY="docker-mirror-proxy.kubevirt-prow.svc"; fi
    ./cluster-provision/gocli/build/cli run-kind vgpu \
        --with-extra-mounts=true \
        --nodes=$KUBEVIRT_NUM_NODES \
        --registry-proxy=$REGISTRY_PROXY
}
