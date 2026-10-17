package executor

import (
	json "encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"9router/proxy/internal/providers"
)

func TestForwardOpencodeGo_DeepSeekStripsResponseFormat(t *testing.T) {
	var gotResponseFormat any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		gotResponseFormat = parsed["response_format"]
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	cfg := &providers.ProviderConfig{BaseURL: srv.URL}

	t.Run("drops response_format for deepseek family", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := &Request{
			Client: srv.Client(),
			Config: cfg,
			APIKey: "key",
			Body:   []byte(`{"model":"deepseek-chat","response_format":{"type":"json_object"},"messages":[{"role":"user","content":"hi"}]}`),
		}
		err := ForwardOpencodeGo(rec, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotResponseFormat != nil {
			t.Errorf("expected response_format to be dropped for deepseek model, got %v", gotResponseFormat)
		}
	})

	t.Run("leaves response_format for non-deepseek models", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := &Request{
			Client: srv.Client(),
			Config: cfg,
			APIKey: "key",
			Body:   []byte(`{"model":"gpt-4o","response_format":{"type":"json_object"},"messages":[{"role":"user","content":"hi"}]}`),
		}
		err := ForwardOpencodeGo(rec, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rfMap, ok := gotResponseFormat.(map[string]any)
		if !ok || rfMap["type"] != "json_object" {
			t.Errorf("expected response_format to be preserved for non-deepseek model, got %v", gotResponseFormat)
		}
	})
}

func TestInjectReasoningContent_OpencodeZen(t *testing.T) {
	input := []byte(`{
		"model": "deepseek-chat",
		"messages": [
			{"role": "user", "content": "hello"},
			{"role": "assistant", "content": "hi there"}
		]
	}`)

	res := InjectReasoningContent(input, "antigravity-zen")

	var reqMap map[string]any
	if err := json.Unmarshal(res, &reqMap); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	msgs := reqMap["messages"].([]any)
	asst := msgs[1].(map[string]any)
	if rc, ok := asst["reasoning_content"].(string); !ok || rc != " " {
		t.Errorf("expected reasoning_content to be ' ' for antigravity-zen, got %v", asst["reasoning_content"])
	}
}
