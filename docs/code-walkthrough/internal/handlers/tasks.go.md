# internal/handlers/tasks.go

`TaskHandler`が`/tasks`系5エンドポイント+カレンダー集計を実装する。このリポジトリで最も大きいファイルであり、共通ヘルパーと各エンドポイントの実装に分かれる。

## 共通部品

- `currentUserID(w, r)`: `auth.UserIDFromContext`でリクエストコンテキストから認証ユーザーIDを取り出す。取れなければ`401 UNAUTHORIZED`を書いて`ok=false`を返す。全エンドポイントの先頭で呼ばれる
- `scanTask(row)`: DBの1行を`models.Task`に変換する。NULL許容カラムは`sql.NullXxx`型で受けてから、値があるときだけポインタフィールドにセットする(`internal/models`のポインタ設計とDBのNULLを橋渡しする層)
- `taskColumns`: SELECT対象カラムの列挙を定数化し、`scanTask`のスキャン順序と一致させている
- `loadTask(r, id, userID)`: `WHERE id = ? AND user_id = ?`で1件取得。**他ユーザーのタスクは常に「見つからない」として扱われる**ため、この関数が所有者チェックを実質的に兼ねている
- `parseTaskID(r)`: パスパラメータ`{id}`を`int64`にパース

## エンドポイントごとの実装

### `List` (`GET /tasks`)

- `page`/`per_page`をクエリから取得(デフォルト1/20、`per_page`は最大100)
- `status`・`parent_id`(`"null"`指定で親タスクのみ、というAPI仕様書の挙動を含む)でWHERE句を動的に組み立て
- `COUNT(*)`で総件数を取得してから本体を`LIMIT`/`OFFSET`で取得し、`meta`(total_count等)付きで返す

### `Create` (`POST /tasks`)

- `title`必須(トリム後空文字はNG)、`estimated_weight`必須(nilなら`422 ESTIMATED_WEIGHT_REQUIRED`)
- `due_date`は指定時に`time.Parse`で`YYYY-MM-DD`形式かを検証(不正なら`422 VALIDATION_ERROR`)
- `parent_id`指定時は、そのタスクが存在し**かつ自分の所有物であること**をSELECTで確認してから作成([ADR-0007](../../../adr/0007-self-referencing-task-hierarchy.md))
- 作成後は`loadTask`で読み直してから返す(DBのデフォルト値やAUTO_INCREMENTのIDを含めて正確なレスポンスを作るため)

### `Get` (`GET /tasks/{id}`)

- 対象タスクを`loadTask`で取得(存在しない/他ユーザーのものなら`404 TASK_NOT_FOUND`)
- 追加で`parent_id = 対象ID`の子タスク一覧を取得し、レスポンスの`children`フィールドに含める(API仕様書の「子タスク・孫タスク」表示に対応)
- このエンドポイントだけレスポンスを`models.Task`ではなく`map[string]any`で手動組み立てしている(`children`を追加するため)

### `Update` (`PATCH /tasks/{id}`)

- 送られてきたフィールドだけを動的にSET句に組み込む部分更新(PATCHセマンティクス)
- `title`を送る場合はCreateと同じくトリム後空文字を`422 VALIDATION_ERROR`で拒否。`due_date`もCreateと同じ`time.Parse`検証を通す(作成時と更新時で許容する値がずれないようにするため)
- `status`を`done`に変更する瞬間(`completingNow`)だけ`actual_weight`必須というビジネスルールを強制し([ADR-0008](../../../adr/0008-separate-estimated-actual-weight.md))、同時に`completed_at`を現在時刻で自動セットする
- 既に`done`のタスクを再度`done`にしても`completingNow`は`false`になる(`existing.Status != models.StatusDone`の条件があるため、`completed_at`が上書きされない)
- 逆に`done`から`todo`/`in_progress`に戻す(`uncompletingNow`)と`completed_at`を`NULL`に戻す。「未完了なのに完了日時が残る」という矛盾を防ぐための仕様

### `Delete` (`DELETE /tasks/{id}`)

- `DELETE FROM tasks WHERE id = ? AND user_id = ?`。`RowsAffected() == 0`なら`404 TASK_NOT_FOUND`
- 子タスクは`schema.sql`の`ON DELETE CASCADE`によりDB側で連動削除される(アプリケーションコードでは削除しない)

### `Calendar` (`GET /tasks/calendar`)

- `parseCalendarRange`で`from`/`to`を解決(下記参照)
- `completed_at`の日付でGROUP BYし、`actual_weight`の合計を日別に集計して返す([ADR-0010](../../../adr/0010-calendar-date-range-params.md))

## `parseCalendarRange`

- `from`/`to`が未指定なら「当日から365日前〜当日」がデフォルト
- 指定されている場合は`YYYY-MM-DD`形式でパース。形式不正は`errInvalidDateFormat`(→`400`)、`from`が`to`より後ろなら`errInvalidDateRange`(→`422`)を返す
- 日付比較用のエラー型を`errors.New`でパッケージ変数として定義し、呼び出し側は`errors.Is`で判定している(Go標準のエラー判定パターン)

## 読むときのポイント

- 「他ユーザーのタスクは404」という一貫した挙動(`403 Forbidden`を使わない)は、タスクの存在自体を他ユーザーに知らせないための意図的な設計。API仕様書のエラー説明にも「タスクが存在しない、または他ユーザーのタスク」と明記されている
- SQL文字列を`strings.Join`で動的組み立てしている箇所(`List`のWHERE句、`Update`のSET句)は、すべてプレースホルダ(`?`)+`args`スライスで値を渡しており、値そのものを文字列連結していないためSQLインジェクションの心当たりはない
- `Update`が「送られたフィールドだけ更新する」ためにポインタ型のリクエスト構造体(`updateTaskRequest`)を使っている。これは`internal/models`のNULL表現と同じ「ポインタ=未指定/設定済みの区別」というパターンの再利用
