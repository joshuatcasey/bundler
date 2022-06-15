FROM alpine

RUN apk update && apk add ruby

COPY ../../compilation/compile.sh /entrypoint

ENTRYPOINT ["/entrypoint"]
