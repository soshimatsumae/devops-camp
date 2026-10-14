"use client";

import { logout } from "@/lib/auth";

// S-03の共通ヘッダー。ログアウトはここからのみ行える
// (モーダル表示中は背後のヘッダーを押せないため、docs/SCREEN_FLOW.md参照)。
export default function AppHeader() {
  return (
    <header className="flex items-center justify-end px-4 pt-2">
      <nav>
        <button
          type="button"
          onClick={logout}
          className="rounded-md border border-dashed border-zinc-300 px-3 py-1 text-sm text-zinc-600 hover:bg-zinc-50"
        >
          ログアウト
        </button>
      </nav>
    </header>
  );
}
