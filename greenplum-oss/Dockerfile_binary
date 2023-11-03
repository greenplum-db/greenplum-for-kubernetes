FROM ubuntu:18.04

RUN apt-get update \
 && apt-get install -y libapr1 libaprutil1 krb5-multidev libcurl3-gnutls libcurl4 libevent-2.1-6 libreadline7 libxml2 libyaml-0-2 libldap-2.4-2 openssh-client openssh-server perl rsync zip net-tools less iproute2 iputils-ping locales

RUN apt-get update \
 && apt-get install -y wget \
 && wget -q -O /tmp/gpdb.deb https://github.com/greenplum-db/gpdb/releases/download/6.23.0/greenplum-db-6.23.0-ubuntu18.04-amd64.deb \
 && dpkg -i /tmp/gpdb.deb
