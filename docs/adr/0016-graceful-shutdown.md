# 0016: Graceful shutdownを実装する

## Status

Accepted

## Context

`main.go`はサーバーを起動したら`http.ListenAndServe`(または[ADR-0015](./0015-http-server-timeouts.md)後は`srv.ListenAndServe`)がプロセス終了までブロックし続けるだけで、終了シグナルを受けたときに処理中のリクエストを待たずに強制終了する構造だった。Step3提出後のレビューで、将来コンテナ化(Step9-10)やデプロイのたびにプロセスが再起動されることを踏まえ、処理中のリクエストを取りこぼさない仕組みが必要だと指摘を受けた。

## Decision

- `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`で終了シグナルを待ち受けるコンテキストを作る
- サーバー本体(`srv.ListenAndServe()`)は別goroutineで起動し、メインの流れをブロックしない
- シグナルを受けたら`srv.Shutdown(ctx)`を呼び、新規リクエストの受付を止めつつ、最大10秒だけ処理中のリクエストの完了を待ってから終了する

## Consequences

- `Ctrl+C`やコンテナのSIGTERM送信時に、処理中のリクエストを中断させずに終了できる
- `main`関数の構造が「サーバーを起動してブロックする」から「別goroutineで起動し、シグナル待ちしてからシャットダウンする」に変わった
- シャットダウンの猶予時間(10秒)を超えて処理が終わらないリクエストがあった場合は、`Shutdown`がタイムアウトしてエラーを返す(現状は`log.Fatalf`でプロセスを終了させている)
