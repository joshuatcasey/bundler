#!/usr/bin/env sh

set -u

image="${1}"

case "${image}" in
  "paketobuildpacks/build:full")
    echo "no preparation neeeded"
    ;;

  "ubuntu:22.04")
    apt-get update
    apt-get install -y jq ruby-full make
    ;;

  "alpine")
    apk update
    apk add jq ruby make
    ;;

  "macos")
    brew install coreutils
    ;;

  *)
    echo "no preparation neeeded"
    ;;
esac