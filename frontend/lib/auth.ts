// JWTトークンをlocalStorageに保存・取得するヘルパー。
// バックエンドはAuthorization: Bearerヘッダー方式(Cookieなし、ADR-0004)のため、
// フロントエンド側でトークンを保持する必要がある。

const TOKEN_KEY = "access_token";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  window.localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  window.localStorage.removeItem(TOKEN_KEY);
}

export function isLoggedIn(): boolean {
  return getToken() !== null;
}

export function logout(): void {
  clearToken();
  // コンポーネント外からも呼べるよう、ルーターではなくハードリロードで遷移する。
  // ログアウト時はReactの状態を丸ごと破棄したいので、むしろ都合が良い。
  // eslint-disable-next-line @next/next/no-location-assign-relative-destination
  window.location.href = "/login";
}
