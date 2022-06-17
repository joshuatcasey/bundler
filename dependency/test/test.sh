#!/usr/bin/env bash

set -u
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function hasParams() {
  if [ "$#" -ne 2 ]; then
      echo "Test requires 2 positional params: tarball_name version"
  fi

  local tarball_name version
  tarball_name="${1}"
  version="${2}"

  if [[ -z "${tarball_name}" ]]; then
    echo " ⛔ specify tarball_name as the first parameter"
    exit 1
  fi

  if [[ -z "${version}" ]]; then
    echo " ⛔ specify tarball_name as the first parameter"
    exit 1
  fi
}

function itExists() {
  echo -n "tarball exists"

  local tarball_name
  tarball_name="${1}"

  if [[ -z "${tarball_name}" ]]; then
    echo " ⛔ specify tarball_name as the first parameter"
    exit 1
  fi

  if [[ ! -f "${tarball_name}" ]]; then
    echo " ⛔ ${tarball_name} does not exist"
    exit 1
  fi

  echo "... ✅"
}

function itHasTheRightShebang() {
  echo -n "bin/* files have shebang of '#!/usr/bin/env ruby'"

  local tarball_name
  tarball_name="${1}"

  local shebang

  shebang=$(tar -O -xf "${tarball_name}" ./bin/bundle | head -n1)
  if [[ "${shebang}" != "#!/usr/bin/env ruby" ]]; then
    echo " ⛔ bin/bundle must have shebang of '#!/usr/bin/env ruby'"
    echo "shebang=${shebang}"
    exit 1
  fi

  shebang=$(tar -O -xf "${tarball_name}" ./bin/bundler | head -n1)
  if [[ "${shebang}" != "#!/usr/bin/env ruby" ]]; then
    echo " ⛔ bin/bundler must have shebang of '#!/usr/bin/env ruby'"
    echo "shebang=${shebang}"
    exit 1
  fi

  echo "... ✅"
}

function itHasTheRightVersion() {
  echo -n "'bin/bundle -v' prints the right version"

  local tarball_name version full_tarball_name
  tarball_name="${1}"
  version="${2}"

  full_tarball_name=$(realpath "${tarball_name}")
  temp_dir="$(mktemp -d)"

  pushd "${temp_dir}" > /dev/null || exit 1
    tar xzf "${full_tarball_name}"

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

  rm -rf "${temp_dir}"

  echo "... ✅"
}

function main(){
  hasParams "${@:-}"
  echo -n "1. " && itExists "${@:-}"
  echo -n "2. " && itHasTheRightShebang "${@:-}"
  echo -n "3. " && itHasTheRightVersion "${@:-}"
}

main "${@:-}"
