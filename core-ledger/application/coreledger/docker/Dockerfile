FROM --platform=$TARGETPLATFORM golang:alpine

ARG TARGETOS TARGETARCH BUILDPLATFORM TARGETPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"
RUN echo "TARGETOS = $TARGETOS"
RUN echo "TARGETARCH = $TARGETARCH"

LABEL org.opencontainers.image.source=https://github.com/openreserveio/core/core-ledger
LABEL org.opencontainers.image.description="openreserve core ledger"
LABEL org.opencontainers.image.licenses="Proprietary"

RUN mkdir -p /opt/core && \
    mkdir -p /etc/core/src

COPY . /opt/core/src/core-ledger

RUN cd /opt/core/src/core-ledger && go mod download
RUN cd /opt/core/src/core-ledger/application/coreledger && env GOOS=$TARGETOS GOARCH=$TARGETARCH go build && go install
RUN rm -Rf /opt/core/src


CMD ["coreledger", "start"]
