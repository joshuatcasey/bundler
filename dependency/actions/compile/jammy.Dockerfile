FROM ubuntu:22.04

RUN apt-get update
RUN apt-get install -y jq ruby-full make

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
