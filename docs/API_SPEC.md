## API仕様書 {#api仕様書}


### APIの概要 {#apiの概要}

見積もり振り返り型タスク管理ツールのバックエンドAPI。ユーザー認証機能と、タスクのCRUD・親子階層管理・見積もり/実績の工数(時間)の記録・日別集計(コントリビューションカレンダー用)を提供する。対象データはusers(ユーザー)とtasks(タスク、親子関係を持つ)。

### 共通仕様 {#共通仕様}

| 項目 | 内容 |
| :---- | :---- |
| ベースURL | https://api.example.com/v1 |
| データ形式 | リクエスト/レスポンスともにapplication/json |
| 日時フォーマット | ISO 8601(例: 2026-09-09T10:00:00Z、UTC) |
| ページネーション | クエリパラメータpage(デフォルト1)・per\_page(デフォルト20、最大100) レスポンスはdata配列とmeta(total\_count, page, per\_page, total\_pages)を含む |

### 認証方式 {#認証方式}

* 方式: JWT(Bearer トークン)

```
Authorization: Bearer <access_token>
```

* パスワードはbcryptでハッシュ化して保存

### ステータスコード一覧 {#ステータスコード一覧}

| コード | 意味 |
| :---- | :---- |
| 200 OK | 取得・更新成功 |
| 201 Created | 作成成功 |
| 204 No Content | 削除成功 |
| 400 Bad Request | リクエスト形式・パラメータ不正 |
| 401 Unauthorized | 未認証・トークン無効 |
| 403 Forbidden | 他ユーザーのリソースへのアクセス |
| 404 Not Found | リソースが存在しない |
| 422 Unprocessable Entity | バリデーションエラー |
| 500 Internal Server Error | サーバー内部エラー |

認証が必要な全API(TASK-01〜06, LIST-01〜04)は、上記に加えて共通で **401 Unauthorized**(トークン未指定・無効・期限切れ)を返し得る。

### エンドポイント一覧 {#エンドポイント一覧}

| No. | API ID | API名 | メソッド | エンドポイント | 認証 |
| :---- | :---- | :---- | :---- | :---- | :---- |
| 1 | AUTH-01 | ユーザー登録 | POST | /auth/register | 不要 |
| 2 | AUTH-02 | ログイン | POST | /auth/login | 不要 |
| 3 | TASK-01 | タスク一覧取得 | GET | /tasks | 必要 |
| 4 | TASK-02 | タスク作成 | POST | /tasks | 必要 |
| 5 | TASK-03 | タスク詳細取得 | GET | /tasks/{id} | 必要 |
| 6 | TASK-04 | タスク更新 | PATCH | /tasks/{id} | 必要 |
| 7 | TASK-05 | タスク削除 | DELETE | /tasks/{id} | 必要 |
| 8 | TASK-06 | 日毎に完了したタスク合計量取得 | GET | /tasks/calendar | 必要 |
| 9 | LIST-01 | リスト一覧取得 | GET | /lists | 必要 |
| 10 | LIST-02 | リスト作成 | POST | /lists | 必要 |
| 11 | LIST-03 | リスト更新 | PATCH | /lists/{id} | 必要 |
| 12 | LIST-04 | リスト削除 | DELETE | /lists/{id} | 必要 |

### POST /auth/register(ユーザー登録) {#post-/auth/register(ユーザー登録)}

リクエストパラメータ

| 名前 | 型 | 必須 | 説明 |
| :---- | :---- | :---- | :---- |
| email | string | ○ | メールアドレス |
| password | string | ○ | パスワード(8文字以上) |
| name | string | ○ | 表示名 |

リクエスト例

```json
{
  "email": "taro@example.com",
  "password": "password123",
  "name": "太郎"
}
```

レスポンス(201)

| 名前 | 型 | 説明 |
| :---- | :---- | :---- |
| id | integer | ユーザーID |
| email | string | メールアドレス |
| name | string | 表示名 |
| created\_at | string | 作成日時 |

