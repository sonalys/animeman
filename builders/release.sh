#!/bin/bash

# Check if version is provided
if [ -z "$1" ]; then
  echo "No version provided. Usage: ./build.sh <version>"
  exit 1
fi

VERSION=$1

# Define the build targets
declare -A targets=(
  ["windows/amd64"]="animeman.exe"
  ["linux/amd64"]="animeman"
  ["linux/arm64"]="animeman"
)

# Build for each target
for target in "${!targets[@]}"; do
  OS=$(echo $target | cut -d '/' -f 1)
  ARCH=$(echo $target | cut -d '/' -f 2)
  OUTPUT="./bin/${target}/${targets[$target]}"

  echo "Building for $OS/$ARCH..."
  GOOS=$OS GOARCH=$ARCH go build -ldflags="-X 'main.version=${VERSION}'" -o $OUTPUT ./cmd/service/main.go
done

# Create releases directory
mkdir -p releases

# Package the binaries
for target in "${!targets[@]}"; do
  OS=$(echo $target | cut -d '/' -f 1)
  ARCH=$(echo $target | cut -d '/' -f 2)
  OUTPUT="./bin/${target}/${targets[$target]}"
  ZIP_NAME="releases/animeman_${OS}_${ARCH}.zip"

  echo "Packaging $OUTPUT into $ZIP_NAME..."
  zip -j $ZIP_NAME $OUTPUT
done

echo "Build and packaging completed with version ${VERSION}"
