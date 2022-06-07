#!/usr/bin/env bash

set -u
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function hasParams() {
    local tarball_path version
    tarball_path="${1}"
    version="${2}"

    if [[ -z "${tarball_path}" ]]; then
      echo " ⛔ specify tarball_path as the first parameter"
      exit 1
    fi

    if [[ -z "${version}" ]]; then
      echo " ⛔ specify tarball_path as the first parameter"
      exit 1
    fi
}

function itExists() {
  echo -n "tarball exists"

  local tarball_path
  tarball_path="${1}"

  if [[ -z "${tarball_path}" ]]; then
    echo " ⛔ specify tarball_path as the first parameter"
    exit 1
  fi

  if [[ ! -f "${tarball_path}" ]]; then
    echo " ⛔ ${tarball_path} does not exist"
    exit 1
  fi

  echo "... ✅"
}

function itHasTheRightShebang() {
  echo -n "bin/* files have shebang of ''#!/usr/bin/env ruby'"

  local tarball_path
  tarball_path="${1}"

  local shebang

  tar -O -xf "${tarball_path}" ./bin/bundle
  tar -O -xf "${tarball_path}" ./bin/bundler

  shebang=$(tar -O -xf "${tarball_path}" ./bin/bundle | head -n1)
  if [[ "${shebang}" != "#!/usr/bin/env ruby" ]]; then
    echo " ⛔ bin/bundle must have shebang of '#!/usr/bin/env ruby'"
    echo "shebang=${shebang}"
    exit 1
  fi

  shebang=$(tar -O -xf "${tarball_path}" ./bin/bundler | head -n1)
  if [[ "${shebang}" != "#!/usr/bin/env ruby" ]]; then
    echo " ⛔ bin/bundler must have shebang of '#!/usr/bin/env ruby'"
    echo "shebang=${shebang}"
    exit 1
  fi

  echo "... ✅"
}

function itHasTheRightVersion() {
  echo -n "'bin/bundle -v' prints the right version"

  local tarball_path version full_tarball_path
  tarball_path="${1}"
  version="${2}"
  full_tarball_path=$(realpath "${tarball_path}")
  temp_dir="$(mktemp -d)"

  pushd "${temp_dir}" > /dev/null || exit 1
    tar xzf "${full_tarball_path}"

    local output status

    output=$(GEM_HOME="${temp_dir}" \
      GEM_PATH="${temp_dir}" \
      ./bin/bundle -v)

    echo "${output}" | grep "Bundler version ${version}" > /dev/null
    status=$?

    if [[ $status -ne 0 ]]; then
      echo " ⛔ 'bin/bundle -v' expected 'Bundler version ${version}'"
      exit 1
    fi

  popd > /dev/null || exit 1

  echo "... ✅"
}

function main(){
  hasParams "${@:-}"
  echo -n "1. " && itExists "${@:-}"
  echo -n "2. " && itHasTheRightShebang "${@:-}"
  echo -n "3. " && itHasTheRightVersion "${@:-}"
}

main "${@:-}"

# https://github.com/paketo-buildpacks/dep-server/blob/main/actions/test-dependency/dependency-tests/bundler/run