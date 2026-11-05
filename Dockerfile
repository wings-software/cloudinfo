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
FROM harness0.harness.io/oci/docker_artifacts/hhi/go-fips:1.26-dev AS builder

ENV GOFLAGS="-mod=readonly"

USER root

RUN mkdir -p /workspace
WORKDIR /workspace

ARG GOPROXY

COPY go.* /workspace/
RUN go mod download

COPY Makefile main-targets.mk /workspace/

#COPY --from=frontend /web/dist/web /workspace/web/dist/web
COPY . /workspace

# Rewrite config path before copy into the runtime image (runtime tag may lack sed/shell tooling intent).
RUN sed -i "s|dataLocation: ./configs/|dataLocation: /etc/cloudinfo/serviceconfig/|g" \
      /workspace/configs/services.yaml

ARG BUILD_TARGET

RUN set -xe && \
    if [ "${BUILD_TARGET}" = "debug" ]; then \
        cd /tmp; GOBIN=/workspace/build/debug go get github.com/go-delve/delve/cmd/dlv; cd -; \
        make build-debug; \
        mv build/debug /build; \
    else \
        make build-release; \
        mv build/release /build; \
    fi


# Final image
FROM harness0.harness.io/oci/docker_artifacts/hhi/go-fips:1.26
USER root

# The curated base image may set a non-root WORKDIR. cloudinfo discovers its
# config.toml via viper's relative "./config" search path, which only resolves
# to the mounted /config when the working directory is "/". Pin it explicitly so
# the mounted configuration is found regardless of the base image default.
WORKDIR /

ARG BUILD_TARGET

COPY --from=builder /build/* /usr/local/bin/

COPY --from=builder /workspace/configs /etc/cloudinfo/serviceconfig

ENV CLOUDINFO_SERVICELOADER_SERVICECONFIGLOCATION "/etc/cloudinfo/serviceconfig"

CMD ["cloudinfo"]
