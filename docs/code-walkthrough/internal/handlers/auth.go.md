# internal/handlers/auth.go

`AuthHandler`が`POST /auth/register`と`POST /auth/login`を実装する([docs/API_SPEC.md](../../../API_SPEC.md)のAUTH-01/02に対応)。

## やっていること

### `Register`

1. JSONデコード失敗 → `400 INVALID_JSON`
2. `email`/`name`をトリムして空チェック、`password`が8文字未満 → `422 VALIDATION_ERROR`
3. `auth.HashPassword`でbcryptハッシュ化
4. `INSERT INTO users`。MySQLのユニーク制約違反(エラー番号`1062`)を`errors.As`で検知し`422 EMAIL_ALREADY_REGISTERED`として返す
5. 作成したユーザーの`created_at`を改めて`SELECT`して`201`でレスポンス

### `Login`

1. JSONデコード失敗 → `400 INVALID_JSON`
2. `email`でユーザー検索。見つからない場合と、見つかったがパスワード不一致の場合を同じ条件式でまとめて`401 INVALID_CREDENTIALS`
3. それ以外のDBエラーは`500 INTERNAL_ERROR`
4. 成功したら`auth.GenerateToken`でJWTを発行し`200`でレスポンス

## 読むときのポイント

- MySQLのエラー番号`1062`(Duplicate entry)を`*mysql.MySQLError`型アサーションで検知しているのは、アプリケーション側で事前に「既に存在するか」を`SELECT`してから`INSERT`する(TOCTOU: 確認から使用までの間に競合が起きうる)のではなく、DBのユニーク制約に検知を委ねてレースコンディションを避けるため
- `Login`で「メールアドレスが存在しない」ケースと「パスワードが違う」ケースを区別せず同じ`INVALID_CREDENTIALS`にしているのは、登録済みメールアドレスの有無をエラーメッセージから推測されるユーザー列挙攻撃を防ぐため
- パスワードの最小文字数(8文字)は`Register`のみでチェックしており、`schema.sql`側には制約がない。バリデーションはAPI層に閉じている
