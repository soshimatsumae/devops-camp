#!/bin/bash
# Go API用EC2のユーザーデータ。起動テンプレート(api-template)の「ユーザーデータ」に全文を貼り付ける。
# EC2の初回起動時にrootで1回だけ実行され、出力は /var/log/cloud-init-output.log に残る。
# 手順: GitHubからコードを取得 → Goでビルド → Parameter Storeから設定を読む → テーブル作成 → systemdで起動
#
# 前提: EC2にIAMロール(devops-camp-api-role)が付いていて、Parameter Storeの /devops-camp/* を読めること。
# 秘密情報がログに残らないよう、set -x(実行したコマンドの表示)は使わない。
set -euo pipefail

REGION=ap-northeast-1
REPO_URL=https://github.com/soshimatsumae/devops-camp.git
APP_DIR=/opt/devops-camp
PARAM_PREFIX=/devops-camp
DB_NAME=devops_camp

# ユーザーデータの実行時はHOMEが未設定で、Goのビルドキャッシュの置き場所が決まらないため指定する
export HOME=/root

echo "== install packages"
# mariadb105 は mysql コマンド(テーブル作成に使う)を含む
dnf install -y git mariadb105

echo "== clone repository"
git clone --depth 1 "$REPO_URL" "$APP_DIR"
cd "$APP_DIR"

echo "== install Go"
# dnfで入るGoはgo.modの指定より古いことがあるため、go.modに書かれたバージョンを公式から入れる
GO_VERSION=$(awk '/^go /{print $2}' go.mod)
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /usr/local -xz
export PATH=/usr/local/go/bin:$PATH

echo "== build"
go build -o /usr/local/bin/devops-camp-api .

echo "== read parameters"
get_param() {
  aws ssm get-parameter --region "$REGION" --name "$PARAM_PREFIX/$1" \
    --with-decryption --query Parameter.Value --output text
}
DB_HOST=$(get_param db-host)
DB_USER=$(get_param db-user)
DB_PASSWORD=$(get_param db-password)
JWT_SECRET=$(get_param jwt-secret)

echo "== create tables"
# schema.sql は CREATE TABLE IF NOT EXISTS なので、2台目以降のEC2が実行しても既存のデータは消えない
export MYSQL_PWD="$DB_PASSWORD"
mysql -h "$DB_HOST" -u "$DB_USER" -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4"
mysql -h "$DB_HOST" -u "$DB_USER" "$DB_NAME" < schema.sql
unset MYSQL_PWD

echo "== write environment file"
# アプリはこのファイルの値を環境変数として受け取る(internal/config/config.go)。rootだけが読める権限にする
install -d -m 700 /etc/devops-camp
cat > /etc/devops-camp/api.env <<EOF
PORT=8080
APP_ENV=production
JWT_SECRET=${JWT_SECRET}
MYSQL_DSN=${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:3306)/${DB_NAME}?parseTime=true&charset=utf8mb4
EOF
chmod 600 /etc/devops-camp/api.env

echo "== start service"
# systemdに登録し、EC2の再起動時やアプリが落ちたときにも自動で起動し直すようにする
cat > /etc/systemd/system/devops-camp-api.service <<'EOF'
[Unit]
Description=devops-camp Go API
After=network-online.target
Wants=network-online.target

[Service]
EnvironmentFile=/etc/devops-camp/api.env
ExecStart=/usr/local/bin/devops-camp-api
User=ec2-user
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable --now devops-camp-api

echo "== done"
