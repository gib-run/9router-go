package headroom

import (
	"context"
	json "encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCompress_HealthyMessageCompressionWithLargeTools proves upstream db81a8c9 (#4855) parity:
// When tools dominate the outbound body, healthy message compression (e.g. 50% shrink)
// must NOT be rejected by the phantom savings guard.
func TestCompress_HealthyMessageCompressionWithLargeTools(t *testing.T) {
	origMessage := strings.Repeat("x", 1000)
	compressedMessage := strings.Repeat("x", 500)
	largeTools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "lookup",
				"description": strings.Repeat("tool schema description ", 2000), // ~44KB tool schema
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"messages": []any{
				map[string]any{"role": "user", "content": compressedMessage},
			},
			"tokens_before": 1000,
			"tokens_after":  500,
			"tokens_saved":  500,
		}
		data, _ := json.Marshal(resp)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	bodyMap := map[string]any{
		"model": "gpt-4o",
		"messages": []any{
			map[string]any{"role": "user", "content": origMessage},
		},
		"tools": largeTools,
	}
	body, err := json.Marshal(bodyMap)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	client := NewClient(ts.URL, "", nil)
	got, compressed, err := client.Compress(context.Background(), body, "gpt-4o", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !compressed {
		t.Fatalf("expected healthy message compression to be committed even when tools dominate body")
	}

	var gotMap map[string]any
	if err := json.Unmarshal(got, &gotMap); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	msgs, ok := gotMap["messages"].([]any)
	if !ok || len(msgs) != 1 {
		t.Fatalf("expected 1 message in compressed result, got: %v", gotMap["messages"])
	}
	firstMsg := msgs[0].(map[string]any)
	if firstMsg["content"] != compressedMessage {
		t.Errorf("content = %v, want compressedMessage", firstMsg["content"])
	}
}

// TestCompress_RejectsBarelySmallerMessages verifies that when messages and body
// both shrink by less than the 5% threshold, phantom savings is rejected.
func TestCompress_RejectsBarelySmallerMessages(t *testing.T) {
	origMessage := strings.Repeat("x", 1000)
	nearlySameMessage := strings.Repeat("x", 995) // only 0.5% shrink

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"messages": []any{
				map[string]any{"role": "user", "content": nearlySameMessage},
			},
			"tokens_before": 1000,
			"tokens_after":  995,
			"tokens_saved":  500, // reporting positive token savings
		}
		data, _ := json.Marshal(resp)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	bodyMap := map[string]any{
		"model": "gpt-4o",
		"messages": []any{
			map[string]any{"role": "user", "content": origMessage},
		},
	}
	body, _ := json.Marshal(bodyMap)

	client := NewClient(ts.URL, "", nil)
	_, compressed, err := client.Compress(context.Background(), body, "gpt-4o", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if compressed {
		t.Fatalf("expected compression to be rejected when both messages and body shrink <5%%")
	}
}

// TestIsHeadroomPhantomSavings_Table mirrors upstream tests/unit/headroom.test.js:
func TestIsHeadroomPhantomSavings_Table(t *testing.T) {
	tests := []struct {
		name        string
		beforeMsg   int
		afterMsg    int
		wantPhantom bool
	}{
		{"healthy message compression with large tools", 1000, 500, false},
		{"barely smaller messages", 1000, 990, true},
		{"unchanged messages", 1000, 1000, true},
		{"larger messages", 1000, 1100, true},
		{"missing message sizes", 0, 0, false},
		{"empty original messages", 0, 10, false},
		{"missing compressed size", 1000, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &HeadroomStats{TokensSaved: 500}
			diag := &HeadroomDiagnostics{
				Before: HeadroomSizes{BodyBytes: 100000, MessageBytes: tt.beforeMsg},
				After:  HeadroomSizes{BodyBytes: 99500, MessageBytes: tt.afterMsg},
			}
			got := IsHeadroomPhantomSavings(stats, diag)
			if got != tt.wantPhantom {
				t.Errorf("IsHeadroomPhantomSavings() = %v, want %v", got, tt.wantPhantom)
			}
		})
	}
}

func TestIsHeadroomPhantomSavings_NoPositiveTokensSaved(t *testing.T) {
	diag := &HeadroomDiagnostics{
		Before: HeadroomSizes{BodyBytes: 100000, MessageBytes: 1000},
		After:  HeadroomSizes{BodyBytes: 100000, MessageBytes: 1000},
	}

	for _, stats := range []*HeadroomStats{nil, {}, {TokensSaved: 0}, {TokensSaved: -1}} {
		if got := IsHeadroomPhantomSavings(stats, diag); got {
			t.Errorf("IsHeadroomPhantomSavings(%v) = true, want false", stats)
		}
	}
}

func TestIsHeadroomPhantomSavings_MinShrinkRatioThreshold(t *testing.T) {
	diag := &HeadroomDiagnostics{
		Before: HeadroomSizes{BodyBytes: 100000, MessageBytes: 1000},
		After:  HeadroomSizes{BodyBytes: 99900, MessageBytes: 900}, // 10% shrink
	}
	stats := &HeadroomStats{TokensSaved: 100}

	// Default 5%: 10% >= 5%, so not phantom
	if got := IsHeadroomPhantomSavings(stats, diag); got {
		t.Errorf("expected false with default 0.05 threshold, got %v", got)
	}

	// Custom 20%: 10% < 20%, so phantom
	if got := IsHeadroomPhantomSavings(stats, diag, 0.20); !got {
		t.Errorf("expected true with 0.20 threshold, got %v", got)
	}
}

func TestIsHeadroomPhantomSavings_ClaudeSystemPromptSavings(t *testing.T) {
	// Top-level system prompt compression for Claude:
	// bodyBytes shrinks by 50%, while messageBytes does not shrink.
	stats := &HeadroomStats{TokensSaved: 500}
	diag := &HeadroomDiagnostics{
		Before: HeadroomSizes{BodyBytes: 100000, MessageBytes: 1000},
		After:  HeadroomSizes{BodyBytes: 50000, MessageBytes: 1000},
	}
	if got := IsHeadroomPhantomSavings(stats, diag); got {
		t.Errorf("expected Claude system prompt savings not to be flagged as phantom, got %v", got)
	}
}

func TestHeadroomPhantomWarning_Text(t *testing.T) {
	diag := &HeadroomDiagnostics{
		Before: HeadroomSizes{BodyBytes: 100000, MessageBytes: 1000},
		After:  HeadroomSizes{BodyBytes: 99500, MessageBytes: 990},
	}
	warning := FormatHeadroomPhantomWarning(diag)
	if !strings.Contains(warning, "reported token delta, but messages and outbound JSON shrank <5%") {
		t.Errorf("expected updated warning text in %q", warning)
	}
	if strings.Contains(warning, "reported token delta, but outbound JSON shrank <5%;") {
		t.Errorf("warning still contains obsolete text: %q", warning)
	}
}
