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
# hhios/go1 is the central hhios FIPS Go image. It sets GOFIPS140 and
# GOTOOLCHAIN=local, so go.mod's "go" line must be satisfied by the image's Go.
FROM harness0.harness.io/oci/docker_artifacts/hhios/go1:1.26-fips AS builder

ENV GOFLAGS="-mod=readonly"

# The hhios image runs as uid 65534; root is needed to create /workspace and
# to write the Go module cache. Only the builder stage runs as root.
USER root

RUN mkdir -p /workspace
WORKDIR /workspace

ARG GOPROXY

COPY go.* /workspace/
RUN go mod download

COPY Makefile main-targets.mk /workspace/

#COPY --from=frontend /web/dist/web /workspace/web/dist/web
COPY . /workspace

# Rewrite the config path here so the runtime stage needs no RUN step (and no root).
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
# The runtime stage is also hhios/go1 because no generic hhios FIPS OS runtime
# image is published. The image already sets
# ENTRYPOINT ["/usr/local/bin/fips-entrypoint.sh"], which turns FIPS mode on or
# off from FIPS_ENABLED and then execs CMD. Do not set an ENTRYPOINT here (it
# would replace the baked-in one) and do not copy fips-entrypoint.sh.
FROM harness0.harness.io/oci/docker_artifacts/hhios/go1:1.26-fips

# The base image may set its own WORKDIR. cloudinfo discovers its config.toml via
# viper's relative "./config" search path, which only resolves to the mounted
# /config when the working directory is "/". Pin it explicitly so the mounted
# configuration is found regardless of the base image default.
WORKDIR /

# Nothing below needs root: the binary and configs are copied in as root-owned,
# world-readable files, and cloudinfo only has to read them (plus the mounted
# /config). Run as the hhios default non-root user (the previous runtime stage
# ran as root).
USER 65534

ARG BUILD_TARGET

COPY --from=builder /build/* /usr/local/bin/

COPY --from=builder /workspace/configs /etc/cloudinfo/serviceconfig

ENV CLOUDINFO_SERVICELOADER_SERVICECONFIGLOCATION "/etc/cloudinfo/serviceconfig"

CMD ["cloudinfo"]
