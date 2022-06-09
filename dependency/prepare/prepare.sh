#!/usr/bin/env bash

set -u
set -o pipefail

function main() {
  local os isMacOs
  os="${1}"

  echo "Running on ${os}"
  echo "${os}" | grep "macos"
  isMacOs=$?

  if [[ $isMacOs -eq 0 ]]; then
    brew install coreutils
    alias gensha256=shasum --algorithm=256
  else # assume ubuntu
    alias gensha256=sha256sum
  fi

  return 0
}

main "${@:-}"