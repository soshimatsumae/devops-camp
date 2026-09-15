# internal/handlers/calendar_range_test.go

`tasks.go`内の非公開関数`parseCalendarRange`だけを対象にしたユニットテスト。`package handlers`(内部テスト)として書かれているため、DB不要かつ非公開関数に直接アクセスできる。

## テストケース

| テスト名 | 検証内容 |
|---|---|
| `TestParseCalendarRange_Default` | `from`/`to`とも未指定なら、`to`が当日・`from`が当日から365日前になる |
| `TestParseCalendarRange_ExplicitRange` | `from`/`to`を明示指定した場合、それぞれの値がそのままパースされる |
| `TestParseCalendarRange_InvalidFormat` | `"2026/01/01"`のように`YYYY-MM-DD`でない形式はエラーになる |
| `TestParseCalendarRange_FromAfterTo` | `from`が`to`より後の日付だとエラーになる |

## 読むときのポイント

- このファイルだけ`package handlers`(`_test`サフィックスなし)であり、`tasks_test.go`等の`package handlers_test`とは別物。パッケージ外部に公開する必要のないロジック(日付範囲の解決)を、外部APIとは切り離して単体テストするための使い分け
- DB結合テスト(`tasks_test.go`)とは異なり`TEST_MYSQL_DSN`に依存しないため、`go test ./...`だけで常に実行される(CIやローカルの`go build`確認時に最初に落ちるのはこちら)
