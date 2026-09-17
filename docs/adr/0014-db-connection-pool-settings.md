# 0014: DB接続プールの上限・生存時間を明示的に設定する

## Status

Accepted

## Context

`internal/db.Open`は`sql.Open`でハンドルを作り`Ping`するだけで、接続プールの挙動(同時接続数の上限・アイドル接続の保持数・接続の生存時間)を`database/sql`のデフォルトに任せていた。Step3提出後のレビューで、接続数が想定外に増えたり、MySQL側が長時間アイドルな接続を切断した後にGo側が古い接続を掴んでしまう、といったリスクを指摘された。

## Decision

`db.Open`で以下を明示的に設定する。

- `SetMaxOpenConns(25)`
- `SetMaxIdleConns(25)`(`MaxOpenConns`と同値にし、負荷変動のたびに接続を切断・再接続するコストを避ける)
- `SetConnMaxLifetime(5 * time.Minute)`

値は環境ごとの調整を想定せず、まずは固定値としてハードコードする。

## Consequences

- MySQL側の切断によって古い接続を掴み続ける問題を、`ConnMaxLifetime`による定期的な再接続で予防できる
- `MaxIdleConns`を`MaxOpenConns`と揃えたことで、アクセス数が増減しても接続の切断・再確立が起きにくくなる
- 将来、環境(ローカル/本番)ごとに値を変えたくなった場合は、`internal/config`経由で環境変数化することを検討する(現時点では規模的にオーバーエンジニアリングと判断し見送った)
