package handlers

import (
	"net/http"

	"github.com/smatsumae/devops-camp/internal/httpx"
)

// Health はALBのヘルスチェック用エンドポイント。認証なしで200を返す。
// DBの疎通は確認しない(RDSのMulti-AZ切り替え中に全台が異常と判定されるのを避けるため。
// docs/assignment/step8.md「講師からの回答」参照)。
func Health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
