package middleware

import (
	"strings"
)

// IsModelAllowedForKey checks whether requestedModel is permitted by allowedModels.
//
// Rules:
// - If allowedModels is empty, access is denied (returns false).
// - Patterns starting with '!' or '-' are treated as negations (denied).
// - Matching is case-insensitive and supports exact match, wildcards (*), prefix, suffix, and substring/contains.
// - If the list contains only negations, access is open-except-denied.
// - If allowed patterns exist, the model must match at least one allowed pattern and no denied pattern.
func IsModelAllowedForKey(allowedModels []string, requestedModel string) bool {
	if len(allowedModels) == 0 {
		return false
	}

	modelLower := strings.ToLower(strings.TrimSpace(requestedModel))
	if modelLower == "" {
		return false
	}

	var allowed []string
	var denied []string

	for _, raw := range allowedModels {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "!") || strings.HasPrefix(p, "-") {
			stripped := strings.TrimSpace(p[1:])
			if stripped != "" {
				denied = append(denied, strings.ToLower(stripped))
			}
		} else {
			allowed = append(allowed, strings.ToLower(p))
		}
	}

	if len(allowed) == 0 && len(denied) == 0 {
		return false
	}

	// Any match against a denied pattern immediately denies access.
	for _, d := range denied {
		if matchesModelPattern(d, modelLower) {
			return false
		}
	}

	// If only denied patterns exist, access is open-except-denied.
	if len(allowed) == 0 {
		return true
	}

	// Must match at least one allowed pattern.
	for _, a := range allowed {
		if matchesModelPattern(a, modelLower) {
			return true
		}
	}

	return false
}

// matchesModelPattern checks if model matches pattern case-insensitively.
func matchesModelPattern(pattern, model string) bool {
	if pattern == "*" || pattern == model {
		return true
	}
	if strings.Contains(pattern, "*") {
		return matchWildcard(pattern, model)
	}
	return pattern == model || strings.Contains(model, pattern)
}

// matchWildcard matches a wildcard pattern where '*' matches any sequence of characters.
func matchWildcard(pattern, s string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == s
	}

	// Must start with prefix
	if !strings.HasPrefix(s, parts[0]) {
		return false
	}
	s = s[len(parts[0]):]

	// Match inner segments
	for i := 1; i < len(parts)-1; i++ {
		seg := parts[i]
		if seg == "" {
			continue
		}
		idx := strings.Index(s, seg)
		if idx == -1 {
			return false
		}
		s = s[idx+len(seg):]
	}

	// Must end with suffix
	return strings.HasSuffix(s, parts[len(parts)-1])
}
