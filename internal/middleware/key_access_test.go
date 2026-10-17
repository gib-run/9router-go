package middleware

import (
	"testing"
)

func TestIsModelAllowedForKey(t *testing.T) {
	tests := []struct {
		name           string
		allowedModels  []string
		requestedModel string
		want           bool
	}{
		{
			name:           "empty allowed list denies everything",
			allowedModels:  []string{},
			requestedModel: "openai/gpt-4o",
			want:           false,
		},
		{
			name:           "empty string patterns deny",
			allowedModels:  []string{"", "  "},
			requestedModel: "openai/gpt-4o",
			want:           false,
		},
		{
			name:           "empty requested model returns false",
			allowedModels:  []string{"*"},
			requestedModel: "",
			want:           false,
		},
		{
			name:           "exact match success",
			allowedModels:  []string{"openai/gpt-4o"},
			requestedModel: "openai/gpt-4o",
			want:           true,
		},
		{
			name:           "exact match case insensitive",
			allowedModels:  []string{"OpenAI/GPT-4O"},
			requestedModel: "openai/gpt-4o",
			want:           true,
		},
		{
			name:           "exact match different model fails",
			allowedModels:  []string{"openai/gpt-4o"},
			requestedModel: "openai/gpt-3.5-turbo",
			want:           false,
		},
		{
			name:           "wildcard prefix oc/* matches",
			allowedModels:  []string{"oc/*"},
			requestedModel: "oc/space-bunny-free",
			want:           true,
		},
		{
			name:           "wildcard prefix oc/* rejects non-match",
			allowedModels:  []string{"oc/*"},
			requestedModel: "openai/gpt-4o",
			want:           false,
		},
		{
			name:           "wildcard infix *claude* matches various positions",
			allowedModels:  []string{"*claude*"},
			requestedModel: "anthropic/claude-3-opus",
			want:           true,
		},
		{
			name:           "wildcard infix *claude* matches suffix position",
			allowedModels:  []string{"*claude*"},
			requestedModel: "my-custom-claude",
			want:           true,
		},
		{
			name:           "wildcard prefix google/* matches",
			allowedModels:  []string{"google/*"},
			requestedModel: "google/gemini-1.5-pro",
			want:           true,
		},
		{
			name:           "only negations: open except denied - allowed model",
			allowedModels:  []string{"!opus", "-haiku"},
			requestedModel: "anthropic/claude-3-sonnet",
			want:           true,
		},
		{
			name:           "only negations: denied with exclamation",
			allowedModels:  []string{"!opus", "-haiku"},
			requestedModel: "anthropic/claude-3-opus",
			want:           false,
		},
		{
			name:           "only negations: denied with dash",
			allowedModels:  []string{"!opus", "-haiku"},
			requestedModel: "claude-3-haiku",
			want:           false,
		},
		{
			name:           "mixed allowed and denied: permitted model",
			allowedModels:  []string{"*claude*", "!opus"},
			requestedModel: "anthropic/claude-3-sonnet",
			want:           true,
		},
		{
			name:           "mixed allowed and denied: matches allowed but blocked by denied",
			allowedModels:  []string{"*claude*", "!opus"},
			requestedModel: "anthropic/claude-3-opus",
			want:           false,
		},
		{
			name:           "mixed allowed and denied: does not match allowed",
			allowedModels:  []string{"*claude*", "!opus"},
			requestedModel: "openai/gpt-4o",
			want:           false,
		},
		{
			name:           "wildcard star allows everything",
			allowedModels:  []string{"*"},
			requestedModel: "any/arbitrary-model",
			want:           true,
		},
		{
			name:           "wildcard star with negation",
			allowedModels:  []string{"*", "!deepseek*"},
			requestedModel: "deepseek/deepseek-chat",
			want:           false,
		},
		{
			name:           "wildcard star with negation allowed",
			allowedModels:  []string{"*", "!deepseek*"},
			requestedModel: "openai/gpt-4o",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsModelAllowedForKey(tt.allowedModels, tt.requestedModel)
			if got != tt.want {
				t.Errorf("IsModelAllowedForKey(%v, %q) = %v, want %v", tt.allowedModels, tt.requestedModel, got, tt.want)
			}
		})
	}
}
