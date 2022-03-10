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


# Build this image using "envsubst < containers/registry-tools/registry-linters/Dockerfile | docker build -t gcr.io/$REGISTRY_PROJECT_IDENTIFIER/registry-linters -f - ."
# docker push gcr.io/$REGISTRY_PROJECT_IDENTIFIER/registry-linters:latest
# Make sure REGISTRY_PROJECT_IDENTIFIER is set

# Build stage
ARG REGISTRY_TOOL_IMAGE="ghcr.io/apigee/registry-tools:latest"

FROM golang:alpine as builder

# Copy some commonly-needed tools into the image.
RUN apk update && apk upgrade
RUN apk add --no-cache bash git

#Install api-linter
WORKDIR /api-linter
RUN GO111MODULE=on go get github.com/googleapis/api-linter/cmd/api-linter

# Install gnostic linter
WORKDIR /gnostic-linter
RUN git clone https://github.com/google/gnostic
RUN cd gnostic && go install ./...

# Final image
FROM $REGISTRY_TOOL_IMAGE
ARG REGISTRY_TOOL_IMAGE

# Install spectral linter
RUN apk add --no-cache nodejs npm
RUN npm install -g @stoplight/spectral@5.9.2

# Copy other linter binaries
COPY --from=builder /go/bin/api-linter /bin/
COPY --from=builder /go/bin/gnostic* /bin/
