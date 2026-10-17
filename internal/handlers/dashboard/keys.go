package dashboard

import (
	"net/http"
	"strconv"

	"9router/proxy/internal/handlerutil"
	"9router/proxy/internal/keikey"
)

// HandleResetKeyQuota handles POST /api/keys/{id}/reset-quota.
func (h *DashboardHandler) HandleResetKeyQuota(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if id == "" {
		handlerutil.WriteJSONError(w, http.StatusBadRequest, "missing apiKey id")
		return
	}
	existing, err := h.Repo.GetApiKeyByID(id)
	if err != nil {
		handlerutil.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		handlerutil.WriteJSONError(w, http.StatusNotFound, "api key not found")
		return
	}

	var resetErr error
	if intID, err := strconv.ParseInt(id, 10, 64); err == nil {
		resetErr = h.Repo.ResetKeyUsedTokens(intID)
	} else {
		resetErr = h.Repo.ResetKeyUsedTokensByID(id)
	}
	if resetErr != nil {
		handlerutil.WriteJSONError(w, http.StatusInternalServerError, resetErr.Error())
		return
	}

	keikey.DefaultAuthCache().InvalidateByID(id)
	handlerutil.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"id":     id,
	})
}

// ResetKeyQuota is an alias to HandleResetKeyQuota.
func (h *DashboardHandler) ResetKeyQuota(w http.ResponseWriter, r *http.Request) {
	h.HandleResetKeyQuota(w, r)
}
