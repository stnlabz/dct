#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

SKIP_DEPS=0

usage() {
  echo "Usage: $0 [--skip-deps]"
  echo "  --skip-deps    Do not apt install dependencies; only report missing."
  exit 1
}

# Parse flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-deps) SKIP_DEPS=1; shift ;;
    -h|--help) usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

# If we plan to install deps, require apt-get.
if (( SKIP_DEPS == 0 )); then
  if ! command -v apt-get >/dev/null 2>&1; then
    echo "ERR: apt-get not found. Use --skip-deps on non-Debian/Ubuntu systems." >&2
    exit 1
  fi
fi

# sudo if needed
SUDO=""
if [ "$EUID" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  elif (( SKIP_DEPS == 0 )); then
    echo "ERR: need root or sudo to install packages." >&2
    exit 1
  fi
fi

# ---------- dependency check ----------
need_cmds=( tar zip bzip2 md5sum sha1sum sha256sum b3sum go )

pkg_for_cmd() {
  case "$1" in
    tar)       echo "tar" ;;
    zip)       echo "zip" ;;
    bzip2)     echo "bzip2" ;;
    md5sum)    echo "coreutils" ;;
    sha1sum)   echo "coreutils" ;;
    sha256sum) echo "coreutils" ;;
    go)        echo "golang-go" ;;
    *)         echo "" ;;
  esac
}

missing_cmds=()
missing_pkgs=()

for c in "${need_cmds[@]}"; do
  if ! command -v "$c" >/dev/null 2>&1; then
    missing_cmds+=("$c")
    p="$(pkg_for_cmd "$c")"
    if [[ -n "$p" && " ${missing_pkgs[*]-} " != *" $p "* ]]; then
      missing_pkgs+=("$p")
    fi
  fi
done

if (( SKIP_DEPS == 0 )); then
  if ((${#missing_pkgs[@]} > 0)); then
    echo "Installing missing dependencies via apt: ${missing_pkgs[*]}"
    $SUDO apt-get update -y
    # Try to install; continue even if some (like blake3) have no candidate.
    DEBIAN_FRONTEND=noninteractive $SUDO apt-get install -y "${missing_pkgs[@]}" || true
  else
    echo "All apt-managed dependencies present."
  fi
else
  if ((${#missing_cmds[@]} > 0)); then
    echo "NOTE: --skip-deps set. Missing commands (not installed): ${missing_cmds[*]}"
  else
    echo "All required commands present (deps skipped)."
  fi
fi


# ---------- build & install dct ----------
echo "Building dct..."
cd "$ROOT_DIR"
mkdir -p bin
go build -o bin/dct .

echo "Installing to /usr/local/bin..."
${SUDO:+$SUDO }install -m 0755 bin/dct /usr/local/bin/dct

# runtime dirs
mkdir -p projects publish

echo
echo "OK."
echo "Binary: /usr/local/bin/dct"
echo "Usage : dct -src ./your_project"

