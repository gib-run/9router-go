package proxy_test

import (
	"testing"

	"9router/proxy/internal/proxy"
)

func TestBuildOpenCodeHeaders(t *testing.T) {
	headers := proxy.BuildOpenCodeHeaders(nil, "my-session-123", true)
	if headers["User-Agent"] != proxy.DefaultOpenCodeUA {
		t.Errorf("expected User-Agent %s, got %s", proxy.DefaultOpenCodeUA, headers["User-Agent"])
	}
	if headers["x-opencode-client"] != "cli" {
		t.Errorf("expected x-opencode-client cli, got %s", headers["x-opencode-client"])
	}
	if headers["x-opencode-project"] != "global" {
		t.Errorf("expected x-opencode-project global, got %s", headers["x-opencode-project"])
	}
	session := headers["x-opencode-session"]
	if !proxy.OpenCodeSessionRegex.MatchString(session) {
		t.Errorf("expected canonical x-opencode-session matching regex, got %s", session)
	}
	if len(session) != 30 {
		t.Errorf("expected session length 30, got %d (%s)", len(session), session)
	}
	for _, h := range []string{"x-opencode-session-id", "x-session-affinity", "x-session-id"} {
		if headers[h] != session {
			t.Errorf("expected %s to equal session %s, got %s", h, session, headers[h])
		}
	}
	if headers["Accept"] != "text/event-stream" {
		t.Errorf("expected Accept text/event-stream, got %s", headers["Accept"])
	}

	// Non-stream accepts */*
	headersNonStream := proxy.BuildOpenCodeHeaders(nil, "", false)
	if headersNonStream["Accept"] != "*/*" {
		t.Errorf("expected Accept */*, got %s", headersNonStream["Accept"])
	}

	// Custom non-global project override
	custom := proxy.BuildOpenCodeHeaders(map[string]string{"x-opencode-project": "proj-xyz"}, "ses-1", true)
	if custom["x-opencode-project"] != "proj-xyz" {
		t.Errorf("expected custom x-opencode-project proj-xyz, got %s", custom["x-opencode-project"])
	}
}

func TestOpenCodeCanonicalSession(t *testing.T) {
	for range 20 {
		ses := proxy.GenerateOpenCodeSessionID()
		if !proxy.OpenCodeSessionRegex.MatchString(ses) || len(ses) != 30 {
			t.Fatalf("invalid session ID generated: %s (len %d)", ses, len(ses))
		}

		req := proxy.GenerateOpenCodeRequestID()
		if !proxy.OpenCodeRequestRegex.MatchString(req) || len(req) != 30 {
			t.Fatalf("invalid request ID generated: %s (len %d)", req, len(req))
		}
	}

	// Preserves valid canonical session
	valid := "ses_f52b0d414ffeObbCKcHUQYZR2I"
	if translated := proxy.TranslateOpenCodeSessionID(valid, ""); translated != valid {
		t.Errorf("expected %s preserved, got %s", valid, translated)
	}

	// Translates foreign session to canonical 30 chars
	translated := proxy.TranslateOpenCodeSessionID("uuid-1234-5678-abcdef", "claude")
	if !proxy.OpenCodeSessionRegex.MatchString(translated) || len(translated) != 30 {
		t.Errorf("expected valid canonical session from foreign ID, got %s", translated)
	}
}

func TestHasValidOpenCodeVersion(t *testing.T) {
	cases := []struct {
		ua    string
		valid bool
	}{
		{"opencode", false},
		{"opencode/1.16.5", false},
		{"opencode/1.17.0", true},
		{"opencode/1.18.31", true},
		{"opencode/2.0.0", true},
		{"opencode/latest/2.0.29/cli", true},
		{"opencode/latest/test", true},
		{"Claude-Code/1.0", false},
		{"curl/8.7.1", false},
	}
	for _, c := range cases {
		if got := proxy.HasValidOpenCodeVersion(c.ua); got != c.valid {
			t.Errorf("HasValidOpenCodeVersion(%q) = %v, want %v", c.ua, got, c.valid)
		}
	}
}
