# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.13 as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o registry-server ./cmd/registry-server

# Use an Envoy release image to get envoy in the image.
# This is the last version that supports allow_origin in CorsPolicy
# https://www.envoyproxy.io/docs/envoy/latest/version_history/v1.12.0
FROM envoyproxy/envoy:v1.11.2

COPY container/envoy.yaml /etc/envoy/envoy.yaml
COPY container/RUN.sh /RUN.sh

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/registry-server /registry-server

# Run the web service on container startup.
CMD ["/RUN.sh"]
