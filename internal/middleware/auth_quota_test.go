package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"9router/proxy/internal/db"
	"9router/proxy/internal/models"
)

func TestCheckApiKeyQuota(t *testing.T) {
	limit500 := int64(500)
	limit0 := int64(0)

	tests := []struct {
		name    string
		key     *models.APIKey
		wantErr bool
	}{
		{
			name:    "nil key returns nil",
			key:     nil,
			wantErr: false,
		},
		{
			name: "nil token limit returns nil",
			key: &models.APIKey{
				UsedTokens: 1000,
			},
			wantErr: false,
		},
		{
			name: "zero token limit returns nil (unlimited)",
			key: &models.APIKey{
				TokenLimit: &limit0,
				UsedTokens: 1000,
			},
			wantErr: false,
		},
		{
			name: "under quota returns nil",
			key: &models.APIKey{
				TokenLimit: &limit500,
				UsedTokens: 499,
			},
			wantErr: false,
		},
		{
			name: "exact quota reached returns error",
			key: &models.APIKey{
				TokenLimit: &limit500,
				UsedTokens: 500,
			},
			wantErr: true,
		},
		{
			name: "quota exceeded returns error",
			key: &models.APIKey{
				TokenLimit: &limit500,
				UsedTokens: 600,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckApiKeyQuota(tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckApiKeyQuota() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireApiKeyQuotaEnforcement(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Seed key with quota exceeded (limit 1000, used 1500)
	_, err := database.Exec(`INSERT INTO apiKeys (id, key, name, machineId, isActive, createdAt, token_limit, used_tokens) VALUES
		('quota-exceeded', 'sk-quota-exceeded', 'Quota Exceeded', 'mac', 1, '2026-07-18T00:00:00Z', 1000, 1500),
		('quota-ok', 'sk-quota-ok', 'Quota OK', 'mac', 1, '2026-07-18T00:00:00Z', 1000, 500);`)
	if err != nil {
		t.Fatalf("failed to seed apiKeys: %v", err)
	}

	repo := db.NewRepo(database)
	handler := RequireApiKey(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	// 1. Key with exceeded quota should receive HTTP 429
	req1 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req1.Header.Set("Authorization", "Bearer sk-quota-exceeded")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429 Too Many Requests (body: %s)", rec1.Code, rec1.Body.String())
	}

	var errResp struct {
		Error struct {
			Code    string `json:"code"`
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec1.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if errResp.Error.Code != "key_quota_exceeded" {
		t.Errorf("error.code = %q, want key_quota_exceeded", errResp.Error.Code)
	}

	// 2. Key with available quota should succeed
	req2 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req2.Header.Set("Authorization", "Bearer sk-quota-ok")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 OK (body: %s)", rec2.Code, rec2.Body.String())
	}
}
