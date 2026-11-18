# Go APIのコンテナイメージ。マルチステージビルドで、ビルド用と実行用の2段に分ける。
# 1段目(build)にはGoのコンパイラ一式が入るが、最終的なイメージには2段目(実行用)だけが残る。

# ---- 1段目: ビルド ----
FROM golang:1.26.5-alpine AS build
WORKDIR /src

# 依存ライブラリの取得を先に行う。go.mod・go.sumが変わらない限り、
# この層はキャッシュが使われ、コードを直しただけの再ビルドが速くなる
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal ./internal

RUN go build -o /out/api .

# ---- 2段目: 実行用 ----
FROM alpine:3.22

# root以外の専用ユーザーで動かす。万一アプリが乗っ取られても、コンテナ内でできることを減らすため
RUN adduser -D -u 10001 app
USER app

COPY --from=build /out/api /usr/local/bin/api

EXPOSE 8080
ENTRYPOINT ["api"]
