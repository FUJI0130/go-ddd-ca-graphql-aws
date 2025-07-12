package handler

import (
	"encoding/json"
	"net/http"
)

// HealthHandler はヘルスチェックを処理するハンドラーです
type HealthHandler struct {
	Version string
}

// NewHealthHandler はHealthHandlerの新しいインスタンスを作成します
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		Version: version,
	}
}

// HealthResponse はヘルスチェックのレスポンス構造体です
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

// Check はヘルスチェックエンドポイントを処理します
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "UP",
		Version: h.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
