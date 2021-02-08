# This Dockerfile builds an image that runs the registry-server with no proxy.

# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.15 as builder
RUN apt-get update
RUN apt-get install unzip
RUN curl -L https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-linux-x86_64.zip > /tmp/protoc.zip
RUN unzip /tmp/protoc.zip -d /usr/local
# Create and change to the app directory.
WORKDIR /app
# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download
# Copy local code to the container image.
COPY . ./
# Compile protos.
RUN make protos
# Build app-server.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o app-server ./cmd/app-server

# Build linter tool binaries
FROM golang:1.15-alpine as linter
RUN apk update && apk upgrade
RUN apk add --no-cache git
WORKDIR /linter
RUN git clone https://github.com/google/gnostic
WORKDIR /linter/gnostic
RUN go install ./...

# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3
RUN apk add --no-cache ca-certificates
# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/app-server /app-server
# Copy the linter binaries
COPY --from=builder /go/bin/api-linter /bin/
COPY --from=linter /go/bin/* /bin/
ENV PATH=$PATH:/bin/
ARG registry_endpoint
ENV APG_REGISTRY_ADDRESS=$registry_endpoint
CMD ["/app-server"]
