# 0011: 本番環境ではJWT_SECRETのデフォルト値を拒否する

## Status

Accepted

## Context

`JWT_SECRET`のデフォルト値が`"dev-secret-change-me"`という、誰でも推測できる固定文字列になっている([ADR-0004](./0004-jwt-auth.md))。ローカル開発では便利だが、`JWT_SECRET`を設定し忘れたまま本番相当の環境で起動できてしまう構造でもあり、誰でも推測できる秘密鍵でJWTを署名してしまうリスクがある。Step3提出後のレビューで指摘を受けた。

## Decision

`internal/config.Load()`が環境変数`APP_ENV`を確認し、`APP_ENV=production`かつ`JWT_SECRET`がデフォルト値のままの場合はエラーを返す。`Load()`の戻り値を`(Config, error)`に変更し、`main.go`側で`db.Open`と同様に`log.Fatalf`でプロセスを終了させる(既存のエラーハンドリングパターンとの一貫性を優先し、`config`パッケージ自身がプロセスを終了させる案は採らなかった)。

## Consequences

- ローカル開発(`APP_ENV`未設定)では従来通り`JWT_SECRET`を省略できる
- `APP_ENV=production`で起動する際、`JWT_SECRET`を明示的に設定していないとサーバーが起動しない。Step7-8でのデプロイ設定時に、この環境変数を必ず設定する必要がある
- `config.Load()`の呼び出し側(`main.go`)は戻り値が1つ増えたエラーハンドリングに対応する必要がある
