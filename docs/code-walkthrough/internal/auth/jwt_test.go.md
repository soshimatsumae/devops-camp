# internal/auth/jwt_test.go

`jwt.go`のユニットテスト。DB不要で実行できる。

## テストケース

| テスト名 | 検証内容 |
|---|---|
| `TestGenerateAndParseToken` | 発行したトークンを同じシークレットでパースすると、元のユーザーIDが復元できる(ハッピーパス) |
| `TestParseToken_WrongSecret` | 別のシークレットで署名されたトークンはパース時にエラーになる |
| `TestParseToken_Expired` | TTLに負の値(`-time.Hour`)を渡して即座に期限切れのトークンを作り、パースがエラーになることを確認 |
| `TestParseToken_Garbage` | JWTとして解釈できない文字列(`"not-a-jwt"`)を渡すとエラーになる |

## 読むときのポイント

- `TestParseToken_Expired`が`-time.Hour`をTTLとして渡しているのは、実際に1時間待たずに期限切れの状態を再現するテクニック。`GenerateToken`側は期限切れかどうかをチェックしないので、発行時点で過去の`ExpiresAt`を設定できる
- いずれのテストも成功パス・失敗パスをそれぞれ1つずつ検証しており、`ParseToken`が返すエラーの種類までは区別していない(`jwt.go`のコメント通り、意図的に`ErrInvalidToken`へ丸めているため)
