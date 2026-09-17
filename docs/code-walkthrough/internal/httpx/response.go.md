# internal/httpx/response.go

JSON応答の共通ヘルパー。全ハンドラがレスポンス・エラーレスポンスをここ経由で書き出す。

## やっていること

- `ErrorDetail` / `ErrorBody`: エラーレスポンスの型。`{"error":{"code":"...","message":"..."}}`という形を型として表現している([ADR-0009](../../../adr/0009-unified-error-response-format.md))
- `WriteJSON(w, status, v)`: `Content-Type: application/json`ヘッダーを設定し、ステータスコードを書き、`v`をJSONエンコードして書き出す汎用関数。`v`が`nil`ならボディなし(`204 No Content`用)
- `WriteError(w, status, code, message)`: `ErrorBody`を組み立てて`WriteJSON`に渡すだけのラッパー
- `WriteInternalError(w, err)`: `500`エラー専用のラッパー。`slog.Error`で元の`err`をサーバー側のログにだけ出力し、クライアントには詳細を含まない汎用メッセージ(`"internal server error"`)を返す。全ハンドラの`500`エラーはこの関数経由に統一されている

## 読むときのポイント

- `json.NewEncoder(w).Encode(v)`のエラーを`_`で握りつぶしている。これはレスポンスヘッダー・ステータスコードを書き込んだ後に発生するエラー(クライアント切断等)であり、この時点でエラーを返してもクライアントには伝えようがないため
- 全ハンドラのエラー応答がこの1箇所を経由することで、[ADR-0009](../../../adr/0009-unified-error-response-format.md)で定めたレスポンス形式の一貫性が保たれている
- 以前は`500`エラーごとに`"failed to create user"`のような個別メッセージをクライアントに返していたが、実装の詳細を外部に漏らさないため`WriteInternalError`導入時に汎用メッセージへ統一した。元のエラー内容は`slog.Error`でサーバー側のログにのみ残るため、動作確認時に500が出ても`internal server error`としか見えないのは意図した挙動(原因追跡はログを見る)
