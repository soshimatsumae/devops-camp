# 0017: 500エラーの詳細はサーバー側ログにのみ残し、クライアントには汎用メッセージを返す

## Status

Accepted

## Context

各ハンドラの`500`エラーは`httpx.WriteError(w, 500, "INTERNAL_ERROR", "failed to create user")`のように、処理内容に応じた個別メッセージをクライアントへ返す一方、サーバー側には元の`error`の中身を一切ログに残していなかった。そのため、動作確認や本番運用で500が発生しても、原因を追跡する手がかりがない構造だった。Step3提出後のレビューで指摘を受けた。

## Decision

`internal/httpx`に`WriteInternalError(w http.ResponseWriter, err error)`を追加する。

- `slog.Error("internal server error", "error", err)`でサーバー側のログにのみ元のエラーを出力する
- クライアントには常に`{"error":{"code":"INTERNAL_ERROR","message":"internal server error"}}`という汎用レスポンスを返す

`internal/handlers`配下の全20箇所の`500`エラーをこのヘルパー経由に置き換え、個別だったメッセージ文言(`"failed to create user"`等)は廃止して汎用メッセージに統一した。

## Consequences

- クライアントは`error.code`(`INTERNAL_ERROR`)だけを見て分岐する設計([ADR-0009](./0009-unified-error-response-format.md))なので、メッセージ文言の統一による互換性への影響はない
- 実装の詳細(どのSQL文が失敗したか等)をクライアントに漏らさなくなり、セキュリティ的にも改善する
- 500発生時の原因調査は、サーバー側のログ(`slog`の出力)を見ることが前提になる。ログの永続化・集約先(将来的なStep7-8のインフラ)は別途検討が必要
