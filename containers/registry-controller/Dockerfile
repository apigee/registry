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

FROM golang:1.17 as builder
RUN apt-get update \
    && apt-get install -y unzip git

# Create and change to the app directory.
WORKDIR /app

# Copy local code to the container image.
COPY . ./
RUN ./tools/FETCH-PROTOC.sh

# Retrieve application dependencies
RUN go mod download

# Build registry tool.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o registry ./cmd/registry

# Build spectral plugin
RUN go build -o registry-lint-spectral ./cmd/registry/plugins/registry-lint-spectral

# Install spectral linter
WORKDIR /spectral-linter
RUN wget -q https://github.com/stoplightio/spectral/releases/download/v5.9.2/spectral-linux \
               -O spectral \
         && chmod 755 spectral

# Prepare bash dependencies for the final image.
RUN apt-get -y install wget bash-static \
    && wget -q http://landley.net/toybox/bin/toybox-x86_64 \
               -O /usr/local/bin/toybox \
           && chmod 755 /usr/local/bin/toybox \
           && cd /usr/local/bin; for i in $(./toybox); do ln -s toybox $i; done

RUN wget -q https://busybox.net/downloads/binaries/1.31.0-i686-uclibc/busybox_WGET \
        -O /usr/local/bin/wget \
    && chmod a+x /usr/local/bin/wget

# Final production image
FROM gcr.io/distroless/cc-debian10:latest

# Copy bash dependencies.
COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /bin/bash-static /bin/bash
RUN [ "/usr/local/bin/toybox", "ln", "-s", "/bin/bash", "/bin/sh" ]
RUN chmod -R 755 /usr/local/bin

# Copy tool binaries to the production image
COPY --from=builder /app/registry /bin/registry

# Copy linter plugin
COPY --from=builder /app/registry-lint-spectral /bin/registry-lint-spectral

# Copy linter binaries
COPY --from=builder /spectral-linter/spectral /bin/
