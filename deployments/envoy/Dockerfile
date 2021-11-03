ARG PROTOBUF_VERSION=3.18.1
ARG ENVOY_VERSION=v1.20-latest
ARG GO_VERSION=1.17

FROM golang:${GO_VERSION} as builder
ARG PROTOBUF_VERSION

WORKDIR /app

RUN apt-get update -y && apt-get install curl unzip git make -y && \
    PB_REL="https://github.com/protocolbuffers/protobuf/releases" && \
    curl -LO $PB_REL/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOBUF_VERSION}-linux-x86_64.zip -d /usr/local
COPY . .
RUN make protos


FROM envoyproxy/envoy:${ENVOY_VERSION}
ENV REGISTRY_SERVER_HOST=127.0.0.1
ENV REGISTRY_SERVER_PORT=8080

RUN apt-get update -y && apt-get install gettext -y

COPY deployments/envoy/envoy.yaml.tmpl /etc/envoy/envoy.yaml.tmpl

COPY --from=builder /app/deployments/envoy/proto.pb /proto.pb


RUN echo "#!/bin/sh" > /startup.sh && \
    echo "set -e" >> /startup.sh && \
    echo "envsubst < /etc/envoy/envoy.yaml.tmpl > /etc/envoy/envoy.yaml" >> /startup.sh && \
    echo "envoy -c /etc/envoy/envoy.yaml" >> /startup.sh && \
    chmod +x /startup.sh

ENTRYPOINT ["/startup.sh"]