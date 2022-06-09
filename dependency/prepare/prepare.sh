#!/usr/bin/env bash

set -eu
set -o pipefail

function main() {
  local os isMacOs
  os="${1}"

  echo "${os}" | grep --quiet "macos"
  isMacOs=$?

  if [[ $isMacOs -eq 0 ]]; then
    brew install coreutils
  fi

  return 0
}

main "${@:-}"