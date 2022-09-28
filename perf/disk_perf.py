#!/usr/bin/python

import datetime
import os
import re
import subprocess
import sys
import threading
import time


def usage():
    print """
usage:  disk_perf <thread count>   
  -- runs perf test in multiple threads and puts output in file inside /tmp/perf_tests/test_<thread number>.txt 
 
usage:  disk_perf --compute <thread count>
  -- parses output files found in /tmp/perf_tests/ and assembles averages
"""


SERIAL_RUNS = 10

WRITE_BANDWIDTH_STR = "disk write bandwidth (MB/s): "

READ_BANDWIDTH_STR = "disk read bandwidth (MB/s): "

PERF_OUTPUT_DIR = "/tmp/perf_tests"



def run_gpcheckperf(thread_index):
    output_file = os.path.join(PERF_OUTPUT_DIR, str(thread_index) + ".txt")
    data_dir = "/greenplum/test" + str(thread_index)
    # note that python -u required to get quick flush of stdout to the redirected file.
    cmd = "python -u /usr/local/greenplum-db/bin/gpcheckperf -v -h localhost -r ds -D -d %s >> %s 2>&1" % (data_dir, output_file)
    print("starting command: %s" % cmd)
    subprocess.check_output(cmd, shell=True)
    time.sleep(2)


def filter_lines(read_write_filter, output_file):
    with open(output_file, 'r') as perf:
        lines = perf.readlines()
        return [line for line in lines if read_write_filter in line]


def compute_average(read_write_filter):
    total = 0
    num_iterations = 0
    files = os.listdir(PERF_OUTPUT_DIR)
    for f in files:
        lines = filter_lines(read_write_filter, os.path.join(PERF_OUTPUT_DIR, f))
        pattern = re.compile(".* (\d+\.\d+) .*")
        for l in lines:
            match = pattern.match(l)
            if not match:
                print "no match for: %s" % l
            else:
                total += float(match.groups()[0])
                num_iterations += 1
    if num_iterations == 0:
        print("0 iterations found; no data to compute average.")
        return 0, 0
    return total, num_iterations


def validate_disk_size(num_threads):
    out = subprocess.check_output("cat /proc/meminfo | grep MemTotal", shell=True)
    # MemTotal:        7659096 kB
    words = out.split()
    # multidd creates a file of 2 * RAM in order to defeat any swap, which is assumed < 100% of RAM
    file_size = long(words[1]) * 2
    if "kB" in words[2]:
        file_size = file_size * 1024

    # /dev/sdb       209612800 31498228 178114572  16% /greenplum
    out = subprocess.check_output("df | grep /greenplum", shell=True)
    words = out.split()
    disk_size = long(words[1]) * 1024
    if disk_size < file_size * num_threads:
        raise Exception("disk size %s bytes < required size of single file %s bytes (2*RAM) for goal "
                        "of running %s simultaneous threads" %
                        (disk_size, file_size, num_threads))


def main():
    # todo use python standard arg parser so that --compute option can go anywhere in command line
    if len(sys.argv) < 2:
        usage()
        exit(1)

    if not os.path.isdir(PERF_OUTPUT_DIR):
        os.mkdir(PERF_OUTPUT_DIR)

    if sys.argv[1] == "--compute":
        num_threads = int(sys.argv[2])
        read_sum, read_iterations = compute_average(READ_BANDWIDTH_STR)
        if read_iterations:
            print("After %s runs, single thread average for READ: %s MB/s" % (read_iterations, read_sum / read_iterations))
            print("Average for whole system READ: %s MB/s\n\n" % (read_sum * num_threads / read_iterations))

        write_sum, write_iterations = compute_average(WRITE_BANDWIDTH_STR)
        if write_iterations:
            print("After %s runs, single thread average for WRITE: %s MB/s" % (write_iterations, write_sum / write_iterations))
            print("Average for whole system WRITE: %s MB/s\n\n" % (write_sum * num_threads / write_iterations))

    else:
        num_threads = int(sys.argv[1])
        validate_disk_size(num_threads)
        print("%s starting disk_perf with %s threads." % (datetime.datetime.now().isoformat(), num_threads))
        for i in range(0, num_threads):
            t = threading.Thread(target=worker, args=[i])
            t.start()


def worker(thread_index):
    for j in range(0, SERIAL_RUNS):
        run_gpcheckperf(thread_index)


if __name__ == "__main__":
    main()
