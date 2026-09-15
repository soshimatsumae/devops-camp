# 0005: パスワードハッシュにbcryptを採用する

## Status

Accepted

## Context

ユーザーのパスワードを平文でDBに保存するわけにはいかない。最低限のセキュリティ対応として、ハッシュ化方式を決める必要があった。

## Decision

`golang.org/x/crypto/bcrypt`を採用し、`bcrypt.DefaultCost`でハッシュ化する(`internal/auth/password.go`)。

## Consequences

- ソルト付与・ストレッチングが標準で組み込まれており、実装コストを抑えつつ最低限の安全性を確保できる
- メール確認・パスワードリセットのフローは今回のスコープでは省略している。本番運用するならこれらの追加実装が前提になる
