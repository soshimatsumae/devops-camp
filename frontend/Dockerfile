# Next.jsのコンテナイメージ。Go APIと同じく、マルチステージビルドでビルド用と実行用の2段に分ける。
# next.config.ts の output: "standalone" で、ビルド結果のうち動かすのに必要なものだけを .next/standalone に集め、
# 実行用の段にはそれだけをコピーする。

# ---- 1段目: ビルド ----
FROM node:22.12.0-alpine AS build
WORKDIR /app

# パッケージのインストールを先に行う。package.json・package-lock.jsonが変わらない限り、
# この層はキャッシュが使われ、コードを直しただけの再ビルドが速くなる(Go APIのgo mod downloadと同じ考え方)
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# NEXT_PUBLIC_ で始まる環境変数は、実行時ではなくビルド時に画面のJavaScriptへ埋め込まれる。
# そのため、APIの接続先はビルド時に決める。ローカルでは、ブラウザから直接 localhost:8080 のapiコンテナへ送る
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
# Next.jsの利用状況の送信を止める
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# ---- 2段目: 実行用 ----
FROM node:22.12.0-alpine
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
# コンテナの外(ポート転送)から届くよう、すべてのネットワークで待ち受ける
ENV HOSTNAME=0.0.0.0

# standaloneの出力には、画像などの public/ と、CSS・JSなどの .next/static/ が含まれないため、別にコピーする。
# nodeイメージに最初からある一般ユーザー node を持ち主にする
COPY --from=build --chown=node:node /app/.next/standalone ./
COPY --from=build --chown=node:node /app/.next/static ./.next/static
COPY --from=build --chown=node:node /app/public ./public

# root以外のユーザーで動かす(nodeイメージには node ユーザーが用意されている)
USER node

EXPOSE 3000
CMD ["node", "server.js"]
