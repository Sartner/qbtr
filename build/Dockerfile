# This docker file is used to build the qbtr binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Version information
ARG VERSION="dev"
ARG BUILD_DATE="unknown"
ARG COMMIT_HASH="unknown"

# Install required dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Pre-copy dependencies to cache them in docker layer
COPY ../go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY "../cmd" .

# Build argument for target platform
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# Build the application with version information
RUN mkdir -p /app/build && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -o /app/build/qbtr-${TARGETOS}-${TARGETARCH}${TARGETVARIANT} \
    -ldflags="-s -w \
    -X 'main.Version=${VERSION}' \
    -X 'main.BuildDate=${BUILD_DATE}' \
    -X 'main.CommitHash=${COMMIT_HASH}'" \
    .

# Using a new stage for a clean output
FROM scratch AS export-stage
COPY --from=builder /app/build / 