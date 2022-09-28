#!/usr/bin/python

import os
import sys
from subprocess import check_call, check_output
import time
import json


class Node:
    def __init__(self, name, label):
        self.name = name
        self.label = label


def existing_node_count(cluster_name):
    json_str = check_output("pks cluster %s --json" % cluster_name, shell=True)
    cluster = json.loads(json_str)
    return cluster["parameters"]["kubernetes_worker_instances"]


def main():
    """
    resize adjusts the size of the pool of nodes (VMs).
    It also assigns labels to those nodes, taking into account
    that nodes (VMs) may already exist with labels.
    """
    if len(sys.argv) < 2:
        print("usage: resize_node_pool <cluster_name>")
        exit(1)

    if not os.environ.get("SEGMENT_COUNT", None):
        print("SEGMENT_COUNT must be exported in current environment")
        exit(1)

    cluster_name = sys.argv[1]
    segment_count = int(os.environ['SEGMENT_COUNT'])
    required_nodes = 2 * int(segment_count) + 2

    # pks resize takes long when it is a no-op. check manually:
    existing = existing_node_count(cluster_name)
    print("existing nodes: %s, required: %s" % (existing, required_nodes))

    if existing >= required_nodes:
        print("already have enough (%s) nodes." % required_nodes)
        exit(0)

    print("Calling pks resize. This can take many minutes. Each dot is 5 seconds. This routine has no timeout.")

    check_call("pks resize %s --num-nodes=%s " %
               (cluster_name, required_nodes), shell=True)

    # todo pks 1.1 will have a --wait flag.  Remove this wait loop when it becomes available
    while True:
        pks_status = check_output("pks show-cluster %s | grep 'Last Action State'" % cluster_name,
                                  shell=True)
        if "succeeded" in pks_status:
            break
        time.sleep(5)
        sys.stdout.write('.')
        sys.stdout.flush()  # otherwise it buffers

    print("\n")


if __name__ == "__main__": main()
