package dashboard

import (
	"bytes"
	json "encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiKeyResetQuotaAndTokenLimit(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	if err := repo.UpdateSettingsRaw(map[string]any{"requireLogin": false}); err != nil {
		t.Fatalf("disable requireLogin: %v", err)
	}
	router := setupTestRouter(repo)

	// 1. Create a key with a tokenLimit
	createBody, _ := json.Marshal(map[string]any{
		"name":       "Quota Key",
		"tokenLimit": 50000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/keys", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("create key expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var created struct {
		ID         string `json:"id"`
		Key        string `json:"key"`
		TokenLimit *int64 `json:"tokenLimit"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create key response: %v", err)
	}
	keyID := created.ID

	// Verify key in GET /api/keys
	req = httptest.NewRequest(http.MethodGet, "/api/keys", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var keys []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &keys); err != nil {
		t.Fatalf("unmarshal get keys response: %v", err)
	}
	var found map[string]any
	for _, k := range keys {
		if k["id"] == keyID {
			found = k
			break
		}
	}
	if found == nil {
		t.Fatalf("created key %s not found in list", keyID)
	}
	if val, ok := found["tokenLimit"].(float64); !ok || int64(val) != 50000 {
		t.Errorf("tokenLimit = %v, want 50000", found["tokenLimit"])
	}
	if val, ok := found["usedTokens"].(float64); !ok || int64(val) != 0 {
		t.Errorf("usedTokens = %v, want 0", found["usedTokens"])
	}

	// Increment usedTokens in DB
	if err := repo.IncrementKeyUsedTokensByID(keyID, 12345); err != nil {
		t.Fatalf("IncrementKeyUsedTokensByID: %v", err)
	}

	// Verify usedTokens is 12345
	keyObj, err := repo.GetApiKeyByID(keyID)
	if err != nil || keyObj == nil {
		t.Fatalf("GetApiKeyByID failed: %v", err)
	}
	if keyObj.UsedTokens != 12345 {
		t.Fatalf("usedTokens = %d, want 12345", keyObj.UsedTokens)
	}

	// 2. Call POST /api/keys/{id}/reset-quota
	req = httptest.NewRequest(http.MethodPost, "/api/keys/"+keyID+"/reset-quota", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("reset quota expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resetResp struct {
		Status string `json:"status"`
		ID     string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resetResp); err != nil {
		t.Fatalf("unmarshal reset response: %v", err)
	}
	if resetResp.Status != "ok" || resetResp.ID != keyID {
		t.Errorf("reset response = %+v, want status ok and id %s", resetResp, keyID)
	}

	// Verify usedTokens is 0 after reset
	keyObj, err = repo.GetApiKeyByID(keyID)
	if err != nil || keyObj == nil {
		t.Fatalf("GetApiKeyByID after reset failed: %v", err)
	}
	if keyObj.UsedTokens != 0 {
		t.Fatalf("usedTokens after reset = %d, want 0", keyObj.UsedTokens)
	}

	// 3. Update tokenLimit via PUT /api/keys/{id}
	updateBody, _ := json.Marshal(map[string]any{
		"tokenLimit": 100000,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/keys/"+keyID, bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("update key expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	keyObj, err = repo.GetApiKeyByID(keyID)
	if err != nil || keyObj == nil {
		t.Fatalf("GetApiKeyByID after update failed: %v", err)
	}
	if keyObj.TokenLimit == nil || *keyObj.TokenLimit != 100000 {
		t.Errorf("tokenLimit after update = %v, want 100000", keyObj.TokenLimit)
	}
}
