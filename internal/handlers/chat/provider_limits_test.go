package chat

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"9router/proxy/internal/db"
	"9router/proxy/internal/middleware"
)

// Provider-scope enforcement, request side. These tests cover the properties the
// plan promises and that nothing else can prove: a refusal is a failed attempt
// rather than an ending, a request remembers which providers were full, a refusal
// never parks a healthy account, and an unreadable policy fails open.

// seedLimitedProviderNode inserts a custom provider node carrying a policy.
func seedLimitedProviderNode(t *testing.T, repo *db.Repo, id string, limits db.ProviderLimits) {
	t.Helper()
	if _, err := repo.CreateProviderNode(id, "openai-compatible", id,
		`{"prefix":"`+id+`","apiType":"openai-compatible","baseUrl":"https://example.test/v1"}`,
	); err != nil {
		t.Fatalf("CreateProviderNode %s: %v", id, err)
	}
	if err := repo.SetProviderLimits(context.Background(), id, limits); err != nil {
		t.Fatalf("SetProviderLimits %s: %v", id, err)
	}
}

// newProviderLimitHandler builds a chat handler with a real repository and a real
// limiter, because the enforcement under test is exactly the seam between them.
func newProviderLimitHandler(t *testing.T) (*ChatHandler, *db.Repo) {
	t.Helper()
	database, cleanup := setupChatTestDB(t)
	t.Cleanup(cleanup)
	repo := db.NewRepo(database)
	handler := NewChatHandler(repo)
	handler.RateLimiter = middleware.NewRateLimiter(time.Minute)
	return handler, repo
}

func TestAcquireProviderAttempt_UnlimitedProviderReturnsNoSlot(t *testing.T) {
	handler, _ := newProviderLimitHandler(t)
	ctx := withProviderLimitState(context.Background())

	// A registry provider (no providerNodes row) is the common case: every
	// request to deepseek, groq, claude and so on goes through here.
	slot, err := handler.acquireProviderAttempt(ctx, "deepseek", "deepseek-chat", "conn-1", []byte(`{"messages":[]}`))
	if slot != nil || err != nil {
		t.Fatalf("unlimited provider = slot %v err %v, want nil/nil", slot, err)
	}
}

func TestAcquireProviderAttempt_RefusalIsARateLimitErrorNamingTheAttempt(t *testing.T) {
	handler, repo := newProviderLimitHandler(t)
	seedLimitedProviderNode(t, repo, "node-nara", db.ProviderLimits{RPM: 1})
	ctx := withProviderLimitState(context.Background())

	slot, err := handler.acquireProviderAttempt(ctx, "node-nara", "glm-5.3", "conn-1", []byte(`{"messages":[]}`))
	if slot == nil || err != nil {
		t.Fatalf("first attempt = slot %v err %v, want an admitted slot", slot, err)
	}
	slot.Release()

	_, err = handler.acquireProviderAttempt(ctx, "node-nara", "glm-5.3", "conn-1", []byte(`{"messages":[]}`))
	if err == nil {
		t.Fatal("second attempt was admitted while the provider's window was full")
	}

	// The existing writers answer a *upstreamError, and a provider refusal has to
	// reach them with the 429 and a readable body rather than a 502.
	var ue *upstreamError
	if !errors.As(err, &ue) {
		t.Fatalf("refusal = %T, want something errors.As can read as *upstreamError", err)
	}
	if ue.StatusCode != 429 {
		t.Errorf("refusal status = %d, want 429", ue.StatusCode)
	}
	body := string(ue.Body)
	for _, want := range []string{`"attempts"`, `"scope":"provider"`, "node-nara", `"reason":"rpm"`} {
		if !strings.Contains(body, want) {
			t.Errorf("refusal body misses %s: %s", want, body)
		}
	}

	attempt, ok := providerLimitRefusal(err)
	if !ok {
		t.Fatalf("providerLimitRefusal(%v) = not a provider refusal", err)
	}
	if attempt.Provider != "node-nara" || attempt.Axis != "rpm" || attempt.Limit != 1 {
		t.Errorf("attempt = %+v, want node-nara/rpm/1", attempt)
	}
}

// TestAcquireProviderAttempt_RefusedProviderIsNotWaitedOnTwice is the failover
// cost the request-scoped state exists to pay once: the first refusal waits out
// the queue timeout, and every later attempt on that provider must return at once
// so the pass can reach its other candidates.
func TestAcquireProviderAttempt_RefusedProviderIsNotWaitedOnTwice(t *testing.T) {
	handler, repo := newProviderLimitHandler(t)
	seedLimitedProviderNode(t, repo, "node-nara", db.ProviderLimits{RPM: 1, QueueTimeoutMs: 120})
	ctx := withProviderLimitState(context.Background())

	slot, err := handler.acquireProviderAttempt(ctx, "node-nara", "glm-5.3", "conn-1", []byte(`{}`))
	if slot == nil || err != nil {
		t.Fatalf("first attempt = slot %v err %v, want an admitted slot", slot, err)
	}
	slot.Release()

	start := time.Now()
	if _, err := handler.acquireProviderAttempt(ctx, "node-nara", "glm-5.3", "conn-1", []byte(`{}`)); err == nil {
		t.Fatal("expected a refusal once the window was full")
	}
	first := time.Since(start)
	if first < 100*time.Millisecond {
		t.Errorf("first refusal took %v, want it to wait out the 120ms queue timeout", first)
	}

	start = time.Now()
	if _, err := handler.acquireProviderAttempt(ctx, "node-nara", "glm-5.3", "conn-1", []byte(`{}`)); err == nil {
		t.Fatal("expected the provider to stay refused for the rest of the request")
	}
	if second := time.Since(start); second > 60*time.Millisecond {
		t.Errorf("second refusal took %v, want it to return at once from the request's own memory", second)
	}
}

