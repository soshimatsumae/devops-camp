# internal/config/config.go

環境変数から設定値を読み込む唯一の場所。

## やっていること

- `Config`構造体: `Port`・`MySQLDSN`・`JWTSecret`・`TokenTTL`の4フィールド
- `Load()`: 各値を`getEnv(key, fallback)`経由で環境変数から取得し、未設定ならデフォルト値を使う
- `getEnv`: `os.Getenv`が空文字列を返したらフォールバック値を使う小さなヘルパー

## デフォルト値

| 変数名 | デフォルト値 |
|---|---|
| `PORT` | `8080` |
| `MYSQL_DSN` | `root@tcp(127.0.0.1:3306)/devops_camp?parseTime=true&charset=utf8mb4` |
| `JWT_SECRET` | `dev-secret-change-me` |

`TokenTTL`は環境変数化されておらず、コード上で`time.Hour`固定。

## 読むときのポイント

- `getEnv`は「空文字列 = 未設定」とみなす簡易実装。環境変数に意図的に空文字列を設定するケースは区別できないが、このアプリの設定項目ではその必要がないため割り切っている
- `MYSQL_DSN`のデフォルト値はパスワードなしのroot接続を前提にしている。ローカルのMySQLにパスワードが設定されている場合は、この環境変数を明示的に上書きする必要がある(README.mdのセットアップ手順参照)
