#!/bin/sh
# Build a Linux x86_64 SDL2 binary using a native ARM64 Go toolchain on Apple Silicon.
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
docker build --platform linux/arm64 -f "$ROOT/tools/docker/Dockerfile.cross" -t shamogu-zh-cross:local "$ROOT"
mkdir -p "$ROOT/output/linux"
docker run --rm --platform linux/arm64 -v "$ROOT:/src" shamogu-zh-cross:local go build -mod=vendor -tags sdl -o output/linux/shamogu-zh
