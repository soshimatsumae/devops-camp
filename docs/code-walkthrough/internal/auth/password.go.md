# internal/auth/password.go

`bcrypt`のラッパーのみの2関数構成([ADR-0005](../../../adr/0005-bcrypt-password-hashing.md))。

## やっていること

- `HashPassword(password string) (string, error)`: `bcrypt.GenerateFromPassword`を`bcrypt.DefaultCost`で呼び出し、ハッシュ文字列を返す
- `CheckPassword(hash, password string) bool`: `bcrypt.CompareHashAndPassword`の結果が`nil`(一致)かどうかをboolで返す

## 読むときのポイント

- `bcrypt.DefaultCost`(現行ライブラリでは10)を明示的に指定せずデフォルトのまま使っている。コストを上げると総当たり耐性は上がるがハッシュ計算時間も増える(ログイン・登録のレイテンシに直結)ため、要件が変わらない限りデフォルトのままで問題ない
- `CheckPassword`がエラーの中身を握りつぶして`bool`にしているのは、呼び出し側(`handlers.AuthHandler.Login`)が「一致した/しなかった」の2値だけ知れれば十分なため。ハッシュ形式不正などの内部エラーも「不一致」として扱われる
