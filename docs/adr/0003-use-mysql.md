# 0003: データベースにMySQLを採用する

## Status

Accepted

## Context

Step3(バックエンド実装)ではローカル環境で動作確認できるDBが必要だった。DB自体の選定はキャンプ側から指定されておらず、個人の選択に委ねられていた。

## Decision

データベースにMySQLを採用する。ローカル開発は`root@tcp(127.0.0.1:3306)`への接続を前提とする(`internal/config`のデフォルト値)。

## Consequences

- ローカルではHomebrew経由のMySQLサーバーを都度起動しておく必要がある
- rootアカウントの認証情報(パスワード有無)がローカル環境ごとに異なりうるため、`MYSQL_DSN`/`TEST_MYSQL_DSN`は環境変数で都度指定する運用にしている(README.md参照)
- 将来Step7-8でクラウドインフラへ展開する際は、マネージドMySQL(RDS等)への差し替えを想定する
