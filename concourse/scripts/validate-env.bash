#!/usr/bin/env bash

function validate_env_vars() {
  declare -a env_vars=("${@}")
  exit_code=0
  for env_var in "${env_vars[@]}"; do
    if ! validate_env_var "${env_var}" ; then
      exit_code=1
    fi
  done
  return $exit_code
}

function validate_env_var() {
  env_var="${1}"
  env_var_value="${!1}"
  if [ -z "${env_var_value}" ]; then
    echo "${env_var} environment variable is required"
    return 1
  fi
  return 0
}
