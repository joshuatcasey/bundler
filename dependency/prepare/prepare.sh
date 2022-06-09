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
  fi

  return 0
}

main "${@:-}"