# syntax = docker/dockerfile:experimental
ARG TAG_PREFIX=""
ARG GREENPLUM_VERSION=unknown
FROM golang:1.19 as build-in-docker

COPY greenplum-instance/ /greenplum-for-kubernetes/greenplum-instance/
COPY greenplum-operator/ /greenplum-for-kubernetes/greenplum-operator/
COPY pkg/ /greenplum-for-kubernetes/pkg/
COPY go.mod go.sum /greenplum-for-kubernetes/

WORKDIR /greenplum-for-kubernetes/greenplum-instance

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    mkdir buildcmd && \
    go build -o ./buildcmd/ \
    ./cmd/sshKeyScan \
    ./cmd/startGreenplumContainer  \
    ./cmd/initializeCluster \
    ./cmd/startPXF \
    ./cmd/runGpexpand \
    ./cmd/waitForKnownHosts

# build greenplum-instance image from here
FROM gcr.io/gp-kubernetes/ubuntu-gpdb-ent:${TAG_PREFIX}${GREENPLUM_VERSION}

# remove gppkg because it's misleading to users in the kubernetes environment
RUN rm -f /usr/local/greenplum-db/bin/gppkg

ARG TOOLS_DIR=/tools

COPY --from=build-in-docker \
    /greenplum-for-kubernetes/greenplum-instance/buildcmd/sshKeyScan \
    /greenplum-for-kubernetes/greenplum-instance/buildcmd/initializeCluster \
    /greenplum-for-kubernetes/greenplum-instance/buildcmd/startGreenplumContainer \
    /greenplum-for-kubernetes/greenplum-instance/buildcmd/startPXF \
    /greenplum-for-kubernetes/greenplum-instance/buildcmd/runGpexpand \
    /greenplum-for-kubernetes/greenplum-instance/buildcmd/waitForKnownHosts \
    ${TOOLS_DIR}/

COPY \
    greenplum-instance/scripts/gpexpand_job.sh \
    ${TOOLS_DIR}/

COPY greenplum-instance/scripts/gpadmin-limits.conf /etc/security/limits.d/

RUN \
    rm -f /etc/ssh/ssh_host_*_key* && \
    # comment out all other host key files
    sed -ri 's@^HostKey /etc/ssh/ssh_host_dsa_key$@#&@' /etc/ssh/sshd_config && \
    sed -ri 's@^HostKey /etc/ssh/ssh_host_ecdsa_key$@#&@' /etc/ssh/sshd_config && \
    sed -ri 's@^HostKey /etc/ssh/ssh_host_ed25519_key$@#&@' /etc/ssh/sshd_config && \
    # uncomment the rsa host key file
    sed -ri '\@^#HostKey /etc/ssh/ssh_host_rsa_key$@s@^#@@g' /etc/ssh/sshd_config && \
    # See https://gist.github.com/gasi/5691565
    sed -ri 's/UsePAM yes/UsePAM no/g' /etc/ssh/sshd_config && \
    # Disable password authentication so builds never hang given bad keys
    sed -ri 's/#PasswordAuthentication yes/PasswordAuthentication no/g' /etc/ssh/sshd_config && \
    mkdir -p /var/run/sshd && \
    chmod 0755 /var/run/sshd && \
    echo "gpadmin ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

ENV PXF_CONF=/etc/pxf
RUN mkdir -p ${PXF_CONF} && chown -R gpadmin:gpadmin ${PXF_CONF}

# Install tini
RUN wget https://github.com/krallin/tini/releases/download/v0.19.0/tini_0.19.0-amd64.deb && \
    dpkg -i tini_0.19.0-amd64.deb && \
    rm tini_0.19.0-amd64.deb
ENTRYPOINT ["/usr/bin/tini", "--"]

USER gpadmin

# USER env is referenced by `gpconfig`, `gpcheckcat`, `gpssh`, `gpscp`, `gpinitsegment`,
# and `gpload` utilities through `os.environ.get('USER')`
ENV USER=gpadmin
ENV PGPORT=5432
WORKDIR /home/gpadmin

# TODO update the deploy.sh etc. to only depends on /tools instead of /home/gpadmin/tools
RUN ln -s ${TOOLS_DIR} /home/gpadmin/tools

# Initialize pxf directories; needs GPHOME set
RUN /bin/bash -c 'source /usr/local/greenplum-db/greenplum_path.sh && /usr/local/pxf-gp6/bin/pxf init'

EXPOSE 22

ARG VERSION=dev
ARG DATE=unknown
ARG LICENSE="https://network.pivotal.io/legal_documents/vmware_eula"

LABEL name="Greenplum for Kubernetes" \
    vendor="Pivotal Software" \
    version="${VERSION}" \
    build_date="${DATE}" \
    license="${LICENSE}"
