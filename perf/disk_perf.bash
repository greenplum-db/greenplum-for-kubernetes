#!/usr/bin/env bash

export LD_LIBRARY_PATH=/usr/local/greenplum-db/lib
export PYTHONPATH=/usr/local/greenplum-db/lib/python
export GPHOME=/usr/local/greenplum-db

source /usr/local/greenplum-db/greenplum_path.sh

nohup /home/gpadmin/disk_perf.py 10 &
