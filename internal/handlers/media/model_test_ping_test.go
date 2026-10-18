package media

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"9router/proxy/internal/handlerutil"
)

func TestProbeContextFlag(t *testing.T) {
	bg := context.Background()
	if handlerutil.IsProbeContext(bg) {
		t.Error("expected background context to not be probe context")
	}

	probeCtx := handlerutil.WithProbeContext(bg)
	if !handlerutil.IsProbeContext(probeCtx) {
		t.Error("expected probe context to report true")
	}
}

func TestReadProbeResult_BlockedCooldownResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusBadGateway)
	rec.Write([]byte(`{"error":{"message":"no available connections for provider: deepseek (all in cooldown, earliest reset 2026-10-09T01:00:00Z)","code":502}}`))

	res := readProbeResult(rec, "")
	if res.OK {
		t.Error("expected OK to be false for 502")
	}
	if !res.Blocked {
		t.Error("expected Blocked to be true for all in cooldown")
	}
	if res.ResetAt != "2026-10-09T01:00:00Z" {
		t.Errorf("ResetAt = %q, want 2026-10-09T01:00:00Z", res.ResetAt)
	}
}

func TestReadProbeResult_StandardErrorIsNotBlocked(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusInternalServerError)
	rec.Write([]byte(`{"error":{"message":"upstream error: internal failure"}}`))

	res := readProbeResult(rec, "")
	if res.OK {
		t.Error("expected OK to be false")
	}
	if res.Blocked {
		t.Error("expected Blocked to be false for standard 500 error")
	}
	if res.ResetAt != "" {
		t.Errorf("expected empty ResetAt, got: %q", res.ResetAt)
	}
}

func TestReadProbeResult_MidStreamTimeoutInconclusive(t *testing.T) {
	// 1. SSE stream emitting a data: prefix then hanging until deadline -> ok = true (success on first token)
	rec1 := httptest.NewRecorder()
	rec1.WriteHeader(http.StatusOK)
	rec1.Write([]byte(`data: {"choices":[{"delta":{"content":"hello"}}]} `))

	res1 := readProbeResult(rec1, "")
	if !res1.OK {
		t.Errorf("expected OK=true for stream with data prefix, got OK=false, error=%q", res1.Error)
	}

	// 2. Probe that never receives an answer within the window -> ok = false with empty error string (inconclusive)
	rec2 := httptest.NewRecorder()
	rec2.WriteHeader(499)
	rec2.Write([]byte(``))

	res2 := readProbeResult(rec2, "")
	if res2.OK {
		t.Error("expected OK=false for empty aborted probe")
	}
	if res2.Error != "" {
		t.Errorf("expected empty error string for inconclusive probe, got %q", res2.Error)
	}
}