レスポンス例

```json
{
  "id": 1,
  "email": "taro@example.com",
  "name": "太郎",
  "created_at": "2026-09-09T10:00:00Z"
}
```

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 422 | EMAIL\_ALREADY\_REGISTERED | メールアドレスが既に登録されている |
| 422 | VALIDATION\_ERROR | パスワードは8文字以上必要 |

### POST /auth/login(ログイン) {#post-/auth/login(ログイン)}

リクエストパラメータ

| 名前 | 型 | 必須 | 説明 |
| :---- | :---- | :---- | :---- |
| email | string | ○ | メールアドレス |
| password | string | ○ | パスワード |

リクエスト例

```json
{ "email": "taro@example.com", "password": "password123" }
```

レスポンス(200)

| 名前 | 型 | 説明 |
| :---- | :---- | :---- |
| id | integer | ユーザーID |
| access\_token | string | JWTアクセストークン |
| token\_type | string | トークン種別(常にBearer) |
| expires\_in | integer | 有効期限(秒) |

レスポンス例

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 604800
}
```

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 401 | INVALID\_CREDENTIALS | メールアドレスまたはパスワードが正しくない |

### GET /tasks(タスク一覧取得) {#get-/tasks(タスク一覧取得)}

リクエストパラメータ

| 名前 | 型 | 必須 | 説明 |
| :---- | :---- | :---- | :---- |
| status | string | \- | todo/in\_progress/doneで絞り込み |
| parent\_id | integer | \- | 指定した親タスクの子タスクのみ取得(未指定時は全件、null指定で親タスクのみ) |
| list\_id | integer | \- | 指定したリストに所属するタスクのみ取得(未指定時は全件、null指定でどのリストにも属さないタスクのみ) |
| q | string | \- | titleまたはdescriptionに部分一致するキーワードで絞り込み(大文字小文字は区別しない) |
| sort | string | \- | 並び替え。`due_date_asc`/`due_date_desc`/`estimated_hours_asc`/`estimated_hours_desc`のいずれか(未指定時は`created_at`の降順) |
| page | integer | \- | ページ番号 |
| per\_page | integer | \- | 1ページあたりの件数 |

レスポンス(200)

| 名前 | 型 | 説明 |
| :---- | :---- | :---- |
| id | integer | タスクID |
| parent\_id | integer/null | 親タスクのID |
| list\_id | integer/null | 所属するリストのID(親タスクのみ設定可) |
| title | string | タイトル |
| description | string/null | 詳細説明 |
| status | string | ステータス(todo/in\_progress/done) |
| estimated\_hours | number | 想定の工数(時間単位、0.5〜9999の小数) |
| actual\_hours | number/null | 実績の工数(時間単位、0.5〜9999の小数) |
| due\_date | string(date)/null | 締切日 |
| completed\_at | string/null | 完了日時 |
| created\_at | string | 作成日時 |
| updated\_at | string | 更新日時 |

レスポンス例

```json
{
  "data": [
    {
      "id": 10,
      "parent_id": null,
      "list_id": null,
      "title": "DevOpsキャンプ最終課題",
      "description": "タスク管理アプリを作る",
      "status": "in_progress",
      "estimated_hours": 5,
      "actual_hours": null,
      "due_date": "2026-09-20",
      "completed_at": null,
      "created_at": "2026-09-01T09:00:00Z",
      "updated_at": "2026-09-05T09:00:00Z"
    }
  ],
  "meta": { "total_count": 1, "page": 1, "per_page": 20, "total_pages": 1 }
}
```

### POST /tasks(タスク作成) {#post-/tasks(タスク作成)}

リクエストパラメータ

| 名前 | 型 | 必須 | 説明 |
| :---- | :---- | :---- | :---- |
| title | string | ○ | タスクのタイトル |
| description | string | \- | 詳細説明 |
| parent\_id | integer | \- | 親タスクのID(子タスクを作成する場合) |
| list\_id | integer | \- | 所属させるリストのID(親タスクのみ指定可。`parent_id`と同時指定は不可) |
| estimated\_hours | number | ○ | 想定の工数(時間単位、0.5刻みの小数、0.5〜9999) |
| due\_date | string(date) | \- | 締切日 |

リクエスト例

```json
{
  "title": "ER図を作成する",
  "parent_id": 10,
  "estimated_hours": 1.5,
  "due_date": "2026-09-10"
}
```

レスポンス(201): GET /tasksの1件分と同じ形式のオブジェクトを返す

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 422 | ESTIMATED\_HOURS\_REQUIRED | estimated\_hoursが未指定 |
| 422 | VALIDATION\_ERROR | estimated\_hoursが0.5〜9999の範囲外 |
| 422 | VALIDATION\_ERROR | due\_dateがYYYY-MM-DD形式でない |
| 422 | INVALID\_PARENT\_ID | parent\_idが存在しない、または他ユーザーのタスク |
| 422 | INVALID\_LIST\_ID | list\_idが存在しない、または他ユーザーのリスト |
| 422 | LIST\_ID\_NOT\_ALLOWED\_FOR\_CHILD\_TASK | parent\_idとlist\_idを同時に指定した |

### GET /tasks/{id}(タスク詳細取得) {#get-/tasks/{id}(タスク詳細取得)}

レスポンス(200): GET /tasksの1件分と同じ形式(必要に応じて子タスクの一覧をchildrenとして含める)

```json
{
  "id": 10,
  "parent_id": null,
  "list_id": null,
  "title": "DevOpsキャンプ最終課題",
  "status": "in_progress",
  "estimated_hours": 5,
  "actual_hours": null,
  "children": [
    { "id": 11, "title": "ER図を作成する", "status": "todo" }
  ]
}
```

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 404 | TASK\_NOT\_FOUND | タスクが存在しない、または他ユーザーのタスク |

### PATCH /tasks/{id}(タスク更新) {#patch-/tasks/{id}(タスク更新)}

リクエストパラメータ(更新したいフィールドのみ送信)

| 名前 | 型 | 必須 | 説明 |
| :---- | :---- | :---- | :---- |
| title | string | \- | タイトル |
| description | string | \- | 詳細説明 |
| status | string | \- | todo/in\_progress/done |
| list\_id | integer | \- | 所属させるリストのID(親タスクのみ指定可) |
| actual\_hours | number | \- | 実績の工数(完了時に入力、時間単位、0.5刻みの小数、0.5〜9999) |
| due\_date | string(date) | \- | 締切日 |

リクエスト例(完了処理)

```json
{ "status": "done", "actual_hours": 4.5 }
```

※ `status`を`done`に更新するタイミングでサーバー側が`completed_at`に現在日時を自動設定する。逆に`done`から`todo`/`in_progress`に戻した場合、`completed_at`は`null`に戻る。

レスポンス(200): 更新後のタスクオブジェクト

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 422 | ACTUAL\_HOURS\_REQUIRED | statusをdoneに更新する際にactual\_hoursが未指定 |
| 422 | VALIDATION\_ERROR | actual\_hoursが0.5〜9999の範囲外、またはtitleを空文字にしようとした、またはdue\_dateがYYYY-MM-DD形式でない |
| 422 | INVALID\_LIST\_ID | list\_idが存在しない、または他ユーザーのリスト |
| 422 | LIST\_ID\_NOT\_ALLOWED\_FOR\_CHILD\_TASK | 子タスクに対してlist\_idを指定した |
| 404 | TASK\_NOT\_FOUND | タスクが存在しない、または他ユーザーのタスク |

---

### DELETE /tasks/{id}(タスク削除) {#delete-/tasks/{id}(タスク削除)}

レスポンス: 204 No Content(子タスクが存在する場合は子タスクも合わせて削除する)

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 404 | TASK\_NOT\_FOUND | タスクが存在しない、または他ユーザーのタスク |

### GET /tasks/calendar {#get-/tasks/calendar}

リクエストパラメータ(クエリ)

| 名前 | 型 | 必須 | 説明 |
| :---- | :---- | :---- | :---- |
| from | string(date, YYYY-MM-DD) | \- | 集計開始日。未指定時は当日が属する年の1/1以降で最初に来る日曜日 |
| to | string(date, YYYY-MM-DD) | \- | 集計終了日。未指定時はfromの370日後(53週間/371日分) |

パラメータなしで全期間を返すと、利用が長期化した際にレスポンスサイズが線形に増えてしまうため、未指定時は53週間(371日)分をデフォルト範囲として返す。起点を「1/1以降で最初に来る日曜日」にしているのは、GitHub風ヒートマップの曜日列(行)を年によらず一定に保つため([ADR-0010](../adr/0010-calendar-date-range-params.md)参照)。

レスポンス(200)

| 名前 | 型 | 説明 |
| :---- | :---- | :---- |
| date | string(date) | 対象日(例: "2026-09-01") |
| total\_hours | number | その日に完了したタスクのactual\_hoursの合計値 |

レスポンス例

```json
{
  "data": [
    { "date": "2026-09-01", "total_hours": 8 },
    { "date": "2026-09-02", "total_hours": 3.5 }
  ]
}
```

### GET /lists(リスト一覧取得) {#get-/lists(リスト一覧取得)}

タスクを横断的にグルーピングする「リスト」の一覧を取得する([ADR-0021](../adr/0021-task-list-grouping.md))。ページネーションは行わず、作成日時の昇順で全件返す。

レスポンス(200)

| 名前 | 型 | 説明 |
| :---- | :---- | :---- |
| id | integer | リストID |
| name | string | リスト名 |
| created\_at | string | 作成日時 |
| updated\_at | string | 更新日時 |

レスポンス例

```json
{
  "data": [
    { "id": 1, "name": "プロジェクトA", "created_at": "2026-09-01T09:00:00Z", "updated_at": "2026-09-01T09:00:00Z" }
  ]
}
```

### POST /lists(リスト作成) {#post-/lists(リスト作成)}

リクエストパラメータ

| 名前 | 型 | 必須 | 説明 |
| :---- | :---- | :---- | :---- |
| name | string | ○ | リスト名 |

リクエスト例

```json
{ "name": "プロジェクトA" }
```

レスポンス(201): GET /listsの1件分と同じ形式のオブジェクトを返す

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 422 | VALIDATION\_ERROR | nameが空文字または未指定 |

### PATCH /lists/{id}(リスト更新) {#patch-/lists/{id}(リスト更新)}

リクエストパラメータ

| 名前 | 型 | 必須 | 説明 |
| :---- | :---- | :---- | :---- |
| name | string | \- | リスト名 |

レスポンス(200): 更新後のリストオブジェクト

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 422 | VALIDATION\_ERROR | nameを空文字にしようとした |
| 404 | LIST\_NOT\_FOUND | リストが存在しない、または他ユーザーのリスト |

### DELETE /lists/{id}(リスト削除) {#delete-/lists/{id}(リスト削除)}

リストを削除する。所属していたタスクは削除されず、`list_id`が`null`に戻る(`ON DELETE SET NULL`、[ADR-0021](../adr/0021-task-list-grouping.md))。

レスポンス(204): ボディなし

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 404 | LIST\_NOT\_FOUND | リストが存在しない、または他ユーザーのリスト |

エラーレスポンス

| ステータスコード | エラーコード | エラー内容 |
| :---- | :---- | :---- |
| 400 | INVALID\_DATE\_FORMAT | from/toがYYYY-MM-DD形式でない |
| 422 | INVALID\_DATE\_RANGE | fromがtoより後の日付になっている |
