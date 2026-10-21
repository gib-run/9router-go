package dashboard

import (
	"fmt"
	"strings"

	"9router/proxy/internal/db"
)

// Model access in the shape the Node dashboard writes (upstream
// src/shared/utils/keyAccess.js). Go already stores the allowlist as one row
// per pattern; this file maps that row set to and from {restricted, allow} so
// a client that speaks either shape reaches the same policy.
//
// There is deliberately no second store: every write goes through
// SetAllowedModels and every read through GetAllowedModels, so the chat gate
// keeps seeing one source of truth. A separate `restricted` column would be a
// second place for the policy to disagree with itself.

// keyAccessInput is the request-side access object. `Restricted` is a pointer
// so a body carrying only `allow` is refused rather than read as "unrestricted"
// — an omitted restricted flag must never widen a key. `Allow` is []any so a
// non-string entry is a 400 rather than silently dropped.
type keyAccessInput struct {
	Restricted *bool `json:"restricted"`
	Allow      []any `json:"allow"`
}

// keyAccess is the response-side access object, matching upstream's shape.
type keyAccess struct {
	Restricted bool     `json:"restricted"`
	Allow      []string `json:"allow"`
}

// accessFromPatterns maps a stored allowlist to the access object.
//
// An empty list is unrestricted. The deny-all sentinel is a restricted key
// that allows nothing — the one state that must not read as unrestricted,
// which is why it is checked before the empty list is.
func accessFromPatterns(patterns []string) keyAccess {
	if len(patterns) == 0 {
		return keyAccess{Restricted: false, Allow: []string{}}
	}
	if len(patterns) == 1 && patterns[0] == db.SentinelDenyAll {
		return keyAccess{Restricted: true, Allow: []string{}}
	}
	return keyAccess{Restricted: true, Allow: patterns}
}

// validateKeyAccess checks the untrusted object and returns the patterns to
// store. It rejects rather than repairs: a wrong type, an oversized list, or a
// missing restricted flag is a 400, not a silent guess.
func validateKeyAccess(in *keyAccessInput) ([]string, error) {
	if in == nil {
		return nil, fmt.Errorf("access must be an object { restricted, allow }")
	}
	if in.Restricted == nil {
		return nil, fmt.Errorf("access.restricted must be a boolean")
	}
	allow, err := decodeAllowList(in.Allow)
	if err != nil {
		return nil, err
	}
	cleaned := dedupePatterns(cleanModelPatterns(allow))
	if _, err := validateModelAllowlist(cleaned); err != nil {
		return nil, err
	}
	if !*in.Restricted {
		// An unrestricted key carries no patterns. Clearing the list is what
		// makes `{"restricted": false, "allow": ["a"]}` mean "open" rather
		// than leaving a stale restriction behind.
		return nil, nil
	}
	if len(cleaned) == 0 {
		// Restricted with nothing allowed: the sentinel is the stored form, so
		// the gate denies every model instead of reading an empty list as
		// "no allowlist, allow all".
		return []string{db.SentinelDenyAll}, nil
	}
	return cleaned, nil
}

// decodeAllowList keeps the entries as `any` so a non-string element is
// reported as the type error it is, instead of failing the whole body with
// "invalid JSON" — the client could not tell which field was wrong.
func decodeAllowList(raw []any) ([]string, error) {
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("access.allow must be an array of strings")
		}
		out = append(out, s)
	}
	return out, nil
}

// dedupePatterns drops case-insensitive repeats, keeping the first spelling —
// upstream's normalizeAllowList order rule. The exact same list written twice
// must not store two rows that the gate then matches twice.
func dedupePatterns(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		key := strings.ToLower(v)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v)
	}
	return out
}
