# frontend

見積もり振り返り型タスク管理ツールのフロントエンド(Step6)。Next.js(App Router) + TypeScript + Tailwind CSS。バックエンドAPIは[../README.md](../README.md)、画面設計は[../docs/SCREEN_LIST.md](../docs/SCREEN_LIST.md)・[../docs/SCREEN_FLOW.md](../docs/SCREEN_FLOW.md)・[../docs/assignment/step5.md](../docs/assignment/step5.md)を参照。

## セットアップ

1. 依存パッケージをインストールする

   ```sh
   npm install
   ```

2. バックエンドAPIの接続先を設定する(任意、未設定時は`http://localhost:8080`を使う)

   `frontend/.env.local`を作成し、以下を記載する(`.env.local`はgitignore対象)。

   ```
   NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
   ```

3. 開発サーバーを起動する(別ターミナルでバックエンド`go run .`も起動しておく)

   ```sh
   npm run dev
   ```

   [http://localhost:3000](http://localhost:3000) を開く。

## 静的解析・ビルド確認

```sh
npx tsc --noEmit   # 型チェック
npm run lint       # ESLint
npm run build      # 本番ビルド
```

## ディレクトリ構成

- `app/` — ルーティング(App Router)。画面ごとのディレクトリ構成は実装を進めながら追記する
- `lib/api.ts` — バックエンドAPIを呼ぶ薄いクライアント(fetchのラッパー、`docs/API_SPEC.md`に対応)
- `lib/auth.ts` — JWTトークンの保存・取得(localStorage、ADR-0004参照)
- `lib/types.ts` — APIのリクエスト/レスポンスに対応する型定義
