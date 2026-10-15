package chat

import (
	"context"
	"database/sql"
	json "encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"9router/proxy/internal/db"
)

// overflowBody is the wording an upstream uses when the prompt does not fit the
// model's context window. It is a request/model mismatch, not a credential
// problem, so the combo must move to the next member and leave the account
// alone.
const overflowBody = `{"error":{"message":"This model's maximum context length is 8192 tokens. However, your messages resulted in 21000 tokens.","type":"invalid_request_error","code":"context_length_exceeded"}}`

// overflowUpstream answers a context-overflow 400 for the small model and a
// normal completion for the large one, counting what it was asked.
func overflowUpstream(t *testing.T, hits *atomic.Int32, smallModel string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read upstream request: %v", err)
			return
		}
		var req struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Errorf("decode upstream request: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if req.Model == smallModel {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(overflowBody))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-1","object":"chat.completion","created":0,"model":"` + req.Model + `","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// overflowComboEnv seeds one deepseek account pointed at upstream and a combo of
// two deepseek models, so the only thing that can rescue a 400 on the first
// member is the combo advancing to the second.
func overflowComboEnv(t *testing.T, upstream string, models []string) (*sql.DB, func()) {
	t.Helper()
	database, cleanup := setupChatTestDB(t)
	// The helper's own deepseek/groq rows would shadow the account under test.
	if _, err := database.Exec(`DELETE FROM providerConnections WHERE id IN ('conn-1', 'conn-2')`); err != nil {
		t.Fatalf("clear seeded connections: %v", err)
	}
	seedConnDB(t, database, "deepseek", "conn-small", "sk-small", upstream)
	modelsJSON, _ := json.Marshal(models)
	if _, err := database.Exec(`INSERT INTO combos (id, name, kind, models, createdAt, updatedAt) VALUES ('combo-fit', 'combo-fit', 'fallback', ?, '2026-07-18T00:00:00Z', '2026-07-18T00:00:00Z')`, string(modelsJSON)); err != nil {
		t.Fatalf("seed combo: %v", err)
	}
	return database, cleanup
}

// assertNoCooldown fails when a context overflow left the account parked. The
// account did nothing wrong: parking it is how the fix under test would look
// correct to a request count while still degrading every later request.
func assertNoCooldown(t *testing.T, database *sql.DB, connID string) {
	t.Helper()
	var rawData string
	if err := database.QueryRow(`SELECT data FROM providerConnections WHERE id = ?`, connID).Scan(&rawData); err != nil {
		t.Fatalf("read connection row: %v", err)
	}
	if until, ok := db.ConnectionCooldownUntil(rawData); ok {
		t.Errorf("context overflow cooled the account down until %v", until)
	}
	locked, err := db.NewRepo(database).IsConnectionModelLocked(connID, "deepseek-chat")
	if err != nil {
		t.Fatalf("IsConnectionModelLocked: %v", err)
	}
	if locked {
		t.Error("context overflow locked the account for deepseek-chat")
	}
}

// TestComboContextOverflow_MovesToNextModel pins the OpenAI combo contract: a
// model whose context window is too small must not end the request. The combo
// advances to the next member — once, not ten times against the same dead end.
func TestComboContextOverflow_MovesToNextModel(t *testing.T) {
	var hits atomic.Int32
	models := []string{"deepseek/deepseek-chat", "deepseek/deepseek-reasoner"}
	srv := overflowUpstream(t, &hits, "deepseek-chat")

	database, cleanup := overflowComboEnv(t, srv.URL, models)
	defer cleanup()

	h := NewChatHandler(db.NewRepo(database))
	rec := httptest.NewRecorder()
	h.handleComboFallback(context.Background(), rec,
		[]byte(`{"model":"combo-fit","messages":[{"role":"user","content":"hi"}]}`),
		models, "fallback", false, false, "combo-fit", 1)

	if rec.Code != http.StatusOK {
		t.Fatalf("client status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	// One 400 then one success. Without the break the loop keeps re-selecting
	// the same connection for the same model and spends ten rejected requests
	// before it ever reaches the second member.
	if got := hits.Load(); got != 2 {
		t.Errorf("upstream received %d requests, want 2 (1 overflow + 1 success)", got)
	}
	assertNoCooldown(t, database, "conn-small")
}

// TestComboContextOverflow_MessagesEndpoint pins the same contract on the Claude
// /v1/messages combo loop, which carries its own copy of the failover rules.
func TestComboContextOverflow_MessagesEndpoint(t *testing.T) {
	var hits atomic.Int32
	models := []string{"deepseek/deepseek-chat", "deepseek/deepseek-reasoner"}
	srv := overflowUpstream(t, &hits, "deepseek-chat")

	database, cleanup := overflowComboEnv(t, srv.URL, models)
	defer cleanup()

	h := NewChatHandler(db.NewRepo(database))
	translatedReq := map[string]any{
		"model":      "combo-fit",
		"max_tokens": 100,
		"messages":   []map[string]any{{"role": "user", "content": "hi"}},
	}
	rec := httptest.NewRecorder()
	h.handleMessagesComboFallback(context.Background(), rec, translatedReq, models, "fallback", false, "combo-fit", 1)

	if rec.Code != http.StatusOK {
		t.Fatalf("client status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("upstream received %d requests, want 2 (1 overflow + 1 success)", got)
	}
	assertNoCooldown(t, database, "conn-small")
}

// TestComboContextOverflow_AllMembersOverflow surfaces the real cause rather
// than a synthesized failure. Every member was too small, so the client gets the
// upstream's own 400 back — which is what tells it (and the operator) the turn
// needs a bigger context, not a working account.
func TestComboContextOverflow_AllMembersOverflow(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(overflowBody))
	}))
	defer srv.Close()

	models := []string{"deepseek/deepseek-chat", "deepseek/deepseek-reasoner"}
	database, cleanup := overflowComboEnv(t, srv.URL, models)
	defer cleanup()

	h := NewChatHandler(db.NewRepo(database))
	rec := httptest.NewRecorder()
	h.handleComboFallback(context.Background(), rec,
		[]byte(`{"model":"combo-fit","messages":[{"role":"user","content":"hi"}]}`),
		models, "fallback", false, false, "combo-fit", 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("client status = %d, want 400 carried through from the upstream", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "maximum context length") {
		t.Errorf("body = %q, want the upstream's own overflow reason", rec.Body.String())
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("upstream received %d requests, want 2 (one per member, no repeat pass)", got)
	}
	assertNoCooldown(t, database, "conn-small")
}

// TestComboContextOverflow_Unrelated4xxStillCooldown guards the neighbouring
// behaviour: only the context-overflow shape is exempt. A 400 that says nothing
// about fit keeps whatever handling it had, so the exemption cannot become a
// blanket escape hatch for bad requests.
func TestComboContextOverflow_Unrelated4xxStillCooldown(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"invalid request: unknown parameter 'foo'"}}`))
	}))
	defer srv.Close()

	models := []string{"deepseek/deepseek-chat", "deepseek/deepseek-reasoner"}
	database, cleanup := overflowComboEnv(t, srv.URL, models)
	defer cleanup()

	h := NewChatHandler(db.NewRepo(database))
	rec := httptest.NewRecorder()
	h.handleComboFallback(context.Background(), rec,
		[]byte(`{"model":"combo-fit","messages":[{"role":"user","content":"hi"}]}`),
		models, "fallback", false, false, "combo-fit", 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("client status = %d, want the upstream 400", rec.Code)
	}
	// The exemption is model-fit specific: an unrelated 400 still walks the
	// per-entry connection budget before moving on.
	if got := hits.Load(); got <= 2 {
		t.Errorf("upstream received %d requests, want an unrelated 400 to keep walking its connections", got)
	}
}