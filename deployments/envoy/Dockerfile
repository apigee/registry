ARG ENVOY_VERSION=v1.20-latest
ARG GO_VERSION=1.18

FROM golang:${GO_VERSION} as builder

WORKDIR /app

RUN apt-get update -y && apt-get install curl unzip git make -y
COPY . .
RUN ./tools/FETCH-PROTOC.sh
RUN make protos


FROM envoyproxy/envoy:${ENVOY_VERSION}
ENV REGISTRY_SERVER_HOST=127.0.0.1
ENV REGISTRY_SERVER_PORT=8080
ENV PORT=9999

RUN apt-get update -y && apt-get install gettext -y

COPY deployments/envoy/envoy.yaml.tmpl /etc/envoy/envoy.yaml.tmpl

COPY --from=builder /app/deployments/envoy/proto.pb /proto.pb


RUN echo "#!/bin/sh" > /startup.sh && \
    echo "set -e" >> /startup.sh && \
    echo "envsubst < /etc/envoy/envoy.yaml.tmpl > /etc/envoy/envoy.yaml" >> /startup.sh && \
    echo "envoy -c /etc/envoy/envoy.yaml" >> /startup.sh && \
    chmod +x /startup.sh

ENTRYPOINT ["/startup.sh"]