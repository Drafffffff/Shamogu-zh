#!/bin/sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
HOST=${SHAMOGU_SSH_HOST:-bazzite}
REMOTE=Games/shamogu-zh
rsync -az --exclude=.git --exclude=output --exclude=verification "$ROOT/" "$HOST:$REMOTE/source/"
# Existing Fedora build environment; system image is not changed.
ssh "$HOST" 'podman start dcss-zh-builder >/dev/null; podman exec dcss-zh-builder mkdir -p /work/shamogu-zh'
rsync -az --exclude=.git --exclude=output --exclude=verification "$ROOT/" "$HOST:Games/dcss-zh/shamogu-zh/"
ssh "$HOST" 'set -e; podman exec -w /work/shamogu-zh dcss-zh-builder sh -c "go test -mod=vendor codeberg.org/anaseto/gruid && go test -mod=vendor codeberg.org/anaseto/gruid/ui -run TestReplayVariableGrid && go test -mod=vendor ./... && go test -mod=vendor -tags sdl ./... && go build -mod=vendor -tags sdl -o shamogu-zh"; mkdir -p ~/Games/shamogu-zh/game; cp ~/Games/dcss-zh/shamogu-zh/shamogu-zh ~/Games/shamogu-zh/game/shamogu-zh.new; mv ~/Games/shamogu-zh/game/shamogu-zh.new ~/Games/shamogu-zh/game/shamogu-zh'
mkdir -p "$ROOT/output/linux"
rsync -az "$HOST:$REMOTE/game/shamogu-zh" "$ROOT/output/linux/"
