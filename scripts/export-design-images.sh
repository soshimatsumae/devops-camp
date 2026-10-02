#!/usr/bin/env bash
# docs/design/*.html を Chrome のヘッドレスモードで開き、GitHub 表示用の PNG を docs/design/images/ に書き出す。
# リポジトリのルートで実行する: ./scripts/export-design-images.sh
set -euo pipefail

CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DESIGN="$ROOT/docs/design"
OUT="$DESIGN/images"

mkdir -p "$OUT"

shot() {
  local url="$1" size="$2" file="$3"
  local profile
  profile="$(mktemp -d)"

  # --headless=new はスクリーンショット撮影後に自動終了しないことがあるため、
  # 旧来の --headless を使う。さらに、万一ハングしても20秒で強制終了する
  # ウォッチドッグを付けて、スクリプト全体が固まらないようにする。
  "$CHROME" --headless --disable-gpu --hide-scrollbars \
    --user-data-dir="$profile" --virtual-time-budget=5000 \
    --window-size="$size" --screenshot="$OUT/$file" "$url" >/dev/null 2>&1 &
  local pid=$!
  ( sleep 20; kill "$pid" 2>/dev/null || true ) &
  local watchdog=$!

  wait "$pid" 2>/dev/null || true
  kill "$watchdog" 2>/dev/null || true
  wait "$watchdog" 2>/dev/null || true
  rm -rf "$profile"

  if [ -s "$OUT/$file" ]; then
    echo "wrote docs/design/images/$file"
  else
    echo "failed: docs/design/images/$file (タイムアウトまたはエラー)" >&2
  fi
}

shot "file://$DESIGN/flowmap.html#export" "1860,1740" "flowmap.png"

shot "file://$DESIGN/wireframes.html#export-s01" "940,700"  "wireframe-s01.png"
shot "file://$DESIGN/wireframes.html#export-s02" "940,760"  "wireframe-s02.png"
shot "file://$DESIGN/wireframes.html#export-s03" "940,1100" "wireframe-s03.png"
shot "file://$DESIGN/wireframes.html#export-m01" "940,820"  "wireframe-m01.png"
shot "file://$DESIGN/wireframes.html#export-m02" "940,820"  "wireframe-m02.png"
shot "file://$DESIGN/wireframes.html#export-m03" "940,820"  "wireframe-m03.png"
shot "file://$DESIGN/wireframes.html#export-m04" "940,760"  "wireframe-m04.png"
