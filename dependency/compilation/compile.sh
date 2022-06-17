#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly NAME="bundler"

function main() {
  local version output_dir target temp_dir tarball_name
  version="${1}"
  output_dir="${2}"
  target="${3}"

  echo "version=${version}"
  echo "output_dir=${output_dir}"
  echo "target=${target}"

  temp_dir="$(mktemp -d)"
  pwd="$(pwd)"

  pushd "${temp_dir}" > /dev/null
    unset RUBYOPT; \
      GEM_HOME="${temp_dir}" \
      GEM_PATH="${temp_dir}" \
      gem install bundler \
        --version "${version}" \
        --no-document \
        --env-shebang

    rm -f "bundler-${version}.gem"
    rm -rf "cache/bundler-${version}.gem"
    sed -i.bak 's/#!.*ruby.*/#!\/usr\/bin\/env ruby/g' bin/*
    rm bin/*.bak
    tarball_name="bundler-${version}-${target}.tgz"

    if [[ "$output_dir" != /* ]]
    then
      output_dir="$pwd/$output_dir"
    fi

    tar czvf "$output_dir/$tarball_name" .
  popd > /dev/null

  pushd "${output_dir}" > /dev/null
    sha256sum "${tarball_name}" > "${tarball_name}.sha256"
  popd > /dev/null

  rm -rf "${temp_dir}"
}

main "${@:-}"
