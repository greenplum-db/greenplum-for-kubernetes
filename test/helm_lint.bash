#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# run the test
helm lint ${SCRIPT_DIR}/../greenplum-operator/operator
