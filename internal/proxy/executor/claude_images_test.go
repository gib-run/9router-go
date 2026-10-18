package executor

import (
	"encoding/json"
	"strings"
	"testing"
)

const pngB64 = "iVBORw0KGgo="

func claudeMessagesOf(t *testing.T, in string) []any {
	t.Helper()
	out := EnsureClaudeMessagesFor([]byte(in), "claude", "claude-sonnet-5-5")
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	msgs, _ := m["messages"].([]any)
	return msgs
}

func assertNoImageURL(t *testing.T, v any) {
	t.Helper()
	b, _ := json.Marshal(v)
	if strings.Contains(string(b), "image_url") {
		t.Fatalf("OpenAI image_url leaked into the Claude body: %s", b)
	}
}

func contentBlocks(t *testing.T, msg any) []any {
	t.Helper()
	blocks, ok := msg.(map[string]any)["content"].([]any)
	if !ok {
		t.Fatalf("content is not an array: %#v", msg)
	}
	return blocks
}

func assertBase64Image(t *testing.T, block any) {
	t.Helper()
	b := block.(map[string]any)
	src, _ := b["source"].(map[string]any)
	if b["type"] != "image" || src["type"] != "base64" || src["media_type"] != "image/png" || src["data"] != pngB64 {
		t.Fatalf("not a Claude base64 image block: %#v", block)
	}
}

func TestEnsureClaudeMessages_UserImageURLBecomesImageBlock(t *testing.T) {
	msgs := claudeMessagesOf(t, `{"model":"claude-sonnet-5-5","messages":[
		{"role":"user","content":[{"type":"text","text":"what colour?"},
			{"type":"image_url","image_url":{"url":"data:image/png;base64,`+pngB64+`"}}]}]}`)
	assertNoImageURL(t, msgs)
	blocks := contentBlocks(t, msgs[0])
	if blocks[0].(map[string]any)["type"] != "text" {
		t.Fatalf("text part must survive: %#v", blocks[0])
	}
	assertBase64Image(t, blocks[1])
}

func TestEnsureClaudeMessages_RemoteImageURLBecomesURLSource(t *testing.T) {
	msgs := claudeMessagesOf(t, `{"model":"claude-sonnet-5-5","messages":[
		{"role":"user","content":[{"type":"image_url","image_url":{"url":"https://example.com/a.png"}}]}]}`)
	assertNoImageURL(t, msgs)
	src := contentBlocks(t, msgs[0])[0].(map[string]any)["source"].(map[string]any)
	if src["type"] != "url" || src["url"] != "https://example.com/a.png" {
		t.Fatalf("remote image must become a url source: %#v", src)
	}
}

func TestEnsureClaudeMessages_BareStringImageURL(t *testing.T) {
	msgs := claudeMessagesOf(t, `{"model":"claude-sonnet-5-5","messages":[
		{"role":"user","content":[{"type":"image_url","image_url":"data:image/png;base64,`+pngB64+`"}]}]}`)
	assertNoImageURL(t, msgs)
	assertBase64Image(t, contentBlocks(t, msgs[0])[0])
}

// The shape the pi coding agent produces for a tool that returns an image: a
// text-only tool result followed by a user turn carrying the image. The user
// turn is merged into the tool_result turn, so the image is its 3rd block.
func TestEnsureClaudeMessages_ToolResultImageTurn(t *testing.T) {
	msgs := claudeMessagesOf(t, `{"model":"claude-sonnet-5-5","messages":[
		{"role":"user","content":"screenshot it"},
		{"role":"assistant","content":null,"tool_calls":[{"id":"c1","type":"function","function":{"name":"shot","arguments":"{}"}}]},
		{"role":"tool","tool_call_id":"c1","content":"(see attached image)"},
		{"role":"user","content":[{"type":"text","text":"Attached image(s) from tool result:"},
			{"type":"image_url","image_url":{"url":"data:image/png;base64,`+pngB64+`"}}]}]}`)
	assertNoImageURL(t, msgs)
	blocks := contentBlocks(t, msgs[2])
	if blocks[0].(map[string]any)["type"] != "tool_result" {
		t.Fatalf("tool_result must stay first: %#v", blocks[0])
	}
	assertBase64Image(t, blocks[2])
}

func TestEnsureClaudeMessages_ToolResultContentImage(t *testing.T) {
	msgs := claudeMessagesOf(t, `{"model":"claude-sonnet-5-5","messages":[
		{"role":"user","content":"go"},
		{"role":"assistant","content":null,"tool_calls":[{"id":"c1","type":"function","function":{"name":"shot","arguments":"{}"}}]},
		{"role":"tool","tool_call_id":"c1","content":[{"type":"text","text":"shot"},
			{"type":"image_url","image_url":{"url":"data:image/png;base64,`+pngB64+`"}}]}]}`)
	assertNoImageURL(t, msgs)
	tr := contentBlocks(t, msgs[2])[0].(map[string]any)
	if tr["type"] != "tool_result" || tr["tool_use_id"] != "c1" {
		t.Fatalf("tool_result lost: %#v", tr)
	}
	inner := tr["content"].([]any)
	if inner[0].(map[string]any)["type"] != "text" {
		t.Fatalf("tool_result text lost: %#v", inner)
	}
	assertBase64Image(t, inner[1])
}

func TestEnsureClaudeMessages_AssistantToolCallsKeepsImages(t *testing.T) {
	msgs := claudeMessagesOf(t, `{"model":"claude-sonnet-5-5","messages":[
		{"role":"user","content":"go"},
		{"role":"assistant","content":[{"type":"image_url","image_url":{"url":"data:image/png;base64,`+pngB64+`"}}],
			"tool_calls":[{"id":"c1","type":"function","function":{"name":"x","arguments":"{}"}}]},
		{"role":"tool","tool_call_id":"c1","content":"ok"}]}`)
	assertNoImageURL(t, msgs)
	assertBase64Image(t, contentBlocks(t, msgs[1])[0])
}

// A part Claude cannot represent is dropped, not forwarded: forwarding it would
// reproduce the 400 on every later turn of the session.
func TestEnsureClaudeMessages_UnrepresentableImageDropped(t *testing.T) {
	msgs := claudeMessagesOf(t, `{"model":"claude-sonnet-5-5","messages":[
		{"role":"user","content":[{"type":"text","text":"hi"},
			{"type":"image_url","image_url":{"url":"file:///etc/passwd"}},
			{"type":"image_url","image_url":{"url":"data:image/png,notbase64"}}]}]}`)
	assertNoImageURL(t, msgs)
	blocks := contentBlocks(t, msgs[0])
	if len(blocks) != 1 || blocks[0].(map[string]any)["type"] != "text" {
		t.Fatalf("only the text part must remain: %#v", blocks)
	}
}

func TestConvertClaudeImageParts_NoImageIsUntouched(t *testing.T) {
	in := []any{map[string]any{"type": "text", "text": "hi"}}
	out := convertClaudeImageParts(in).([]any)
	if len(out) != 1 || &out[0] != &in[0] {
		t.Fatalf("a content array without images must be returned as-is")
	}
	if convertClaudeImageParts("plain") != "plain" {
		t.Fatalf("string content must be returned as-is")
	}
}

func TestConvertClaudeImageParts_DoesNotMutateInput(t *testing.T) {
	in := []any{map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/png;base64," + pngB64}}}
	_ = convertClaudeImageParts(in)
	if in[0].(map[string]any)["type"] != "image_url" {
		t.Fatalf("input was mutated: %#v", in)
	}
}
