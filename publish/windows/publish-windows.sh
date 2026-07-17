#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/publish/output"
STAGING_ROOT="$ROOT_DIR/publish/staging/windows"
ARCH=""
VERSION=""
SKIP_BUILD=0
SKIP_INSTALLER=0
KEEP_STAGING=0

usage() {
  cat <<'EOF'
Usage:
  publish/windows/publish-windows.sh --arch <amd64|arm64> [options]

Options:
  --arch <amd64|arm64>   Target architecture (required)
  --version <ver>        Package version (default: read from wails.json)
  --skip-build           Skip frontend and Wails build steps
  --skip-installer       Only produce the portable zip, skip the NSIS installer
  --keep-staging         Keep staging directory after packaging
  -h, --help             Show help

Notes:
  - Run on a native Windows host (Git Bash / MSYS) so Wails can build the .exe.
  - The NSIS installer step needs makensis in PATH (choco install nsis).
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --arch)
      ARCH="${2:-}"
      shift 2
      ;;
    --version)
      VERSION="${2:-}"
      shift 2
      ;;
    --skip-build)
      SKIP_BUILD=1
      shift
      ;;
    --skip-installer)
      SKIP_INSTALLER=1
      shift
      ;;
    --keep-staging)
      KEEP_STAGING=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "[ERROR] Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$ARCH" ]]; then
  echo "[ERROR] --arch is required" >&2
  usage
  exit 1
fi

if [[ "$ARCH" != "amd64" && "$ARCH" != "arm64" ]]; then
  echo "[ERROR] unsupported arch: $ARCH (expected amd64 or arm64)" >&2
  exit 1
fi

case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*|Windows_NT) ;;
  *)
    echo "[ERROR] this script must run on a Windows host (Git Bash / MSYS)" >&2
    exit 1
    ;;
esac

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[ERROR] required command not found: $1" >&2
    exit 1
  fi
}

require_cmd wails
require_cmd powershell

# Convert a Unix-style path to a native Windows path for tools like makensis.
to_win_path() {
  if command -v cygpath >/dev/null 2>&1; then
    cygpath -w "$1"
  else
    # Fallback: /d/foo/bar -> D:\foo\bar
    printf '%s\n' "$1" | sed -E 's#^/([a-zA-Z])/#\1:/#; s#/#\\#g'
  fi
}

if [[ -z "$VERSION" ]]; then
  VERSION="$(grep -o '"productVersion"[[:space:]]*:[[:space:]]*"[^"]*"' "$ROOT_DIR/wails.json" | head -n 1 | sed -E 's/.*"productVersion"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')"
fi
if [[ -z "$VERSION" ]]; then
  echo "[ERROR] failed to resolve version; pass --version explicitly" >&2
  exit 1
fi

TARGET="windows-$ARCH"
APP_ICON_SRC="$ROOT_DIR/build/appicon.png"
APP_BIN="$ROOT_DIR/build/bin/ant-chrome.exe"
WAILS_CONFIG="$ROOT_DIR/wails.json"
WINDOWS_ICON="$ROOT_DIR/build/windows/icon.ico"
INSTALLER_NSI="$ROOT_DIR/publish/installer.nsi"
CONFIG_INIT_SRC="$ROOT_DIR/publish/config.init.yaml"

echo "========================================"
echo "  Ant Browser Windows Publish"
echo "========================================"
echo "Target : $TARGET"
echo "Version: $VERSION"
echo "Root   : $ROOT_DIR"
echo

if [[ ! -f "$WAILS_CONFIG" ]]; then
  echo "[ERROR] wails.json missing: $WAILS_CONFIG" >&2
  echo "        This development branch must keep a complete Wails source tree." >&2
  exit 1
fi

if [[ ! -f "$CONFIG_INIT_SRC" ]]; then
  echo "[ERROR] windows config template missing: $CONFIG_INIT_SRC" >&2
  exit 1
fi

if [[ "$SKIP_BUILD" -ne 1 ]]; then
  echo "[1/4] Installing frontend dependencies..."
  (cd "$ROOT_DIR/frontend" && BROWSERSLIST_IGNORE_OLD_DATA=1 npm ci --prefer-offline --no-audit --no-fund)

  echo "[2/4] Building frontend assets..."
  (cd "$ROOT_DIR/frontend" && BROWSERSLIST_IGNORE_OLD_DATA=1 npm run build:clean)

  echo "[3/4] Building Windows binary with Wails..."
  rm -f "$APP_BIN"
  (
    cd "$ROOT_DIR"
    wails build -s -platform "windows/$ARCH" -o ant-chrome.exe
  )
else
  echo "[WARN] skipping build step"
fi

if [[ ! -f "$APP_BIN" ]]; then
  echo "[ERROR] app binary not found: $APP_BIN" >&2
  exit 1
fi

echo "[4/4] Assembling staging files..."
APP_STAGE="$STAGING_ROOT/$TARGET/app"
rm -rf "$APP_STAGE"
mkdir -p "$APP_STAGE" "$OUTPUT_DIR"

cp "$APP_BIN" "$APP_STAGE/ant-chrome.exe"
cp "$CONFIG_INIT_SRC" "$APP_STAGE/config.yaml"

ZIP_NAME="AntBrowser-${VERSION}-windows-${ARCH}.zip"
ZIP_PATH="$OUTPUT_DIR/$ZIP_NAME"
rm -f "$ZIP_PATH"
ZIP_PATH_WIN="$(to_win_path "$ZIP_PATH")"
(
  cd "$APP_STAGE"
  powershell -NoProfile -ExecutionPolicy Bypass -Command \
    "Compress-Archive -Path * -DestinationPath '$ZIP_PATH_WIN' -Force"
)
echo "  portable zip : $ZIP_PATH"

if [[ "$SKIP_INSTALLER" -ne 1 ]]; then
  if ! command -v makensis >/dev/null 2>&1; then
    echo "[WARN] makensis not found in PATH; skipping installer (install NSIS or pass --skip-installer)."
  elif [[ ! -f "$WINDOWS_ICON" ]]; then
    echo "[WARN] windows icon missing: $WINDOWS_ICON; skipping installer." >&2
    echo "       Run a full 'wails build' so it can generate build/windows/icon.ico from build/appicon.png." >&2
  else
    INSTALLER_NAME="AntBrowser-Setup-${VERSION}-${ARCH}.exe"
    INSTALLER_PATH="$OUTPUT_DIR/$INSTALLER_NAME"
    rm -f "$INSTALLER_PATH"
    STAGING_WIN="$(to_win_path "$APP_STAGE")"
    INSTALLER_WIN="$(to_win_path "$INSTALLER_PATH")"
    (
      cd "$ROOT_DIR/publish"
      # Git Bash converts /D... into a filesystem path; use // so makensis receives /D...
      makensis "//DVERSION=$VERSION" "//DSTAGINGDIR=$STAGING_WIN" "//DOUTFILE=$INSTALLER_WIN" installer.nsi
    )
    echo "  installer    : $INSTALLER_PATH"
  fi
else
  echo "[WARN] skipping installer step"
fi

if [[ "$KEEP_STAGING" -ne 1 ]]; then
  rm -rf "$APP_STAGE"
fi

if [[ -n "$APP_ICON_SRC" && ! -f "$APP_ICON_SRC" ]]; then
  echo "[WARN] app icon missing: $APP_ICON_SRC (build may still succeed via wails defaults)" >&2
fi

echo "Done."
