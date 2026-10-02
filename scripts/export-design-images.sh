#!/usr/bin/env bash
# docs/design/*.html を Chrome のヘッドレスモードで開き、GitHub 表示用の PNG を docs/design/images/ に書き出す。
# リポジトリのルートで実行する: ./scripts/export-design-images.sh
set -euo pipefail

CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DESIGN="$ROOT/docs/design"
OUT="$DESIGN/images"
PROFILE="$(mktemp -d)"
trap 'rm -rf "$PROFILE"' EXIT

mkdir -p "$OUT"

shot() {
  local url="$1" size="$2" file="$3"
  "$CHROME" --headless=new --disable-gpu --hide-scrollbars \
    --user-data-dir="$PROFILE" --virtual-time-budget=5000 \
    --window-size="$size" --screenshot="$OUT/$file" "$url" >/dev/null 2>&1
  echo "wrote docs/design/images/$file"
}

shot "file://$DESIGN/flowmap.html#export" "1860,1350" "flowmap.png"

shot "file://$DESIGN/wireframes.html#export-s01" "940,700"  "wireframe-s01.png"
shot "file://$DESIGN/wireframes.html#export-s02" "940,760"  "wireframe-s02.png"
shot "file://$DESIGN/wireframes.html#export-s03" "940,1100" "wireframe-s03.png"
shot "file://$DESIGN/wireframes.html#export-m01" "940,820"  "wireframe-m01.png"
shot "file://$DESIGN/wireframes.html#export-m02" "940,820"  "wireframe-m02.png"
shot "file://$DESIGN/wireframes.html#export-m03" "940,820"  "wireframe-m03.png"
