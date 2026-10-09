import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "見積もり振り返りタスク管理",
  description: "見積もりと実績を記録して振り返るタスク管理ツール",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="ja" className="h-full antialiased">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
