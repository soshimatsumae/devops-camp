# 0015: HTTPサーバーにタイムアウトを設定する

## Status

Accepted

## Context

`main.go`は`http.ListenAndServe(":"+cfg.Port, mux)`を直接呼んでおり、リクエストヘッダー・ボディの読み取りやレスポンス書き込みに時間制限がなかった。`ReadHeaderTimeout`が未設定だと、リクエストヘッダーを意図的にゆっくり送り続けることでサーバーの接続枠を占有し続けるSlowloris型の攻撃を受けやすい(静的解析ツール`gosec`では`G114`として指摘される項目)。Step3提出後のレビューで指摘を受けた。

## Decision

`http.ListenAndServe`の直接呼び出しをやめ、`http.Server`構造体を明示的に組み立てて以下を設定する。

- `ReadHeaderTimeout`: 5秒
- `ReadTimeout`: 10秒
- `WriteTimeout`: 10秒
- `IdleTimeout`: 60秒

## Consequences

- Slowlorisのようにヘッダー送信を引き延ばす攻撃に対して、一定時間で接続が切られるようになる
- 通常のリクエスト(このAPIはボディが小さいJSON中心)がこれらの時間内に収まらないケースは想定していないが、将来大きなペイロードを扱うエンドポイントを追加する場合はこの値を見直す必要がある
- `http.Server`を明示的に持つことで、[ADR-0016](./0016-graceful-shutdown.md)のGraceful shutdownの実装土台にもなっている
