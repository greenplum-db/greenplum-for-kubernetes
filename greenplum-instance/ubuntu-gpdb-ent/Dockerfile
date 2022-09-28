FROM gcr.io/gp-kubernetes/gpdb6-ubuntu18.04-base-gp4k8s as install-gpdb

COPY greenplum-debian-binary/greenplum-db-*.deb /tmp
COPY madlib-deb-ubuntu18/apache-madlib-*.deb /tmp
COPY pxf-gp6-deb-ubuntu18/pxf-gp6-*.deb /tmp

RUN set -x && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get clean && \
    apt-get install -y \
              /tmp/greenplum-db-*.deb \
              /tmp/apache-madlib-*.deb \
              /tmp/pxf-gp6-*.deb && \
    groupadd -g 1000 gpadmin && \
    useradd gpadmin -u 1000 -g gpadmin -G tty -m -p '*' -s /bin/bash && \
    chown -R gpadmin:gpadmin /usr/local/greenplum-db/ && \
    chown -R gpadmin:gpadmin /usr/local/pxf-gp6/ && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ADD gpbackup-ubuntu18/pivotal_greenplum_backup_restore-*.tar.gz /usr/local/greenplum-db/
# maybe need a chown after?
#COPY --from=assemble-components --chown=gpadmin:gpadmin /tmp/gpbackup/ /usr/local/greenplum-db/

ENV PATH=/usr/local/madlib/bin:${PATH}
ENV JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64/

ARG GREENPLUM_VERSION=unknown
ARG MADLIB_VERSION=unknown
ARG PXF_VERSION=unknown
ARG GPBACKUP_VERSION=unknown

ENV GREENPLUM_VERSION=${GREENPLUM_VERSION}
LABEL \
    greenplum_version=${GREENPLUM_VERSION} \
    madlib_version=${MADLIB_VERSION} \
    pxf_version=${PXF_VERSION} \
    gpbackup_version=${GPBACKUP_VERSION} \