// TestAcquireProviderAttempt_MemoisesThePolicyForTheRequest: a combo that tries
// three models on one node reads the policy once, so an operator raising the limit
// mid-request does not change what this pass enforces. The next request reads the
// new value.
func TestAcquireProviderAttempt_MemoisesThePolicyForTheRequest(t *testing.T) {
	handler, repo := newProviderLimitHandler(t)
	seedLimitedProviderNode(t, repo, "node-nara", db.ProviderLimits{RPM: 1})
	ctx := withProviderLimitState(context.Background())

	slot, err := handler.acquireProviderAttempt(ctx, "node-nara", "glm-5.3", "conn-1", []byte(`{}`))
	if slot == nil || err != nil {
		t.Fatalf("first attempt = slot %v err %v, want an admitted slot", slot, err)
	}
	slot.Release()

	if err := repo.SetProviderLimits(context.Background(), "node-nara", db.ProviderLimits{RPM: 500}); err != nil {
		t.Fatalf("SetProviderLimits: %v", err)
	}
	if _, err := handler.acquireProviderAttempt(ctx, "node-nara", "glm-5.3", "conn-1", []byte(`{}`)); err == nil {
		t.Error("the same request re-read the policy; the pass it started with should hold")
	}

	fresh := withProviderLimitState(context.Background())
	if slot, err := handler.acquireProviderAttempt(fresh, "node-nara", "glm-5.3", "conn-1", []byte(`{}`)); slot == nil || err != nil {
		t.Errorf("a new request = slot %v err %v, want it to pick up the raised limit", slot, err)
	}
}

// TestAcquireProviderAttempt_FailsOpenWhenThePolicyCannotBeRead: an unreadable
// policy must not take an endpoint offline. Treating the read failure as "no
// headroom" would turn one storage blip into an outage for every custom provider.
func TestAcquireProviderAttempt_FailsOpenWhenThePolicyCannotBeRead(t *testing.T) {
	handler, _ := newProviderLimitHandler(t)
	ctx := withProviderLimitState(context.Background())
	if err := handler.Repo.RawDB().Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	slot, err := handler.acquireProviderAttempt(ctx, "node-nara", "glm-5.3", "conn-1", []byte(`{}`))
	if slot != nil || err != nil {
		t.Fatalf("unreadable policy = slot %v err %v, want the attempt to proceed", slot, err)
	}
}

// TestRecordProviderLimit_DebugTrailWithoutACooldown pins the two halves of D7.1:
// the refusal is visible (connection lastError, so the node-health table shows it)
// and it does NOT park the account (no cooldown), because the endpoint was busy for
// a moment and the credential is fine.
func TestRecordProviderLimit_DebugTrailWithoutACooldown(t *testing.T) {
	handler, repo := newProviderLimitHandler(t)
	seedLimitedProviderNode(t, repo, "node-nara", db.ProviderLimits{RPM: 1})
	if err := repo.CreateProviderConnectionFull("conn-nara", "node-nara", "apikey", "Nara",
		nil, `{"apiKey":"sk-test-nara"}`); err != nil {
		t.Fatalf("CreateProviderConnectionFull: %v", err)
	}
	ctx := withProviderLimitState(context.Background())
	attempt := providerLimitAttempt{Provider: "node-nara", Axis: "rpm", Limit: 1, RetryAfter: 3}

	handler.recordProviderLimit(ctx, providerLimitStateFrom(ctx), "node-nara", "glm-5.3", "conn-nara", attempt)

	conn, err := repo.GetProviderConnectionByID("conn-nara")
	if err != nil {
		t.Fatalf("GetProviderConnectionByID: %v", err)
	}
	if conn == nil {
		t.Fatal("connection disappeared")
	}
	if !strings.Contains(conn.Data, "rpm") {
		t.Errorf("lastError does not name the axis that refused: %s", conn.Data)
	}
	if _, parked := db.ConnectionCooldownUntil(conn.Data); parked {
		t.Error("a provider limit refusal parked the connection in cooldown; a busy endpoint is not a dead credential")
	}

	attempts := providerLimitStateFrom(ctx).snapshotAttempts()
	if len(attempts) != 1 || attempts[0].Provider != "node-nara" {
		t.Errorf("debug trail = %+v, want one entry for node-nara", attempts)
	}
}
