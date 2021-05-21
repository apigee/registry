# This Dockerfile builds an image that runs the registry-server with no proxy.

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

# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/registry-server /registry-server

# Copy the standard registry configuration from the builder image.
COPY config/registry.yaml /registry.yaml

# Run the web service on container startup.
CMD ["/registry-server", "-c", "registry.yaml"]
