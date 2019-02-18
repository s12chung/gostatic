FROM golang:1.11-alpine3.9

ARG DOCKER_WORKDIR
ENV GO111MODULE on

RUN apk --no-cache add\
    git \
    make \
    # for golangci-lint linters
    build-base

RUN wget -O - -q https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.15.0

WORKDIR $DOCKER_WORKDIR
COPY . .

RUN make install
