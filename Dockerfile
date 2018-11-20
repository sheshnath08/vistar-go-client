FROM golang:1.11

RUN apt-get update && \
  apt-get install ca-certificates openssl curl tzdata wget git make vim \
  protobuf-compiler -y

RUN adduser --group cortex
RUN adduser --home /go --shell /bin/bash --quiet --ingroup cortex \
    --disabled-login --disabled-password --gecos "" cortex

RUN chown -R cortex.cortex /go

USER cortex

WORKDIR /go/src/vistar-ad-client

RUN go get github.com/kyoh86/richgo
