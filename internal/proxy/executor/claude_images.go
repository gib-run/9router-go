package executor

import "strings"

// claudeImageBlock converts one OpenAI `image_url` content part into the block
// the Claude Messages API defines for an image:
//
//	{"type":"image_url","image_url":{"url":"data:image/png;base64,..."}}
//	  -> {"type":"image","source":{"type":"base64","media_type":"image/png","data":"..."}}
//	{"type":"image_url","image_url":{"url":"https://..."}}
//	  -> {"type":"image","source":{"type":"url","url":"https://..."}}
//
// `image_url` may be the object form or a bare string (some clients send it).
// ok is false when the URL is neither a base64 data URI nor http(s), which
// Claude has no way to represent. Mirrors upstream
// open-sse/translator/request/openai-to-claude.js (IMAGE_URL branch).
func claudeImageBlock(part map[string]any) (block map[string]any, ok bool) {
	var url string
	switch u := part["image_url"].(type) {
	case string:
		url = u
	case map[string]any:
		url, _ = u["url"].(string)
	}
	url = strings.TrimSpace(url)

	if rest, found := strings.CutPrefix(url, "data:"); found {
		meta, data, found := strings.Cut(rest, ",")
		mediaType, isBase64 := strings.CutSuffix(meta, ";base64")
		if !found || !isBase64 || mediaType == "" || data == "" {
			return nil, false
		}
		return map[string]any{
			"type": "image",
			"source": map[string]any{
				"type":       "base64",
				"media_type": mediaType,
				"data":       data,
			},
		}, true
	}
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return map[string]any{
			"type":   "image",
			"source": map[string]any{"type": "url", "url": url},
		}, true
	}
	return nil, false
}

// convertClaudeImageParts rewrites every OpenAI `image_url` part of a content
// array into a Claude `image` block, and returns the content unchanged when it
// is not an array or holds none. It never mutates its input.
//
// Anthropic rejects the OpenAI shape with
// "messages.<n>.content.<i>: Input tag 'image_url' found using 'type' does not
// match any of the expected tags", and because the message stays in the
// client's history, every later request of the conversation failed too. A part
// Claude cannot represent (see claudeImageBlock) is dropped rather than
// forwarded: forwarding it would only reproduce that 400 for the whole session.
// Used for user turns, replayed assistant turns, and tool_result content, which
// all accept image blocks.
func convertClaudeImageParts(content any) any {
	parts, ok := content.([]any)
	if !ok {
		return content
	}
	changed := false
	for _, p := range parts {
		if m, isMap := p.(map[string]any); isMap && m["type"] == "image_url" {
			changed = true
			break
		}
	}
	if !changed {
		return content
	}
	out := make([]any, 0, len(parts))
	for _, p := range parts {
		m, isMap := p.(map[string]any)
		if !isMap || m["type"] != "image_url" {
			out = append(out, p)
			continue
		}
		if block, ok := claudeImageBlock(m); ok {
			out = append(out, block)
		}
	}
	return out
}
