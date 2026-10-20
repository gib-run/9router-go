package dashboard

import (
	"net/http"

	"9router/proxy/internal/handlerutil"
	"9router/proxy/internal/systemmetrics"
)

func (h *DashboardHandler) HandleSystemMetrics(w http.ResponseWriter, r *http.Request) {
	handlerutil.WriteJSON(w, http.StatusOK, systemmetrics.Read())
}
