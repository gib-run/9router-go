package db

import (
	"testing"
)

func TestApiKeyTokenLimitAndUsage(t *testing.T) {
	r := openEphemeralSchemaDB(t)

	if err := r.CreateApiKey("100", "sk-token-limit-test", "limit key", "m1"); err != nil {
		t.Fatalf("CreateApiKey: %v", err)
	}

	key, err := r.GetApiKeyByID("100")
	if err != nil {
		t.Fatalf("GetApiKeyByID: %v", err)
	}
	if key.TokenLimit != nil {
		t.Errorf("fresh TokenLimit = %v, want nil", key.TokenLimit)
	}
	if key.UsedTokens != 0 {
		t.Errorf("fresh UsedTokens = %d, want 0", key.UsedTokens)
	}

	limit := int64(5000)
	if err := r.SetApiKeyTokenLimit("100", &limit); err != nil {
		t.Fatalf("SetApiKeyTokenLimit: %v", err)
	}

	key, err = r.GetApiKeyByID("100")
	if err != nil {
		t.Fatalf("GetApiKeyByID after set limit: %v", err)
	}
	if key.TokenLimit == nil || *key.TokenLimit != 5000 {
		t.Fatalf("TokenLimit = %v, want 5000", key.TokenLimit)
	}

	// Increment tokens using int64 id
	if err := r.IncrementKeyUsedTokens(100, 1500); err != nil {
		t.Fatalf("IncrementKeyUsedTokens: %v", err)
	}

	key, err = r.GetApiKeyByID("100")
	if err != nil {
		t.Fatalf("GetApiKeyByID: %v", err)
	}
	if key.UsedTokens != 1500 {
		t.Errorf("UsedTokens = %d, want 1500", key.UsedTokens)
	}

	// Increment tokens using string id
	if err := r.IncrementKeyUsedTokensByID("100", 2000); err != nil {
		t.Fatalf("IncrementKeyUsedTokensByID: %v", err)
	}

	key, err = r.GetApiKeyByID("100")
	if err != nil {
		t.Fatalf("GetApiKeyByID: %v", err)
	}
	if key.UsedTokens != 3500 {
		t.Errorf("UsedTokens = %d, want 3500", key.UsedTokens)
	}

	// Reset tokens using int64 id
	if err := r.ResetKeyUsedTokens(100); err != nil {
		t.Fatalf("ResetKeyUsedTokens: %v", err)
	}

	key, err = r.GetApiKeyByID("100")
	if err != nil {
		t.Fatalf("GetApiKeyByID: %v", err)
	}
	if key.UsedTokens != 0 {
		t.Errorf("UsedTokens after reset = %d, want 0", key.UsedTokens)
	}
}
