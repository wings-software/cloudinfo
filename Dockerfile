# UI build image
# Disabling UI build as it's not used.

#FROM node:16.5.0 as frontend
#
#WORKDIR /web
#
#COPY web/package.json web/package-lock.json /web/
#
#RUN npm install --legacy-peer-deps
#
#COPY web/ /web/
#RUN npm run build-prod


# Build image
#FROM golang:1.19.9-buster AS builder
FROM us.gcr.io/platform-205701/ubi/ubi-go:8.8 AS builder

ENV GOFLAGS="-mod=readonly"

#RUN apk add --update --no-cache ca-certificates make git curl mercurial
#RUN apt-get update && apt-get install -y ca-certificates make git curl mercurial

USER root
RUN microdnf update && microdnf install -y ca-certificates make git curl && microdnf clean all


RUN mkdir -p /workspace
WORKDIR /workspace

ARG GOPROXY

COPY go.* /workspace/
RUN go mod download

COPY Makefile main-targets.mk /workspace/

#COPY --from=frontend /web/dist/web /workspace/web/dist/web
COPY . /workspace

ARG BUILD_TARGET

RUN set -xe && \
    if [[ "${BUILD_TARGET}" == "debug" ]]; then \
        cd /tmp; GOBIN=/workspace/build/debug go get github.com/go-delve/delve/cmd/dlv; cd -; \
        make build-debug; \
        mv build/debug /build; \
    else \
        make build-release; \
        mv build/release /build; \
    fi


# Final image
# FROM alpine:3.14.0
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
USER root

# RUN apk add --update --no-cache ca-certificates tzdata bash curl

SHELL ["/bin/bash", "-c"]

# set up nsswitch.conf for Go's "netgo" implementation
# https://github.com/gliderlabs/docker-alpine/issues/367#issuecomment-424546457
# RUN test ! -e /etc/nsswitch.conf && echo 'hosts: files dns' > /etc/nsswitch.conf

ARG BUILD_TARGET

RUN if [[ "${BUILD_TARGET}" == "debug" ]]; then apk add --update --no-cache libc6-compat; fi

COPY --from=builder /build/* /usr/local/bin/

COPY configs /etc/cloudinfo/serviceconfig

RUN sed -i "s|dataLocation: ./configs/|dataLocation: /etc/cloudinfo/serviceconfig/|g" /etc/cloudinfo/serviceconfig/services.yaml

ENV CLOUDINFO_SERVICELOADER_SERVICECONFIGLOCATION "/etc/cloudinfo/serviceconfig"

CMD ["cloudinfo"]
