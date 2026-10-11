package executor

import (
	json "encoding/json/v2"
	"strings"
	"testing"

	"9router/proxy/internal/providers"
)

// TestEnsureClaudeMessages_ThinkingDisplay covers the request-side half of
// upstream's "return Claude thinking text to OpenAI-format clients": an
// OpenAI client asks for reasoning with reasoning_effort (Chat Completions) or
// reasoning.summary (Responses), and that ask has to become
// thinking.display "summarized" before the OpenAI-only keys are stripped —
// otherwise Claude returns signature-only thinking and the client sees nothing.
func TestEnsureClaudeMessages_ThinkingDisplay(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantDisplay string
	}{
		{
			name:        "chat completions effort becomes summarized",
			body:        `{"model":"m","reasoning_effort":"high","messages":[{"role":"user","content":"hi"}]}`,
			wantDisplay: "summarized",
		},
		{
			name:        "minimal effort still means reasoning",
			body:        `{"model":"m","reasoning_effort":"minimal","messages":[{"role":"user","content":"hi"}]}`,
			wantDisplay: "summarized",
		},
		{
			name:        "responses summary becomes summarized",
			body:        `{"model":"m","reasoning":{"summary":"auto"},"messages":[{"role":"user","content":"hi"}]}`,
			wantDisplay: "summarized",
		},
		{
			name:        "no reasoning ask leaves the field alone",
			body:        `{"model":"m","messages":[{"role":"user","content":"hi"}]}`,
			wantDisplay: "",
		},
		{
			name:        "an explicit opt-out asks for no thinking at all",
			body:        `{"model":"m","reasoning_effort":"none","messages":[{"role":"user","content":"hi"}]}`,
			wantDisplay: "",
		},
		{
			name:        "responses summary none is not a request",
			body:        `{"model":"m","reasoning":{"summary":"none"},"messages":[{"role":"user","content":"hi"}]}`,
			wantDisplay: "",
		},
		{
			name:        "an explicit display on the body wins",
			body:        `{"model":"m","thinking":{"display":"summarized"},"messages":[{"role":"user","content":"hi"}]}`,
			wantDisplay: "summarized",
		},
		{
			name:        "an explicit display is not overwritten",
			body:        `{"model":"m","thinking":{"display":"none"},"reasoning_effort":"high","messages":[{"role":"user","content":"hi"}]}`,
			wantDisplay: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := EnsureClaudeMessages([]byte(tt.body), "m")

			wantSummaries := tt.wantDisplay == "summarized"
			if gotSummaries := WantsThinkingSummaries(out); gotSummaries != wantSummaries {
				t.Fatalf("WantsThinkingSummaries = %v, want %v (body %s)", gotSummaries, wantSummaries, out)
			}
			if tt.wantDisplay == "" {
				return
			}
			display := thinkingDisplay(t, out)
			if display != tt.wantDisplay {
				t.Errorf("thinking.display = %q, want %q", display, tt.wantDisplay)
			}
		})
	}
}

// TestEnsureClaudeMessages_StripsOpenAIOnlyFields makes sure capturing the
// intent did not reintroduce the keys the Messages API rejects.
func TestEnsureClaudeMessages_StripsOpenAIOnlyFields(t *testing.T) {
	out := EnsureClaudeMessages(
		[]byte(`{"model":"m","reasoning_effort":"high","stream_options":{"x":1},"store":true,"messages":[{"role":"user","content":"hi"}]}`),
		"m",
	)

	for _, key := range []string{"reasoning_effort", "stream_options", "store"} {
		if hasJSONKey(out, key) {
			t.Errorf("%q survived the conversion: %s", key, out)
		}
	}
	if !WantsThinkingSummaries(out) {
		t.Errorf("thinking.display was dropped along with reasoning_effort: %s", out)
	}
}

// TestWithoutBetaFlag_AnthropicRedactThinking is the other half: the flag that
// asks Anthropic for signature-only thinking has to leave the header when the
// body wants summaries, without touching the shared registry map.
func TestWithoutBetaFlag_AnthropicRedactThinking(t *testing.T) {
	shared := map[string]string{
		"anthropic-version": "2023-06-01",
		"Anthropic-Beta":    "claude-code-20250219," + providers.AnthropicBetaRedactThinking + ",fast-mode-2026-02-01",
	}

	got := providers.WithoutBetaFlag(shared, providers.AnthropicBetaRedactThinking)

	if hasBeta(got["Anthropic-Beta"], providers.AnthropicBetaRedactThinking) {
		t.Errorf("redact-thinking still present in %q", got["Anthropic-Beta"])
	}
	if !hasBeta(got["Anthropic-Beta"], "claude-code-20250219") || !hasBeta(got["Anthropic-Beta"], "fast-mode-2026-02-01") {
		t.Errorf("unrelated flags lost: %q", got["Anthropic-Beta"])
	}
	if got["anthropic-version"] != "2023-06-01" {
		t.Errorf("other headers lost: %v", got)
	}
	if !hasBeta(shared["Anthropic-Beta"], providers.AnthropicBetaRedactThinking) {
		t.Error("the shared registry header map was mutated")
	}

	t.Run("a header without the flag is returned unchanged", func(t *testing.T) {
		plain := map[string]string{"Anthropic-Beta": "fast-mode-2026-02-01"}
		if out := providers.WithoutBetaFlag(plain, providers.AnthropicBetaRedactThinking); out["Anthropic-Beta"] != "fast-mode-2026-02-01" {
			t.Errorf("got %q", out["Anthropic-Beta"])
		}
	})
	t.Run("a missing Anthropic-Beta header is left alone", func(t *testing.T) {
		if out := providers.WithoutBetaFlag(map[string]string{}, "x"); len(out) != 0 {
			t.Errorf("got %v", out)
		}
	})
}

