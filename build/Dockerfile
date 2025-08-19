FROM golang:1.23.2 AS build

WORKDIR /cloud-gateway

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . ./

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/gateway

FROM alpine:latest

WORKDIR /cloud-gateway

# Copy the binary from the build stage
COPY --from=build /cloud-gateway/main /cloud-gateway/main

# Add a non-root user and group to the container
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

# Switch to the non-root user
USER nonroot

CMD ["/cloud-gateway/main"]