FROM --platform=$TARGETPLATFORM golang:alpine

ARG TARGETOS TARGETARCH BUILDPLATFORM TARGETPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"
RUN echo "TARGETOS = $TARGETOS"
RUN echo "TARGETARCH = $TARGETARCH"

LABEL org.opencontainers.image.source=https://github.com/openreserveio/core/core-ledger-poster
LABEL org.opencontainers.image.description="OpenReserve core ledger poster"
LABEL org.opencontainers.image.licenses="Proprietary"

RUN mkdir -p /opt/core && \
    mkdir -p /etc/core/src

COPY . /opt/core/src/core-ledger-poster

RUN cd /opt/core/src/core-ledger-poster && go mod download
RUN cd /opt/core/src/core-ledger-poster/application/coreledgerposter && env GOOS=$TARGETOS GOARCH=$TARGETARCH go build && go install
RUN rm -Rf /opt/core/src


CMD ["coreledgerposter", "start"]
