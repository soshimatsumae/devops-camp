#!/bin/bash
# Next.js用EC2のユーザーデータ。起動テンプレート(web-template)の「ユーザーデータ」に全文を貼り付ける。
# EC2の初回起動時にrootで1回だけ実行され、出力は /var/log/cloud-init-output.log に残る。
# 手順: スワップ作成 → Node.jsを入れる → GitHubからコードを取得 → 本番ビルド → systemdで起動
set -euo pipefail

REPO_URL=https://github.com/soshimatsumae/devops-camp.git
APP_DIR=/opt/devops-camp
NODE_VERSION=22.12.0

# ユーザーデータの実行時はHOMEが未設定で、npmのキャッシュの置き場所が決まらないため指定する
export HOME=/root

echo "== create swap"
# t3.microはメモリが1GBで、Next.jsのビルド中に足りなくなることがあるため、ディスクに2GBのスワップを作る
fallocate -l 2G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile

echo "== install packages"
dnf install -y git

echo "== install Node.js"
curl -fsSL "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz" \
  | tar -C /usr/local --strip-components=1 -xz

echo "== clone repository"
git clone --depth 1 "$REPO_URL" "$APP_DIR"
cd "$APP_DIR/frontend"

echo "== build"
npm ci
# APIの接続先を空にすると、lib/api.ts は /tasks のような相対パスでリクエストする。
# ブラウザは今開いているALBのアドレスへ送り、ALBがパスでGo APIへ振り分ける
NEXT_PUBLIC_API_BASE_URL= NEXT_TELEMETRY_DISABLED=1 npm run build
chown -R ec2-user:ec2-user "$APP_DIR"

echo "== start service"
# systemdに登録し、EC2の再起動時やアプリが落ちたときにも自動で起動し直すようにする
cat > /etc/systemd/system/devops-camp-web.service <<EOF
[Unit]
Description=devops-camp Next.js frontend
After=network-online.target
Wants=network-online.target

[Service]
WorkingDirectory=$APP_DIR/frontend
Environment=NODE_ENV=production PORT=3000 NEXT_TELEMETRY_DISABLED=1
ExecStart=/usr/local/bin/npm start
User=ec2-user
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable --now devops-camp-web

echo "== done"
