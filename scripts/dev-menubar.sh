#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DERIVED_DATA="${DERIVED_DATA:-$ROOT_DIR/.build/DerivedData}"
CONFIG="${CONFIG:-Debug}"
SHOW_LOGS=0
SKIP_BUILD=0
SKIP_CLI=0
CCW_BIN="${CCW_BIN:-$ROOT_DIR/bin/ccw}"

usage() {
  cat <<'USAGE'
Usage: scripts/dev-menubar.sh [--release] [--no-build] [--logs]

Env:
  DERIVED_DATA  Derived data path (default: <repo>/.build/DerivedData)
  CONFIG        Build configuration (default: Debug)

Flags:
  --release     Build with Release configuration
  --no-build    Skip the build and just relaunch the app
  --no-cli      Skip building the ccw CLI binary
  --logs        Stream app logs after launch (Ctrl+C to stop)
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --release)
      CONFIG="Release"
      shift
      ;;
    --no-build)
      SKIP_BUILD=1
      shift
      ;;
    --no-cli)
      SKIP_CLI=1
      shift
      ;;
    --logs)
      SHOW_LOGS=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown flag: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ $SKIP_CLI -eq 0 ]]; then
  mkdir -p "$(dirname "$CCW_BIN")"
  go build -o "$CCW_BIN" .
fi

if [[ $SKIP_BUILD -eq 0 ]]; then
  xcodebuild \
    -workspace "$ROOT_DIR/CCWMenubar.xcworkspace" \
    -scheme CCWMenubar \
    -configuration "$CONFIG" \
    -derivedDataPath "$DERIVED_DATA" \
    build
fi

APP_PATH="$DERIVED_DATA/Build/Products/$CONFIG/CCWMenubar.app"
if [[ ! -d "$APP_PATH" ]]; then
  echo "App not found at $APP_PATH" >&2
  exit 1
fi

pkill -x CCWMenubar >/dev/null 2>&1 || true
if [[ -x "$CCW_BIN" ]]; then
  CCW_BIN_PATH="$CCW_BIN" "$APP_PATH/Contents/MacOS/CCWMenubar" &
else
  if ! open "$APP_PATH"; then
    "$APP_PATH/Contents/MacOS/CCWMenubar" &
  fi
fi

if [[ $SHOW_LOGS -eq 1 ]]; then
  /usr/bin/log stream --style compact \
    --predicate 'process == "CCWMenubar" || subsystem == "com.justanotheratom.ccw-menubar"' \
    --info --level info
fi
