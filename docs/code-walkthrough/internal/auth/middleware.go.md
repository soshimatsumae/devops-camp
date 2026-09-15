# internal/auth/middleware.go

`Authorization`ヘッダーの検証とユーザーIDのコンテキスト注入を行う認証ミドルウェア。

## やっていること

- `contextKey`型 + `userIDKey`定数: コンテキストのキー衝突を避けるための非公開型(Goの定石)
- `RequireAuth(secret) func(http.HandlerFunc) http.HandlerFunc`: デコレータを返す関数
  1. `Authorization`ヘッダーから`"Bearer "`プレフィックスを`strings.CutPrefix`で取り除く。プレフィックスがない/トークンが空なら`401 UNAUTHORIZED`
  2. `ParseToken(secret, token)`で検証。失敗したら`401 UNAUTHORIZED`
  3. 成功したらユーザーIDを`context.WithValue`でリクエストコンテキストに詰め、次のハンドラを呼ぶ
- `UserIDFromContext(ctx)`: コンテキストからユーザーIDを取り出すgetter。型アサーションが失敗したら`ok=false`
- `ContextWithUserID(ctx, userID)`: `RequireAuth`を経由せずハンドラを直接呼ぶテストのために、同じコンテキストキーへ直接値を設定できるexported関数

## 読むときのポイント

- `net/http`のみで実装する制約([ADR-0002](../../../adr/0002-net-http-only.md))の中で、ミドルウェアチェインを「`http.HandlerFunc`を受け取って`http.HandlerFunc`を返す関数」として自前実装している。`main.go`では`requireAuth(taskHandler.List)`のように使う
- `ContextWithUserID`は本体のロジックには使われず、テストコード(`internal/handlers/testhelpers_test.go`)専用の後付けexport。ハンドラを`RequireAuth`でラップせず直接呼ぶユニットテストのために、ミドルウェアが本来やる「コンテキストへのユーザーID注入」を代替している
- 認証失敗時のエラーコードは常に文字列`"UNAUTHORIZED"`で統一されており、ヘッダー欠落・トークン無効・期限切れを区別しない(`jwt.go`の設計方針と同じ)
