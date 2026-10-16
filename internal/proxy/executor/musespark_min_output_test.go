package executor

import (
	"encoding/json/v2"
	"testing"
)

func TestMuseSparkMinOutputTokensAndSuffix(t *testing.T) {
	t.Run("max_output_tokens < 512 is clamped to 512", func(t *testing.T) {
		body := []byte(`{"model":"muse-spark-1.3","max_output_tokens":100}`)
		out, err := normalizeMuseSparkResponsesBody(body, "muse-spark-1.3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(out, &m); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if m["max_output_tokens"] != float64(512) {
			t.Errorf("expected max_output_tokens 512, got %v", m["max_output_tokens"])
		}
	})

	t.Run("max_tokens < 512 is clamped to 512", func(t *testing.T) {
		body := []byte(`{"model":"muse-spark-1.3","max_tokens":200}`)
		out, err := normalizeMuseSparkResponsesBody(body, "muse-spark-1.3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(out, &m); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if m["max_output_tokens"] != float64(512) {
			t.Errorf("expected max_output_tokens 512, got %v", m["max_output_tokens"])
		}
		if _, ok := m["max_tokens"]; ok {
			t.Errorf("expected max_tokens to be deleted")
		}
	})

	t.Run("budget >= 512 is preserved", func(t *testing.T) {
		body := []byte(`{"model":"muse-spark-1.3","max_output_tokens":2048}`)
		out, err := normalizeMuseSparkResponsesBody(body, "muse-spark-1.3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(out, &m); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if m["max_output_tokens"] != float64(2048) {
			t.Errorf("expected max_output_tokens 2048, got %v", m["max_output_tokens"])
		}
	})

	t.Run("reasoning effort suffix like high is parsed to reasoning.effort", func(t *testing.T) {
		body := []byte(`{"model":"muse-spark-1.3"}`)
		out, err := normalizeMuseSparkResponsesBody(body, "muse-spark-1.3(high)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(out, &m); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		rMap, ok := m["reasoning"].(map[string]any)
		if !ok {
			t.Fatalf("expected reasoning object, got %v", m["reasoning"])
		}
		if rMap["effort"] != "high" {
			t.Errorf("expected reasoning.effort 'high', got %v", rMap["effort"])
		}
	})
}
