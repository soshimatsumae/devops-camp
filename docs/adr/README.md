# Architecture Decision Records

このディレクトリは、devops-campプロジェクトの設計上の意思決定を「なぜそう決めたか」込みで記録するADR(Architecture Decision Record)集。

## フォーマット

各ADRは以下の構成で書く(Nygard形式ベース)。

- **Status**: `Proposed` / `Accepted` / `Deprecated` / `Superseded by ADR-xxxx`
- **Context**: その決定が必要になった背景・制約
- **Decision**: 実際に決めたこと
- **Consequences**: その決定によって生じる影響・トレードオフ

決定を後から覆す場合は既存のADRを書き換えず、新しいADRを追加して`Superseded by`で古い方のStatusを更新する。

## 一覧

| ID | タイトル | Status |
|---|---|---|
| [0001](./0001-use-go.md) | 実装言語にGoを採用する | Accepted |
| [0002](./0002-net-http-only.md) | Webフレームワークを使わず`net/http`のみで実装する | Accepted |
| [0003](./0003-use-mysql.md) | データベースにMySQLを採用する | Accepted |
| [0004](./0004-jwt-auth.md) | 認証にJWTを採用する | Accepted |
| [0005](./0005-bcrypt-password-hashing.md) | パスワードハッシュにbcryptを採用する | Accepted |
| [0006](./0006-no-task-type-master.md) | タスク種別マスタを持たない | Accepted |
| [0007](./0007-self-referencing-task-hierarchy.md) | タスク階層をparent_idの自己参照で表現する | Accepted |
| [0008](./0008-separate-estimated-actual-weight.md) | 想定の大変さと実績の大変さを別カラムに分離する | Accepted |
| [0009](./0009-unified-error-response-format.md) | エラーレスポンス形式を統一する | Accepted |
| [0010](./0010-calendar-date-range-params.md) | カレンダーAPIにfrom/toクエリパラメータを設ける | Accepted |
