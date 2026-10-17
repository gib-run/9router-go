package middleware

import (
	"9router/proxy/internal/log"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"9router/proxy/internal/db"
	"9router/proxy/internal/handlerutil"
	"9router/proxy/internal/keikey"
	"9router/proxy/internal/models"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

// ApiKeyContextKey is the context key for the authenticated API key object.
const ApiKeyContextKey ContextKey = "apiKey"

// ErrKeyQuotaExceeded is returned when an API key has reached its token limit.
var ErrKeyQuotaExceeded = errors.New("key_quota_exceeded")

// CheckApiKeyQuota checks if the API key has exceeded its configured token limit.
func CheckApiKeyQuota(key *models.APIKey) error {
	if key == nil {
		return nil
	}
	if key.TokenLimit != nil && *key.TokenLimit > 0 && key.UsedTokens >= *key.TokenLimit {
		return ErrKeyQuotaExceeded
	}
	return nil
}

// RequireApiKey creates a middleware handler that authenticates requests using client API keys.
// It checks the Authorization header (Bearer <key>) and the query parameter `key`.
// Valid keys are retrieved from the SQLite database; inactive or disabled keys are rejected with 401.
func RequireApiKey(repo *db.Repo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKeyString := ExtractApiKey(r)
			if apiKeyString == "" {
				handlerutil.WriteJSONError(w, http.StatusUnauthorized, "Authentication required. Provide an API key via Authorization: Bearer <key> header or ?key=<key> query parameter.")
				return
			}

			// Validate via SQLite repository and retrieve details
			lookup := keikey.LookupHash(apiKeyString)
			apiKeyObj, err := resolveApiKey(repo, lookup, apiKeyString)
			if err != nil {
				log.Error("auth", "DB lookup error", "error", err)
				handlerutil.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
			if apiKeyObj == nil {
				handlerutil.WriteJSONError(w, http.StatusUnauthorized, "Invalid API key.")
				return
			}

			if !verifyApiKey(apiKeyString, apiKeyObj) {
				handlerutil.WriteJSONError(w, http.StatusUnauthorized, "Invalid API key.")
				return
			}

			// Only a verified key enters the cache, and only for 5s: a
			// revoked or deleted key cannot outlive that window.
			if lookup != "" {
				keikey.DefaultAuthCache().Set(lookup, apiKeyObj)
			}

			if apiKeyObj.IsActive != 1 {
				handlerutil.WriteJSONError(w, http.StatusUnauthorized, "Invalid or inactive API key.")
				return
			}

			// F-14: Check key expiry
			if apiKeyObj.ExpiresAt != nil && *apiKeyObj.ExpiresAt != "" {
				expires, err := time.Parse(time.RFC3339, *apiKeyObj.ExpiresAt)
				if err == nil && time.Now().After(expires) {
					handlerutil.WriteJSONError(w, http.StatusUnauthorized, "API key expired.")
					return
				}
			}

			// Check token quota
			if err := CheckApiKeyQuota(apiKeyObj); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]any{
						"message": "API key token quota exceeded",
						"type":    "key_quota_exceeded",
						"code":    "key_quota_exceeded",
					},
				})
				return
			}

			// F-14 usage accounting. Best-effort: the row write must never turn
			// a served request into a failure, and a failed write only means the
			// resale bookkeeping undercounts by one.
			//
			// It runs after every check that can reject, so a request refused for
			// a bad key, a disabled key, or an expired contract is never counted
			// as usage.
			if apiKeyObj.ID != "" {
				if uErr := repo.UpdateApiKeyUsage(apiKeyObj.ID); uErr != nil {
					log.Warn("auth", "api key usage update failed", "key", apiKeyObj.ID, "error", uErr)
				}
			}

			// Inject API Key info into the request context for downstream handlers/logging
			ctx := context.WithValue(r.Context(), ApiKeyContextKey, apiKeyObj)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// resolveApiKey finds the row for an incoming key. Since issue #199 the
// plaintext column is the primary path again, matching upstream; the argon2id
// lookup remains only so a row created under F-6 keeps authenticating after the
// upgrade. There is no self-heal in either direction: rewriting a row would
// destroy the one property the dashboard now depends on, which is that the
// secret can be read back.
func resolveApiKey(repo *db.Repo, lookup, plaintext string) (*models.APIKey, error) {
	if cached, ok := keikey.DefaultAuthCache().Get(lookup); ok {
		return cached.(*models.APIKey), nil
	}
	apiKey, err := repo.GetApiKeyByKey(plaintext)
	if err != nil || apiKey == nil {
		if err != nil {
			return nil, err
		}
		return repo.FindApiKeyByLookup(lookup)
	}
	return apiKey, nil
}

// verifyApiKey checks the secret against the row. A row written before issue
// #199 carries an argon2id verifier and is confirmed with it; every other row
// was matched on the plaintext column itself, so there is nothing left to check.
func verifyApiKey(plaintext string, apiKey *models.APIKey) bool {
	if apiKey.KeyHash == nil || *apiKey.KeyHash == "" {
		return true
	}
	return keikey.Verify(plaintext, *apiKey.KeyHash)
}

// GetAuthenticatedApiKey retrieves the authenticated APIKey object from the request context.
func GetAuthenticatedApiKey(r *http.Request) *models.APIKey {
	if r == nil {
		return nil
	}
	return GetAuthenticatedApiKeyFromContext(r.Context())
}

// GetAuthenticatedApiKeyFromContext retrieves the authenticated APIKey object from a context.
func GetAuthenticatedApiKeyFromContext(ctx context.Context) *models.APIKey {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(ApiKeyContextKey)
	if val == nil {
		return nil
	}
	keyObj, ok := val.(*models.APIKey)
	if !ok {
		return nil
	}
	return keyObj
}

// ExtractApiKey extracts the client API key from the request.
// For standard API routes, only header-based auth is accepted to avoid leaks.
// For EventSource / SSE stream endpoints (where browsers cannot set custom headers),
// query parameters `key` or `apiKey` are accepted as a fallback.
func ExtractApiKey(r *http.Request) string {
	// 1. Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1])
		}
	}

	// 2. Try custom X-API-Key header as fallback
	if xApiKey := r.Header.Get("X-API-Key"); xApiKey != "" {
		return xApiKey
	}

	// 3. For EventSource / SSE endpoints, accept query parameter key or apiKey
	if strings.HasSuffix(r.URL.Path, "/stream") {
		if qKey := r.URL.Query().Get("key"); qKey != "" {
			return qKey
		}
		if qKey := r.URL.Query().Get("apiKey"); qKey != "" {
			return qKey
		}
	}

	return ""
}
