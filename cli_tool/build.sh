#!/bin/bash
APP_NAME="boardshapes-cli"
OUTPUT_DIR="release"
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
  "js/wasm"
)
VERSION=$1

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

mkdir -p $OUTPUT_DIR

for PLATFORM in "${PLATFORMS[@]}"
do
  IFS="/" read -r GOOS GOARCH <<< "$PLATFORM"
  OUTPUT_NAME="$APP_NAME-$VERSION-$GOOS-$GOARCH"
  [ "$GOOS" = "windows" ] && OUTPUT_NAME+=".exe"
  [ "$GOOS" = "js" ] && OUTPUT_NAME="$APP_NAME-$VERSION.wasm"
  env GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT_DIR/$OUTPUT_NAME"
done