#!/bin/bash
set -euxo pipefail

CURRENT_DIR=`pwd`
apt-get install -y ./${DEBIAN_PACKAGE:-deb_package_ubuntu16/greenplum-db.deb}

locale-gen en_US.UTF-8
${CURRENT_DIR}/gpdb_src/concourse/scripts/setup_gpadmin_user.bash

source /opt/gpdb/greenplum_path.sh

su gpadmin -c "source /opt/gpdb/greenplum_path.sh && \
    make -C gpdb_src/gpAux/gpdemo create-demo-cluster && \
    greenplum-for-kubernetes/concourse/scripts/oss-validation-gpdb.bash "
