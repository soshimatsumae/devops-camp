# 0010: カレンダーAPIにfrom/toクエリパラメータを設ける

## Status

Accepted

## Context

`GET /tasks/calendar`は当初パラメータなしで全期間の集計結果を返す設計だった。Step2のAPI仕様書レビューで「アカウントの利用期間が延びるほどレスポンスサイズが線形に増え、扱いにくい」という指摘を受けた。

## Decision

`GET /tasks/calendar`に`from`/`to`(`YYYY-MM-DD`)のクエリパラメータを追加する。未指定時は「当日から365日前〜当日」をデフォルト範囲とする(`internal/handlers/tasks.go`の`parseCalendarRange`)。`from`が`to`より後の場合は`422 INVALID_DATE_RANGE`、日付形式が不正な場合は`400 INVALID_DATE_FORMAT`を返す。

## Consequences

- デフォルトのレスポンスサイズが1年分に固定され、アカウントの利用期間に比例して肥大化することがなくなる
- フロントエンド側(GitHubのコントリビューションカレンダー風UI)は、表示期間に応じて`from`/`to`を指定する実装が前提になる
- 「全期間の集計を一度に見たい」というユースケースがある場合、クライアント側で`from`/`to`を複数回に分けて呼び出す必要がある(サーバー側で全期間取得のショートカットは提供していない)
