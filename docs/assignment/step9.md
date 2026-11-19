# Step9 課題提出: コンテナ1

Step9の課題ページ([final-challenge/app/step9](https://devopscamp.reheartcloud.com/final-challenge/app/step9))に沿って、Step3のバックエンド(Go)とStep6のフロントエンド(Next.js)をコンテナ化し、Docker ComposeでNext.js・Go API・MySQLの3層をローカルでまとめて起動する。

## 1. 概要

<!-- Dockerfileとcompose構成の特徴・工夫点を簡潔に書く -->

### 構成の特徴

- Go APIはマルチステージビルドでイメージを作る。1段目(`golang:1.26.5-alpine`)でビルドし、2段目(`alpine:3.22`)にはできあがったバイナリだけをコピーする。Goのコンパイラやソースコードは最終的なイメージに残らない
- Next.jsも同じくマルチステージビルドでイメージを作る。`next.config.ts`の`output: "standalone"`で、ビルド結果のうち動かすのに必要なファイルだけを集め、実行用の段(`node:22.12.0-alpine`)にはそれだけをコピーする
- Docker Composeで、Next.js(`web`)・Go API(`api`)・MySQL 8.4(`db`)の3つのコンテナを起動する。APIはサービス名の`db`を接続先にして、Composeが自動で作るネットワークの中でDBにつなぐ
- 画面のJavaScriptはブラウザで動くため、APIへのリクエストは`web`コンテナを経由せず、ブラウザから直接`http://localhost:8080`(`api`コンテナ)へ送る。ブラウザで開くのは`http://localhost:3001`(手元の`npm run dev`が使う3000番とぶつからないようにした)
- テーブルは、MySQLの公式イメージの仕組み(初回起動時に`/docker-entrypoint-initdb.d/`のSQLを実行する)を使い、`schema.sql`で作る。Step8のように、手で`mysql`コマンドを実行する必要はない
- DBのデータは名前付きボリューム(`db-data`)に保存し、コンテナを作り直しても消えないようにした
- DBのパスワードと`JWT_SECRET`は`compose.yaml`に書かず、Gitの管理から外した`.env`に置く。Composeが`.env`を読み、それぞれのコンテナに必要な値を渡す
- 課題ではフロントのコンテナ化は任意のため、まず必須のAPIとDBを動かして確認し、そのあとNext.jsを追加して3層にした

### 工夫した点

| 工夫 | 内容と理由 |
|---|---|
| DBの準備ができてからAPIを起動する | Go APIは起動時にDBへ接続できないと終了する。`depends_on`だけでは「コンテナの起動順」しか制御できないため、DBのヘルスチェック(`mysqladmin ping`)と`condition: service_healthy`を組み合わせ、DBが応答するようになってからAPIを起動する |
| ヘルスチェックの宛先を`127.0.0.1`にした | MySQLの公式イメージは、初回起動時にテーブルを作るための一時サーバーを立ち上げる。一時サーバーはコンテナ内のソケット(`localhost`で接続したときの経路)でしか応答しないため、`localhost`宛てだとテーブル作成中にもヘルスチェックが通ってしまう。`127.0.0.1`(ネットワーク経由)にして、本番のサーバーが起動してから通るようにした |
| 依存ライブラリのダウンロードを先に行う | `go.mod`・`go.sum`だけを先にコピーして`go mod download`を実行する。Dockerは手順の1行ごとに結果をキャッシュするため、コードだけを直したときの再ビルドでは、ダウンロードがキャッシュで済む |
| イメージのバージョンを固定する | `latest`はビルドする時期によって中身が変わるため、`golang:1.26.5-alpine`(`go.mod`のバージョン)・`alpine:3.22`・`mysql:8.4`(Step8のRDSと同じ)のように固定した |
| root以外のユーザーで動かす | 実行用の段で専用ユーザー`app`を作り、`USER app`で切り替える。万一アプリが乗っ取られても、コンテナ内でできることを減らすため |
| ビルドに要らないファイルを送らない | `.dockerignore`で`.git`・`.env`・`docs`・`frontend`などを除き、ビルドを速くし、秘密情報がイメージに入らないようにした。Next.js用の`frontend/.dockerignore`では、手元の`node_modules`・`.next`を除き、コンテナの中で作り直す |
| Next.jsのAPIの接続先はビルド時に渡す | `NEXT_PUBLIC_`で始まる環境変数は、実行時ではなくビルド時に画面のJavaScriptへ埋め込まれる。そのため`compose.yaml`の`environment`では渡さず、Dockerfileの`ARG`(ビルド時の値、初期値は`http://localhost:8080`)で渡すようにした。ALBの後ろに置くときは、Step8と同じく空にしてビルドすればよい |
| APIにもヘルスチェックを付けた | `/health`を`wget`で確認し、`docker compose ps`でAPIの状態も分かるようにした。実行用のベースを`alpine`にしたため、`wget`がそのまま使える |
| パスワードを`.env`に分けた | 講座の「実践的な設定」にならい、DBのパスワードと`JWT_SECRET`を`compose.yaml`に書かず、`.env`に置いた。`.env`は`.gitignore`と`.dockerignore`で除外し、GitHubにもイメージにも入らない。必要な変数の一覧は、値を空にした`.env.example`で共有する |
| APIには必要な値だけを渡す | DBには`env_file`で`.env`の変数をまとめて渡す。APIにも`env_file`を使うと、APIが使わないrootのパスワードまで渡ってしまうため、APIは`${MYSQL_PASSWORD}`のように`compose.yaml`の中で値を差し込み、接続文字列(DSN)と`JWT_SECRET`だけを渡した |

`restart`(再起動ポリシー)と`docker compose watch`(ファイル監視)は、講座で「開発環境では不要」「必要になってから検討する」とされているため、ローカルの動作確認用の今回は使っていない。AWS上では、Step8と同じくParameter Storeの値を使う。

## 2. 実装コード

- GitHubリポジトリ: <https://github.com/soshimatsumae/devops-camp>
- [Dockerfile](../../Dockerfile)(Go API。マルチステージビルド)
- [frontend/Dockerfile](../../frontend/Dockerfile)(Next.js。マルチステージビルド)
- [compose.yaml](../../compose.yaml)(Next.js・Go API・MySQL)
- [.dockerignore](../../.dockerignore)・[frontend/.dockerignore](../../frontend/.dockerignore)
- [frontend/next.config.ts](../../frontend/next.config.ts)(`output: "standalone"`を追加)
- [.env.example](../../.env.example)(`.env`の見本。値は空にしてある)

起動する前に、`.env.example`をコピーして`.env`を作り、値を入れる。パスワードと`JWT_SECRET`は、次のようにランダムな値で作れる。

```bash
cat > .env <<EOF
MYSQL_ROOT_PASSWORD=$(openssl rand -hex 16)
MYSQL_DATABASE=devops_camp
MYSQL_USER=app
MYSQL_PASSWORD=$(openssl rand -hex 16)
JWT_SECRET=$(openssl rand -hex 32)
EOF
docker compose up -d --build
```

ファイル名は、講座で推奨されている`compose.yaml`にした(課題ページの`docker-compose.yml`と同じ役割)。Goのコード(`go.mod`)がリポジトリ直下にあるため、Go APIのDockerfileとComposeファイルはリポジトリ直下に、Next.jsのDockerfileは`frontend/`に置いた。

## 3. 動作確認

<!-- docker compose upの実行結果。バックエンドとDBの連携動作がわかるもの -->

### 3-1. 起動(`docker compose up -d --build`)

DBのヘルスチェックが通って(`Healthy`)から、APIが起動している(`Started`)。

```text
[+] up 18/18
 ✔ Image mysql:8.4             Pulled                                                                       9.9s
 ✔ Image devops-camp-api       Built                                                                       10.6s
 ✔ Volume devops-camp_db-data  Created                                                                      0.0s
 ✔ Network devops-camp_default Created                                                                      0.0s
 ✔ Container devops-camp-db-1  Healthy                                                                      5.7s
 ✔ Container devops-camp-api-1 Started                                                                      5.7s
```

```text
$ docker compose ps
NAME                IMAGE             COMMAND                  SERVICE   CREATED         STATUS                    PORTS
devops-camp-api-1   devops-camp-api   "api"                    api       6 seconds ago   Up Less than a second (health: starting)   0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp
devops-camp-db-1    mysql:8.4         "docker-entrypoint.s…"   db        6 seconds ago   Up 5 seconds (healthy)                    3306/tcp, 33060/tcp
```

### 3-2. ヘルスチェック

```text
$ curl -i http://localhost:8080/health
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

### 3-3. POSTでDBに書き込む

ユーザー登録とタスク作成がどちらも`201 Created`になり、DBに保存された。

```text
$ curl -i -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"テスト"}'
HTTP/1.1 201 Created

{"id":1,"email":"test@example.com","name":"テスト","created_at":"2026-11-18T14:47:55Z"}
```

```text
$ curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
{"access_token":"(省略)","token_type":"Bearer","expires_in":604800}
```

```text
$ curl -i -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Docker動作確認","estimated_hours":1.5}'
HTTP/1.1 201 Created

{"id":1,"parent_id":null,"list_id":null,"title":"Docker動作確認","description":null,"status":"todo","estimated_hours":1.5,"actual_hours":null,"due_date":null,"completed_at":null,"created_at":"2026-11-18T14:48:57Z","updated_at":"2026-11-18T14:48:57Z"}
```

### 3-4. GETでDBから読み取る

書き込んだタスクを、APIからもDBからも取得できた。

```text
$ curl -i http://localhost:8080/tasks -H "Authorization: Bearer $TOKEN"
HTTP/1.1 200 OK

{"data":[{"id":1,"parent_id":null,"list_id":null,"title":"Docker動作確認","description":null,"status":"todo","estimated_hours":1.5,"actual_hours":null,"due_date":null,"completed_at":null,"created_at":"2026-11-18T14:48:57Z","updated_at":"2026-11-18T14:48:57Z"}],"meta":{"page":1,"per_page":20,"total_count":1,"total_pages":1}}
```

DBのコンテナに入り、テーブルの中身を直接確認した。

```text
$ docker compose exec db mysql -u app -plocal-app-password --default-character-set=utf8mb4 devops_camp \
  -e "SELECT id, email, name FROM users; SELECT id, title, estimated_hours FROM tasks;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+------------------+-----------+
| id | email            | name      |
+----+------------------+-----------+
|  1 | test@example.com | テスト    |
+----+------------------+-----------+
+----+--------------------+-----------------+
| id | title              | estimated_hours |
+----+--------------------+-----------------+
|  1 | Docker動作確認     |             1.5 |
+----+--------------------+-----------------+
```

### 3-5. `.env`に切り替えたあとの確認

3-1〜3-4は、パスワードを`compose.yaml`に直接書いていたときの結果。`.env`に切り替えたあと、`docker compose down -v`でDBのデータを消してから起動し直し、同じ確認を行った(MySQLがパスワードを設定するのは、データが空の状態で起動したときだけのため)。

切り替える前に受け取ったトークンでタスクを作ろうとすると、`401 Unauthorized`になった。トークンの署名に使う`JWT_SECRET`が`.env`の値に変わったためで、APIが`.env`の値を使っていることが確認できた。

```text
$ curl -i -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Docker動作確認","estimated_hours":1.5}'
HTTP/1.1 401 Unauthorized

{"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}
```

ユーザー登録とログインをやり直したあとは、書き込みと読み取りができた。

```text
$ curl -i -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Docker動作確認","estimated_hours":1.5}'
HTTP/1.1 201 Created

{"id":1,"parent_id":null,"list_id":null,"title":"Docker動作確認","description":null,"status":"todo","estimated_hours":1.5,"actual_hours":null,"due_date":null,"completed_at":null,"created_at":"2026-11-18T16:58:16Z","updated_at":"2026-11-18T16:58:16Z"}
```

```text
$ curl -i http://localhost:8080/tasks -H "Authorization: Bearer $TOKEN"
HTTP/1.1 200 OK

{"data":[{"id":1,"parent_id":null,"list_id":null,"title":"Docker動作確認","description":null,"status":"todo","estimated_hours":1.5,"actual_hours":null,"due_date":null,"completed_at":null,"created_at":"2026-11-18T16:58:16Z","updated_at":"2026-11-18T16:58:16Z"}],"meta":{"page":1,"per_page":20,"total_count":1,"total_pages":1}}
```

DBの中身は、コンテナに渡した環境変数(`.env`の値)を使って確認した。パスワードをコマンドに直接書かずに済む。

```text
$ docker compose exec db sh -c 'mysql -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" --default-character-set=utf8mb4 "$MYSQL_DATABASE" \
  -e "SELECT id, email, name FROM users; SELECT id, title, estimated_hours FROM tasks;"'
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+------------------+-----------+
| id | email            | name      |
+----+------------------+-----------+
|  1 | test@example.com | テスト    |
+----+------------------+-----------+
+----+--------------------+-----------------+
| id | title              | estimated_hours |
+----+--------------------+-----------------+
|  1 | Docker動作確認     |             1.5 |
+----+--------------------+-----------------+
```

### 3-6. Next.jsを追加した3層での確認

`web`・`api`・`db`の3つのコンテナが起動した。

```text
$ docker compose ps
NAME                IMAGE             COMMAND                  SERVICE   CREATED              STATUS                        PORTS
devops-camp-api-1   devops-camp-api   "api"                    api       About a minute ago   Up About a minute (healthy)   0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp
devops-camp-db-1    mysql:8.4         "docker-entrypoint.s…"   db        About a minute ago   Up About a minute (healthy)   3306/tcp, 33060/tcp
devops-camp-web-1   devops-camp-web   "docker-entrypoint.s…"   web       About a minute ago   Up 51 seconds (healthy)       0.0.0.0:3001->3000/tcp, [::]:3001->3000/tcp
```

`db`には`PORTS`の転送(`0.0.0.0:…->`)がなく、Macからは直接つながらない。DBにつなげるのは、同じネットワークにいる`api`だけになっている。

ブラウザで`http://localhost:3001`を開き、3-5で登録したユーザーでログインすると、APIで作ったタスク「Docker動作確認」が画面に表示された。

## 4. まとめ

### 苦労した点

- **講座の例との違いを判断すること**：講座のDockerfileは`main.go`だけをコピーしてビルドする簡単な例だったが、このアプリは外部ライブラリ(MySQLドライバなど)を使い、コードも`internal/`に分かれている。講座と同じ書き方でよいところと、このアプリに合わせて変えるべきところ(`go.mod`の先コピー、バージョンの固定など)を1つずつ比べて決める必要があった。ビルドのオプション(`CGO_ENABLED=0`など)は、なくても動くため、講座に合わせてシンプルにした
- **実行用イメージのベースを選ぶこと**:`distroless`・`alpine`・`scratch`のどれにするかで迷った。小さく安全なのは`distroless`だが、シェルが入っていないためコンテナに入って調べられない。初めてのコンテナ化で、うまく動かないときに中を確認できるほうがよいと考え、`alpine`にした

### 学んだこと

- **Step8との比べ方で、コンテナの良さが分かった**：Step8ではEC2の起動のたびに、ユーザーデータでGoのインストール・ビルド・テーブル作成を行っていた。コンテナでは、ビルドした結果をイメージとしてまとめておけるため、起動するときはイメージを動かすだけで済む。Step8のまとめで挙げた「起動時にビルドするため時間がかかる」というリスクを、Step10でのコンテナ化で解消できる見込みが持てた
- **マルチステージビルド**：`FROM`を書くたびに新しい段が始まり、最終的なイメージには最後の段だけが残る。ビルドに必要な道具(Goのコンパイラ)と、動かすのに必要なもの(バイナリ)を分けて考えることで、イメージを小さくできる
- **起動の順番と「準備ができた」ことは別**：`depends_on`はコンテナを起動する順番を決めるだけで、DBが接続を受け付けられる状態になるまでは待たない。ヘルスチェックと`condition: service_healthy`を組み合わせることで、はじめて「準備ができてから起動する」ことができる
- **Composeのネットワークとボリューム**：Composeは自動でネットワークを作り、その中ではサービス名(`db`)を接続先として使える。DBのデータは名前付きボリュームに置くことで、コンテナを消しても残る。`docker compose down -v`のようにボリュームまで消すと、データも消える
- **`.env`の2通りの使い方**：`env_file`は`.env`の変数をすべてコンテナに渡し、`${変数名}`は`compose.yaml`の中の値を`.env`の値に置き換える。コンテナごとに、渡す値を必要なものだけに絞れる。また、MySQLの公式イメージがパスワードを設定するのは初回(データが空のとき)だけで、あとから`.env`を変えても、すでにあるDBのパスワードは変わらないことも分かった
