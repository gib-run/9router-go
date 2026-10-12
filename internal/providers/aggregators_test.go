package providers

import "testing"

// TestAggregatorProviders pins the five OpenAI-compatible aggregators upstream
// added in v0.5.91 (tokenharbor, dahl, atria, agnes, bai). They are registry
// entries like any other provider, so every table has to carry them: transport,
// alias resolution, executor registration and the offline model catalogue. A
// provider that reaches KnownProviders but not ProviderAliasMap is invisible
// from a `th/<model>` request, and one that reaches neither the dashboard nor
// the executor never routes at all.
func TestAggregatorProviders(t *testing.T) {
	tests := []struct {
		id          string
		baseURL     string
		models      []string
		aliases     []string
		modelsList  string
		wantFetcher bool
		authHeader   string
		authScheme   string
	}{
		{
			id: "tokenharbor", baseURL: "https://tokenharbor.ai/v1/chat/completions",
			models:     []string{"claude-opus-5.5", "claude-sonnet-5", "gpt-6-astra", "gpt-6-sol", "grok-4.7", "mimo-v2.6-flash:free", "mimo-v2.5:free", "qwen3.8-flash:free", "deepseek-v4.1-flash:free", "deepseek-v4-flash:free"},
			aliases:    []string{"th", "thh"},
			modelsList: "https://tokenharbor.ai/v1/models", wantFetcher: true,
		},
		{
			id: "dahl", baseURL: "https://inference.dahl.global/v1/chat/completions",
			models:     []string{"zai-org/GLM-5.3-Flash", "deepseek-ai/DeepSeek-V4-Flash-0731", "MiniMaxAI/MiniMax-M2.7"},
			aliases:    []string{"dahl-inference"},
			modelsList: "https://inference.dahl.global/v1/models", wantFetcher: true,
		},
		{
			id: "atria", baseURL: "https://api.atria-asi.ai/v1/chat/completions",
			models:     []string{"Atria-Dawn-Preview"},
			aliases:    []string{"atria-asi"},
			modelsList: "https://api.atria-asi.ai/v1/models",
		},
		{
			// agnes seeds a curated catalogue upstream (v0.5.95) because its live
			// /v1/models answers 401 without a token; bai still ships none.
			id: "agnes", baseURL: "https://apihub.agnes-ai.com/v1/chat/completions",
			aliases:    []string{"agnes-ai"},
			models:     []string{"agnes-2.5-flash", "agnes-2.5-pro", "agnes-2.5-pro-beta", "agnes-3.0-flash", "agnes-3.0-pro"},
			modelsList: "https://apihub.agnes-ai.com/v1/models",
		},
		{
			id: "bai", baseURL: "https://api.b.ai/v1/chat/completions",
			aliases:    []string{"b-ai"},
			modelsList: "https://api.b.ai/v1/models", wantFetcher: true,
		},
		{
			// Muse (Meta Model API): every model pins the Responses lane, so the
			// registry also carries a live catalogue endpoint for the dashboard.
			id: "muse", baseURL: "https://api.meta.ai/v1/chat/completions",
			models:     []string{"muse-spark-1.3", "muse-spark-1.2", "muse-spark-1.1", "muse-spark-1.3-contributor", "muse-spark-1.2-contributor"},
			aliases:    []string{"muse-ai", "meta-model-api", "muse-code", "muse-subscription"},
			modelsList: "https://api.meta.ai/v1/models", wantFetcher: true,
		},
		{
			// v1m answers System One only: no chat base URL, one endpoint.
			id: "v1m", baseURL: "",
			models:  []string{"rev-latest", "v1m-decision-engine"},
			aliases: []string{"systemone"},
		},
		{
			// SumoPod — Indonesian OpenAI-compatible gateway. Live /v1/models
			// answers 401 without a token (like agnes), so the catalogue is
			// seeded curated from the vendor's published model list.
			id: "sumopod", baseURL: "https://ai.sumopod.com/v1/chat/completions",
			models:  []string{"claude-haiku-4-5", "claude-sonnet-4-6", "claude-opus-4-6", "deepseek-v3-2", "deepseek-v4-flash", "deepseek-v4-pro", "glm-4-7", "glm-5", "glm-5-turbo", "glm-5.1", "kimi-k2-5-260127", "kimi-k2.6", "seed-2-0-code", "seed-2-0-lite", "seed-2-0-mini", "seed-2-0-pro", "qwen3.6-flash", "qwen3.6-plus", "mimo-v2-flash", "mimo-v2-omni", "mimo-v2-pro", "mimo-v2.5", "mimo-v2.5-pro", "MiniMax-M2.7-highspeed", "nvidia/nemotron-3-nano-30b", "openai/gpt-oss-20b", "qwen/qwen3-30b-a3b-instruct-2507", "zai/glm-4.7-fp8", "gemini/gemini-2.0-flash", "gemini/gemini-2.0-flash-lite", "gemini/gemini-2.5-flash", "gemini/gemini-2.5-flash-lite", "gemini/gemini-2.5-pro", "gemini/gemini-3-flash-preview", "gemini/gemini-3-pro-preview", "gemini/gemini-3.1-flash-lite-preview", "gemini/gemini-3.1-pro-preview", "gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano", "gpt-4o", "gpt-4o-mini", "gpt-5", "gpt-5-mini", "gpt-5-nano", "gpt-5.1", "gpt-5.1-codex", "gpt-5.1-codex-mini", "gpt-5.2", "gpt-5.2-codex"},
			aliases: []string{"sp"},
		},
		{
			id: "tinyfish", baseURL: "https://api.tinyfish.ai",
			authHeader: "x-api-key", authScheme: "raw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			cfg, ok := KnownProviders[tt.id]
			if !ok {
				t.Fatalf("KnownProviders has no entry for %q", tt.id)
			}
			if cfg.BaseURL != tt.baseURL {
				t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, tt.baseURL)
			}
			wantAuthHeader, wantAuthScheme := tt.authHeader, tt.authScheme
			if wantAuthHeader == "" {
				wantAuthHeader, wantAuthScheme = "Authorization", "bearer"
			}
			if cfg.AuthHeader != wantAuthHeader || cfg.AuthScheme != wantAuthScheme {
				t.Errorf("auth = %q/%q, want %q/%q", cfg.AuthHeader, cfg.AuthScheme, wantAuthHeader, wantAuthScheme)
			}

			if got := len(GetProviderModels(tt.id)); got != len(tt.models) {
				t.Errorf("catalog has %d models, want %d", got, len(tt.models))
			}
			for _, m := range tt.models {
				if !slicesContains(GetProviderModels(tt.id), m) {
					t.Errorf("catalog missing %q", m)
				}
			}

			for _, alias := range tt.aliases {
				if got := ResolveAlias(alias); got != tt.id {
					t.Errorf("ResolveAlias(%q) = %q, want %q", alias, got, tt.id)
				}
			}
			// uiAlias equals the id upstream, so an id is its own prefix.
			if got := ResolveAlias(tt.id); got != tt.id {
				t.Errorf("ResolveAlias(%q) = %q, want %q", tt.id, got, tt.id)
			}

			if got := ModelsListURL(tt.id); got != tt.modelsList {
				t.Errorf("ModelsListURL = %q, want %q", got, tt.modelsList)
			}
			if tt.wantFetcher && tt.modelsList == "" {
				t.Error("provider declares modelsFetcher upstream but has no models-list URL")
			}
		})
	}
}

func slicesContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
