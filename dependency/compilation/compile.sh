#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly NAME="bundler"

function main() {
  local version tarball_name temp_dir output_dir
  version="${1}"

  echo "version=${version}"

  tarball_name="${NAME}-${version}.tgz"

  temp_dir="$(mktemp -d)"
  output_dir="$(PWD)"

  pushd "${temp_dir}" > /dev/null
    export GEM_HOME="${PWD}"
    unset RUBYOPT; \
      GEM_HOME="${PWD}" \
      GEM_PATH="${PWD}" \
      gem install bundler \
        --version "${version}" \
        --no-document \
        --env-shebang

    rm -f "bundler-${version}.gem"
    rm -rf "cache/bundler-${version}.gem"
    tar czvf "${output_dir}/${tarball_name}" .
  popd > /dev/null

  rm -rf "${temp_dir}"
}

main "${@:-}"