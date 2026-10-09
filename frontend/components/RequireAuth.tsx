"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, type ReactNode } from "react";
import { isLoggedIn } from "@/lib/auth";

// S-03・M-01〜M-04はすべて認証必須(docs/SCREEN_FLOW.md参照)。
// トークンを持たない状態でアクセスしたらS-01へ強制リダイレクトする。
// トークン期限切れ(401)側の強制リダイレクトはlib/api.tsのrequest()が担当する。
export default function RequireAuth({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!isLoggedIn()) {
      router.replace("/login");
      return;
    }
    setReady(true);
  }, [router]);

  if (!ready) return null;

  return <>{children}</>;
}
