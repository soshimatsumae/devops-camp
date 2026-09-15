# internal/models/models.go

APIレスポンス・DB行のGo表現となる構造体を定義するだけのパッケージ。ビジネスロジックは持たない。

## やっていること

- `User`構造体: `id`/`email`/`password_hash`/`name`/`created_at`/`updated_at`。`PasswordHash`だけ`json:"-"`が付いており、JSONへ絶対にシリアライズされない
- ステータス定数: `StatusTodo` / `StatusInProgress` / `StatusDone`(それぞれ`"todo"`/`"in_progress"`/`"done"`)
- `Task`構造体: DB仕様書([docs/DB_SPEC.md](../../../DB_SPEC.md))の`tasks`テーブルに対応。NULL許容カラムはすべてポインタ型(`*int64`, `*string`, `*int`, `*time.Time`)で表現している
- `CalendarEntry`構造体: `GET /tasks/calendar`のレスポンス1件分(`date`と`total_weight`)

## 読むときのポイント

- `ParentID *int64`のようにポインタ型を使うことで、「0(ゼロ値)」と「未設定(NULL)」を区別できる。もし`int64`のままだと親タスクなしの状態を`0`と誤認しかねない
- `UserID`は`json:"-"`ではないが実質サーバー内部専用(APIレスポンスの`Task`は個別にフィールドを組み立てて返しているため、この構造体をそのままJSONエンコードする箇所は`List`ハンドラのみ)
- DBとのマッピング(NULL値の変換)は`internal/handlers/tasks.go`の`scanTask`関数が担っており、この構造体自体はマッピングのロジックを持たない
