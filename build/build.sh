#!/bin/bash
set -e

# Get the directory where the script is located and its parent directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
TARGET_DIR="$PARENT_DIR/target"

# Create output directory
mkdir -p $TARGET_DIR

# Get build information
VERSION=${VERSION:-$(git describe --tags --always --dirty)}
COMMIT_HASH=${COMMIT_HASH:-$(git rev-parse --short HEAD)}
BUILD_DATE=${BUILD_DATE:-$(date -u '+%Y-%m-%d %H:%M:%S UTC')}

# Remove any previous build artifacts
rm -rf $TARGET_DIR/*

# Define platforms to build for
platforms=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

# Build Docker image
echo "Building Docker image for cross-compilation..."
docker build --target builder -t qbtr-builder -f $SCRIPT_DIR/Dockerfile

# Compile for each platform
for platform in "${platforms[@]}"; do
  # Parse the platform string
  os=$(echo $platform | cut -d/ -f1)
  arch=$(echo $platform | cut -d/ -f2)
  variant=""
  
  # Handle ARM variant
  if [[ $platform == */* && $platform != */v* ]]; then
    variant=""
  elif [[ $platform == */v* ]]; then
    variant=$(echo $platform | cut -d/ -f3)
  fi
  
  echo "Building for $os/$arch$variant..."
  
  # Create container name
  container_name="qbtr-build-$os-$arch$variant"
  
  # Build the binary using Docker
  docker run --rm --name $container_name \
    -v "$PARENT_DIR/target:/output" \
    -e GOOS=$os \
    -e GOARCH=$arch \
    -e GOARM=${variant#v} \
    qbtr-builder \
    go build -o "/output/qbtr-$os-$arch$variant" -ldflags="-s -w -X 'main.Version=$VERSION' -X 'main.CommitHash=$COMMIT_HASH' -X 'main.BuildDate=$BUILD_DATE'" .
  
  # Add extension for Windows binaries
  if [ "$os" = "windows" ]; then
    mv "$TARGET_DIR/qbtr-$os-$arch$variant" "$TARGET_DIR/qbtr-$os-$arch$variant.exe"
  fi
  
  echo "Done building for $os/$arch$variant"
done

# Create checksums
echo "Generating checksums..."
cd $TARGET_DIR
if [ "$(uname)" = "Darwin" ]; then
  # macOS version
  for file in *; do
    shasum -a 256 "$file" >> checksums.txt
  done
else
  # Linux version
  for file in *; do
    sha256sum "$file" >> checksums.txt
  done
fi
cd ..

echo "Build completed successfully!"
echo "Binary files are located in the 'build' directory." 