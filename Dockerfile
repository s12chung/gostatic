FROM golang:1.10.3-alpine3.8

ARG DOCKER_WORKDIR

RUN apk --no-cache add\
    git make dep

WORKDIR $DOCKER_WORKDIR
COPY . .

RUN make install
