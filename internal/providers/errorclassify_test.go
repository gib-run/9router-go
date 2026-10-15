package providers

import (
	"net/http"
	"testing"
)

func TestClassifyError_CapacityExhausted(t *testing.T) {
	// Status 503 with MODEL_CAPACITY_EXHAUSTED
	cls := ClassifyError(http.StatusServiceUnavailable, "No capacity available for model gemini-3.8-flash-tiered on the server MODEL_CAPACITY_EXHAUSTED", 0)
	if !cls.ShouldFallback {
		t.Error("expected fallback for capacity exhausted error")
	}
	if cls.NewBackoffLevel != 1 {
		t.Errorf("expected backoff level 1, got %d", cls.NewBackoffLevel)
	}
	if cls.CooldownMs != 2000 {
		t.Errorf("expected 2000ms cooldown, got %d", cls.CooldownMs)
	}
}

func TestClassifyError_ResourceExhausted429(t *testing.T) {
	cls := ClassifyError(http.StatusTooManyRequests, "Resource has been exhausted (e.g. check quota). RESOURCE_EXHAUSTED", 0)
	if !cls.ShouldFallback {
		t.Error("expected fallback for 429 resource exhausted")
	}
	if cls.NewBackoffLevel != 1 {
		t.Errorf("expected backoff level 1, got %d", cls.NewBackoffLevel)
	}
}

func TestClassifyError_Status503Fallback(t *testing.T) {
	// Plain 503 without specific matching text should trigger backoff via status rule
	cls := ClassifyError(http.StatusServiceUnavailable, "", 0)
	if !cls.ShouldFallback {
		t.Error("expected fallback for 503")
	}
	if cls.NewBackoffLevel != 1 {
		t.Errorf("expected backoff level 1 for 503, got %d", cls.NewBackoffLevel)
	}
}

func TestClassifyError_RequestScoped4xxNoFallback(t *testing.T) {
	// Upstream parity (checkFallbackError): request-scoped 4xx must not lock the account.
	for _, status := range []int{http.StatusBadRequest, http.StatusMethodNotAllowed, http.StatusConflict, http.StatusUnprocessableEntity} {
		cls := ClassifyError(status, "bad request body", 0)
		if cls.ShouldFallback {
			t.Errorf("expected no fallback for %d, got %+v", status, cls)
		}
	}
	// Account-scoped statuses keep fallback.
	for _, status := range []int{http.StatusUnauthorized, http.StatusPaymentRequired, http.StatusForbidden, http.StatusTooManyRequests} {
		cls := ClassifyError(status, "", 0)
		if !cls.ShouldFallback {
			t.Errorf("expected fallback for %d, got %+v", status, cls)
		}
	}
	// 404 keeps its rule (upstream NON_RETRIABLE + status rule).
	cls := ClassifyError(http.StatusNotFound, "", 0)
	if !cls.ShouldFallback {
		t.Errorf("expected fallback for 404, got %+v", cls)
	}
}

func TestIsModelFitError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		text   string
		want   bool
	}{
		{"maximum context length 400", http.StatusBadRequest, "This model's maximum context length is 8192 tokens", true},
		{"openai style code", http.StatusBadRequest, `{"error":{"code":"context_length_exceeded"}}`, true},
		{"google exceed context limit", http.StatusBadRequest, "Please reduce the length of the messages: exceed context limit", true},
		{"anthropic token overflow", http.StatusBadRequest, "prompt is too long: 210000 tokens > 200000 maximum", true},
		{"context window exceeded", http.StatusBadRequest, "Context window exceeded for this model", true},
		{"too many input tokens", http.StatusBadRequest, "too many input tokens: 300000", true},
		{"unrelated 400", http.StatusBadRequest, "invalid api key provided", false},
		{"unrelated 422", http.StatusUnprocessableEntity, "messages is required", false},
		{"429 is account scoped", http.StatusTooManyRequests, "rate limit exceeded", false},
		{"5xx is server scoped", http.StatusServiceUnavailable, "maximum context length is 8192 tokens", false},
		{"200 is not an error", http.StatusOK, "maximum context length", false},
		{"empty text", http.StatusBadRequest, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModelFitError(tt.status, tt.text); got != tt.want {
				t.Errorf("IsModelFitError(%d, %q) = %v, want %v", tt.status, tt.text, got, tt.want)
			}
		})
	}
}

// A context overflow is a request/model mismatch, never an account fault, so it
// must not earn the account a cooldown through the generic fallback rules.
func TestIsModelFitError_IsRequestScopedNotAccountScoped(t *testing.T) {
	cls := ClassifyError(http.StatusBadRequest, "This model's maximum context length is 8192 tokens", 0)
	if cls.ShouldFallback {
		t.Errorf("context overflow must not cool the account down, got %+v", cls)
	}
}
