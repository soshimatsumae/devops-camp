# internal/auth/middleware_test.go

`middleware.go`の`RequireAuth`デコレータのユニットテスト。DB不要、`httptest`のみで完結する。

## テストケース

| テスト名 | 検証内容 |
|---|---|
| `TestRequireAuth_MissingHeader` | `Authorization`ヘッダーなしのリクエストは`401`を返し、次のハンドラが呼ばれない(呼ばれたら`t.Error`) |
| `TestRequireAuth_ValidToken` | 正しいトークンを付けたリクエストは次のハンドラに到達し、`UserIDFromContext`で発行時と同じユーザーIDが取れる |
| `TestRequireAuth_InvalidToken` | 壊れたトークン文字列(`"not-a-real-token"`)を付けたリクエストは`401`を返し、次のハンドラが呼ばれない |

## 読むときのポイント

- 「次のハンドラが呼ばれないこと」を`t.Error`をハンドラ内に仕込むことで検証している(呼ばれたらテスト失敗)。戻り値やモックでの検証ではなく、副作用の有無で確認するシンプルな手法
- `TestRequireAuth_ValidToken`は、ミドルウェアが実際にコンテキストへユーザーIDを詰めていることまで確認しており、`RequireAuth`の主要な責務(検証 + コンテキスト注入)の両方をカバーしている
