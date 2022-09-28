#!/usr/bin/env bash

function main() {
    DEST_DEB_PATH=/tmp/greenplum.deb

    if [ $# -lt 1 ]; then
        echo "usage: update_gpdb.sh <path to new debian package>"
        exit 1
    fi

    local deb=$1

    local pods=$(kubectl get --no-headers=true pods -l app=gpdb -o custom-columns=:metadata.name |xargs)

    for pod in ${pods}; do
        echo "Running on pod: ${pod}"
        # delete existing apt-get-installed app
        kubectl exec ${pod} -- bash -l -c "sudo apt-get remove -y greenplum-db"
        kubectl cp ${deb} ${pod}:${DEST_DEB_PATH}
        kubectl exec ${pod} -- bash -l -c "sudo dpkg -i ${DEST_DEB_PATH}"
    done
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
