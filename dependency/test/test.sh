#!/usr/bin/env bash

set -u
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function hasParams() {
  local tarballPath version
  tarballPath="${1}"
  version="${2}"

  if [[ -z "${tarballPath}" ]]; then
    echo " ⛔ specify tarballPath as the first parameter"
    exit 1
  fi

  if [[ -z "${version}" ]]; then
    echo " ⛔ specify tarballPath as the first parameter"
    exit 1
  fi
}

function itExists() {
  echo -n "tarball exists"

  local tarballPath
  tarballPath="${1}"

  if [[ -z "${tarballPath}" ]]; then
    echo " ⛔ specify tarballPath as the first parameter"
    exit 1
  fi

  if [[ ! -f "${tarballPath}" ]]; then
    echo " ⛔ ${tarballPath} does not exist"
    exit 1
  fi

  echo "... ✅"
}

function itHasTheRightShebang() {
  echo -n "bin/* files have shebang of '#!/usr/bin/env ruby'"

  local tarballPath
  tarballPath="${1}"

  local shebang

  shebang=$(tar -O -xf "${tarballPath}" ./bin/bundle | head -n1)
  if [[ "${shebang}" != "#!/usr/bin/env ruby" ]]; then
    echo " ⛔ bin/bundle must have shebang of '#!/usr/bin/env ruby'"
    echo "shebang=${shebang}"
    exit 1
  fi

  shebang=$(tar -O -xf "${tarballPath}" ./bin/bundler | head -n1)
  if [[ "${shebang}" != "#!/usr/bin/env ruby" ]]; then
    echo " ⛔ bin/bundler must have shebang of '#!/usr/bin/env ruby'"
    echo "shebang=${shebang}"
    exit 1
  fi

  echo "... ✅"
}

function itHasTheRightVersion() {
  echo -n "'bin/bundle -v' prints the right version"

  local tarballPath version full_tarballPath
  tarballPath="${1}"
  version="${2}"

  full_tarballPath=$(realpath "${tarballPath}")
  temp_dir="$(mktemp -d)"

  pushd "${temp_dir}" > /dev/null || exit 1
    tar xzf "${full_tarballPath}"

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
