# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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