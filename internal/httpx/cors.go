package httpx

import "net/http"

// CORS はフロントエンド(別オリジンで動くNext.js開発サーバーなど)からの
// Authorization ヘッダー付きリクエストを許可するミドルウェア。
// 許可するオリジンは allowedOrigins に列挙したものだけに限定し、
// それ以外のオリジンにはAccess-Control-Allow-Originを付けない(=ブラウザ側で拒否される)。
func CORS(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