// hasBeta reports whether a comma-separated Anthropic-Beta header carries flag.
func hasBeta(header, flag string) bool {
	for f := range strings.SplitSeq(header, ",") {
		if strings.TrimSpace(f) == flag {
			return true
		}
	}
	return false
}

// thinkingDisplay reads thinking.display out of a converted Claude body.
func thinkingDisplay(t *testing.T, body []byte) string {
	t.Helper()
	var reqMap map[string]any
	if err := json.Unmarshal(body, &reqMap); err != nil {
		t.Fatalf("unmarshal %s: %v", body, err)
	}
	thinking, _ := reqMap["thinking"].(map[string]any)
	display, _ := thinking["display"].(string)
	return display
}

// hasJSONKey reports whether a top-level key is present.
func hasJSONKey(body []byte, key string) bool {
	var reqMap map[string]any
	if err := json.Unmarshal(body, &reqMap); err != nil {
		return false
	}
	_, ok := reqMap[key]
	return ok
}

// TestEnsureClaudeMessages_ThinkingType: an OpenAI client's reasoning_effort
// must yield a Claude `thinking` object that carries `type` — Anthropic
// answers `400 thinking.type: Field required` otherwise.
func TestEnsureClaudeMessages_ThinkingType(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		model      string
		wantType   string
		wantBudget float64
		wantEffort string
	}{
		{"adaptive model", `{"reasoning_effort":"high","messages":[{"role":"user","content":"hi"}]}`, "claude-sonnet-5", "adaptive", 0, "high"},
		{"adaptive minimal clamps to low", `{"reasoning_effort":"minimal","messages":[{"role":"user","content":"hi"}]}`, "claude-sonnet-5", "adaptive", 0, "low"},
		{"budget model", `{"reasoning_effort":"low","max_tokens":8000,"messages":[{"role":"user","content":"hi"}]}`, "claude-haiku-4-5", "enabled", 1024, ""},
		{"budget capped below max_tokens", `{"reasoning_effort":"high","max_tokens":4096,"messages":[{"role":"user","content":"hi"}]}`, "claude-haiku-4-5", "enabled", 4095, ""},
		{"fable 5 is adaptive, not budget", `{"reasoning_effort":"high","messages":[{"role":"user","content":"hi"}]}`, "claude-fable-5", "adaptive", 0, "high"},
		{"responses effort", `{"reasoning":{"effort":"medium"},"messages":[{"role":"user","content":"hi"}]}`, "claude-sonnet-5", "adaptive", 0, "medium"},
		{"explicit type is kept", `{"reasoning_effort":"high","thinking":{"type":"disabled"},"messages":[{"role":"user","content":"hi"}]}`, "claude-sonnet-5", "disabled", 0, ""},
		{"none sends no thinking", `{"reasoning_effort":"none","messages":[{"role":"user","content":"hi"}]}`, "claude-sonnet-5", "", 0, ""},
		{"no ask sends no thinking", `{"messages":[{"role":"user","content":"hi"}]}`, "claude-sonnet-5", "", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := EnsureClaudeMessages([]byte(tt.body), tt.model)
			var m map[string]any
			if err := json.Unmarshal(out, &m); err != nil {
				t.Fatal(err)
			}
			th, _ := m["thinking"].(map[string]any)
			gotType, _ := th["type"].(string)
			if gotType != tt.wantType {
				t.Fatalf("thinking.type = %q, want %q (body %s)", gotType, tt.wantType, out)
			}
			if tt.wantType == "" && th != nil {
				t.Errorf("thinking should be absent: %s", out)
			}
			if tt.wantBudget != 0 {
				if b, _ := th["budget_tokens"].(float64); b != tt.wantBudget {
					t.Errorf("budget_tokens = %v, want %v", b, tt.wantBudget)
				}
			}
			if tt.wantEffort != "" {
				oc, _ := m["output_config"].(map[string]any)
				if oc["effort"] != tt.wantEffort {
					t.Errorf("output_config.effort = %v, want %q", oc["effort"], tt.wantEffort)
				}
			}
			if _, ok := m["reasoning_effort"]; ok {
				t.Errorf("reasoning_effort survived: %s", out)
			}
		})
	}
}
