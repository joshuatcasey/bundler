FROM ubuntu:22.04

RUN apt-get update && apt-get -y install ruby-full

COPY ../../../compilation/compile.sh /entrypoint

ENTRYPOINT ["/entrypoint"]
