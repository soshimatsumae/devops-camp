# internal/auth/jwt.go

JWTの発行と検証([ADR-0004](../../../adr/0004-jwt-auth.md))。

## やっていること

- `GenerateToken(secret, userID, ttl)`: `jwt.RegisteredClaims`(`Subject`にユーザーIDを文字列化して格納、`ExpiresAt`/`IssuedAt`を設定)をHS256で署名してトークン文字列を返す
- `ParseToken(secret, tokenStr)`: トークンを検証し、成功したら`Subject`をパースしてユーザーIDを返す
  - 署名アルゴリズムがHMAC以外だったら`ErrInvalidToken`を返して拒否する
  - パース失敗・トークン無効(期限切れ含む)・`Subject`が数値でない、のいずれでも同じ`ErrInvalidToken`にまとめて返す

## 読むときのポイント

- `t.Method.(*jwt.SigningMethodHMAC)`のチェックは、JWTライブラリでよく知られる「アルゴリズム混同攻撃」(`alg: none`や非対称鍵アルゴリズムへの差し替え)を防ぐための実装。鍵の型ごとに許可するアルゴリズムを固定するのが定石
- エラーを`ErrInvalidToken`1種類に丸めているのは、呼び出し側(`RequireAuth`ミドルウェア)がトークンの何が悪いかをクライアントに詳細に伝える必要がない(すべて`401 UNAUTHORIZED`扱い)ため。詳細を伝えるとトークン偽造の試行錯誤に使われるリスクもある
- ユーザーIDを`Subject`(文字列)に入れているのは`jwt.RegisteredClaims`の標準クレームをそのまま使うためで、カスタムクレーム構造体を自前定義していない
