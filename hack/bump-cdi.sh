#!/usr/bin/env bash
#
# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright 2021 Red Hat, Inc.

set -euo pipefail

function usage() {
    cat <<EOF
Usage: $0 [-h|--help]

        bump cdi manifests in all k8s provider directories to
        latest version

    Arguments:

         -h|--help  show this help text
EOF
}

if [ $# -gt 0 ]; then
    if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
        usage
        exit 0
    fi
    usage
    echo "Unknown argument $1"
    exit 1
fi


cdi_dir="./cluster-provision/gocli/opts/cdi/"

if [ -d "$cdi_dir" ]; then
    # Execute the fetch-latest-cdi.sh script on the directory
    if ! ./cluster-provision/k8s/fetch-latest-cdi.sh -f "$cdi_dir"; then
        echo "Failed to update cdi manifests for $cdi_dir"
        exit 1
    else
        echo "Updated cdi manifests for $cdi_dir"
    fi

    # Bump versions in the two files within the directory
    for file in "$cdi_dir"/*; do
        # Replace version numbers in the file (example: bumping from 1.0.0 to 1.1.0)
        sed -i 's/1\.0\.0/1.1.0/g' "$file"
    done

else
    echo "Directory $cdi_dir does not exist."
    exit 1
fi

