import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // npm run build のときに、動かすのに必要なファイルだけを .next/standalone に集める。
  // コンテナのイメージにnode_modulesを丸ごと入れずに済み、イメージが小さくなる(frontend/Dockerfile参照)
  output: "standalone",
  cacheComponents: true,
  partialPrefetching: true,
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
};

export default nextConfig;
