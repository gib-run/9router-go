package chat

import (
	"context"

	"9router/proxy/internal/providers"
	"9router/proxy/internal/translator"
)

// inlineImageTargets are the providers whose wire format cannot fetch a remote
// URL itself, so an inbound `http(s)` image reference would reach the model as
// a dead string. Ported from open-sse/translator/concerns/prefetch.js
// (TARGETS_NEED_BASE64); anything not listed here keeps remote URLs as-is.
var inlineImageTargets = map[string]bool{
	"gemini":      true,
	"gemini-cli":  true,
	"vertex":      true,
	"antigravity": true,
	"ollama":      true,
	"kiro":        true,
	"commandcode": true,
}

// imagePrefetchOpts is the policy for every dispatch. A package var so a
// dispatch test can point it at a local image server; production leaves the
// zero value, which means the https-only, SSRF-guarded defaults.
var imagePrefetchOpts = translator.PrefetchOptions{}

// wantsInlineImages reports whether this dispatch target needs the prefetch.
func wantsInlineImages(provider string, cfg *providers.ProviderConfig) bool {
	if inlineImageTargets[provider] {
		return true
	}
	return cfg != nil && cfg.IsGeminiNative()
}

// prefetchImagesForTarget rewrites remote image URLs into inline base64 when
// the target cannot fetch them, and returns the body untouched otherwise.
//
// G11's discovery doc places this before translation, which is here: the body
// at this point is still in the client's own format and the provider that will
// receive it is known. Fail-open — every refusal inside returns the body.
func prefetchImagesForTarget(ctx context.Context, provider string, cfg *providers.ProviderConfig, body []byte) []byte {
	if !wantsInlineImages(provider, cfg) {
		return body
	}
	return translator.PrefetchInlineImages(ctx, body, imagePrefetchOpts)
}
