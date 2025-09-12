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

mkdir -p $OUTPUT_DIR

for PLATFORM in "${PLATFORMS[@]}"
do
  IFS="/" read -r GOOS GOARCH <<< "$PLATFORM"
  OUTPUT_NAME="$APP_NAME-$GOOS-$GOARCH"
  [ "$GOOS" = "windows" ] && OUTPUT_NAME+=".exe"
  [ "$GOOS" = "js" ] && OUTPUT_NAME="$APP_NAME.wasm"
  env GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT_DIR/$OUTPUT_NAME"
done