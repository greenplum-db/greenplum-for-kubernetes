FROM ubuntu:18.04

COPY README.Ubuntu.bash /gpdb_src/README.Ubuntu.bash
COPY concourse/ /gpdb_src/concourse/

WORKDIR /gpdb_src

RUN ./README.Ubuntu.bash \
&& chmod +x concourse/scripts/setup_gpadmin_user.bash \
&& concourse/scripts/setup_gpadmin_user.bash
