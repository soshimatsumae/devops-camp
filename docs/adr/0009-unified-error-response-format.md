# 0009: エラーレスポンス形式を統一する

## Status

Accepted

## Context

各エンドポイントがバラバラなエラー形式を返すと、フロントエンド側でエラーハンドリングの分岐実装が煩雑になる。

## Decision

すべてのエラーレスポンスを`{"error":{"code":"...","message":"..."}}`の形式に統一する。実装は`internal/httpx.WriteError`に集約し、各ハンドラはこの関数経由でのみエラーを返す。`code`はエラー種別ごとに一意な文字列(例: `VALIDATION_ERROR`, `TASK_NOT_FOUND`, `EMAIL_ALREADY_REGISTERED`)を割り当てる。

## Consequences

- フロントエンド側は`error.code`だけを見て分岐でき、`message`の文言変更がクライアントロジックに影響しない
- 新しいエラーケースを追加するたびに一意な`code`を決める規律が必要(現状は各ハンドラ内で個別に文字列リテラルとして定義しており、一覧化・重複チェックの仕組みはない)
- 認証必須の全エンドポイントは、認証エラー時に共通で`401 UNAUTHORIZED`を返しうる(`internal/auth.RequireAuth`, `internal/handlers/tasks.go`の`currentUserID`)
