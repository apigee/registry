# Copyright 2021 Google LLC
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

# This Dockerfile builds an image that runs the registry-server behind an
# included Envoy proxy.

# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.15 as builder

RUN apt-get update
RUN apt-get install unzip

# Create and change to the app directory.
WORKDIR /app

# Get protoc
COPY ./tools/FETCH-PROTOC.sh ./
RUN ./FETCH-PROTOC.sh

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Compile protos.
RUN make protos

# Build registry-server.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o registry-server ./cmd/registry-server

# Build authz-server.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o authz-server ./cmd/authz-server

# Use an Envoy release image to get envoy in the image.
FROM envoyproxy/envoy:v1.16.0

COPY containers/registry-bundle/RUN.sh /RUN.sh
COPY containers/registry-bundle/envoy.yaml /etc/envoy/envoy.yaml
COPY --from=builder /app/deployments/envoy/proto.pb /proto.pb

# Copy the registry-server binary to the production image from the builder stage.
COPY --from=builder /app/registry-server /registry-server

# Copy the authz-server binary to the production image from the builder stage.
COPY --from=builder /app/authz-server /authz-server

# Copy configuration files to the production image.
COPY config/registry-server.yaml /registry-server.yaml
COPY cmd/authz-server/authz.yaml /authz.yaml

# Run as root in the container. Needed to use the Envoy release image.
ENV ENVOY_UID=0

# Run services on container startup.
CMD ["/RUN.sh"]
