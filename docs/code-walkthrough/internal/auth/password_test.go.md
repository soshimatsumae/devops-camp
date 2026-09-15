# internal/auth/password_test.go

`password.go`のユニットテスト。DB不要。

## テストケース

`TestHashAndCheckPassword`の1本のみ:

1. `HashPassword("correct-horse-battery-staple")`でハッシュを生成
2. 正しいパスワードで`CheckPassword`が`true`を返すことを確認
3. 誤ったパスワード(`"wrong-password"`)で`CheckPassword`が`false`を返すことを確認

## 読むときのポイント

- ハッシュ化・照合という往復(round-trip)を1テストで両方確認するシンプルな構成。`bcrypt`自体の正しさはライブラリ側の責務なので、ここでは「ラッパー関数の配線が正しいか」だけを見ている
