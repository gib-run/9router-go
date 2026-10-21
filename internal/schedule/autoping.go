// Package schedule holds the anchored ping schedule and the auto-ping
// scheduler that fires it.
//
// This file is the production wiring: it turns the settings blob into the
// scheduler's ConfigSource and the scheduler's PingFunc into a real request
// against one provider connection.
//
// Upstream (src/shared/services/quotaAutoPing.js) sends a tiny chat request
// to the account whose window is about to reset — "hi", one token — so the
// 5h window starts on time instead of at the first real request. Go reaches
// the same upstream through the chat lane, so the ping is a normal chat
// request pinned to one connection.
package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"9router/proxy/internal/handlers/chat"
	"9router/proxy/internal/log"
)

// Ping models and text, ported from upstream QUOTA_AUTOPING_CONFIG
// (src/shared/constants/config.js:66-93).
//
// claude: "claude-haiku-4-5-20251001", one token.
// codex:  "gpt-5.6-terra" — the lightweight model the upstream comment names
//         for the Codex 5h reset window.
const (
	ClaudePingModel = "claude-haiku-4-5-20251001"
	CodexPingModel  = "gpt-5.6-terra"

	// PingText is the prompt every ping sends. Upstream sends "hi" for both
	// providers; the Codex instructions line ("Reply with OK.") rides along
	// as the Responses `instructions` field.
	PingText = "hi"

	// PingMaxTokens caps the answer. Upstream sends max_tokens: 1 for Claude;
	// the Codex backend rejects a max_output_tokens it did not ask for, so
	// the codex executor strips it (internal/proxy/executor/providers.go).
	PingMaxTokens = 1
)

// PingModel returns the lightweight model one provider's ping uses.
func PingModel(provider string) string {
	switch provider {
	case "codex":
		return CodexPingModel
	default:
		return ClaudePingModel
	}
}

// PingTimeout bounds one ping. A provider that never answers must not hold
// the scheduler's tick open — the tick is shared by every connection, so one
// hung account would delay every other account's ping.
const PingTimeout = 30 * time.Second

// PingConnection sends one ping to a provider connection.
//
// The request travels the ordinary chat lane with the connection pinned, so
// it uses the same credential resolution, OAuth refresh, proxy-pool binding
// and executor the real traffic uses. It is marked as a probe so the usage
// ledger does not record it as a turn an operator sent — upstream makes the
// same distinction (a ping is not operator traffic).
//
// ponytail: the ping asks for one token and the provider's own lightest
// model, exactly as upstream does. It does NOT reproduce upstream's
// post-ping bookkeeping (lastPingedSlotKey / lastPingAt persisted on the
// connection row) — the scheduler's own slot dedupe covers the "fire once
// per slot" contract, so a ping that failed upstream is not retried for the
// rest of the grace window. Upstream instead persists the slot key on the
// connection and re-reads it, which survives a restart. Add the persisted
// slot key when a restart-during-grace-window ping must not repeat.
func PingConnection(ctx context.Context, chatH *chat.ChatHandler, provider, connectionID string) error {
	if chatH == nil {
		return fmt.Errorf("no chat handler")
	}
	if connectionID == "" {
		return fmt.Errorf("no connection id")
	}

	ctx, cancel := context.WithTimeout(ctx, PingTimeout)
	defer cancel()

	model := PingModel(provider)
	body, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": PingMaxTokens,
		"stream":     false,
		"messages": []map[string]any{
			{"role": "user", "content": PingText},
		},
	})
	if err != nil {
		return fmt.Errorf("build ping body: %w", err)
	}

	return chatH.PingConnection(ctx, provider, model, connectionID, body)
}

// SettingsReader reads the settings blob as the dashboard stores it.
//
// The raw blob is read rather than the typed SettingsData because the
// auto-ping halves are not typed fields: they are two nested maps under one
// key, and a typed struct would have to model a shape the rest of the
// settings do not use.
type SettingsReader func() (map[string]any, error)

// ConfigSourceFromSettings reads one provider's auto-ping configuration out of
// the settings blob.
//
// The settings are read on EVERY tick rather than cached, so a schedule saved
// in the dashboard takes effect on the next tick without a restart — the same
// property upstream has (quotaAutoPing.js:290 reads settings each tick).
//
// A missing or malformed settings blob yields an empty configuration, which
// makes the scheduler do nothing. That is the right failure direction: an
// unreadable settings row must not make the scheduler ping accounts whose
// schedule the operator removed.
func ConfigSourceFromSettings(read SettingsReader) ConfigSource {
	return func(provider string) Config {
		empty := Config{Connections: map[string]bool{}, Schedules: map[string]*Schedule{}}
		if read == nil {
			return empty
		}
		raw, err := read()
		if err != nil || raw == nil {
			return empty
		}
		return ParseConfig(raw[SettingsKey(provider)])
	}
}

// PingFuncFromChat builds the scheduler's PingFunc from the chat handler.
func PingFuncFromChat(chatH *chat.ChatHandler) PingFunc {
	return func(ctx context.Context, provider, connectionID string) error {
		err := PingConnection(ctx, chatH, provider, connectionID)
		if err != nil {
			LogPingFailure(provider, connectionID, err)
		}
		return err
	}
}

// LogPingFailure reports a failed ping. The scheduler returns the error to
// its caller; this is the line an operator sees in the log.
func LogPingFailure(provider, connectionID string, err error) {
	log.Warn("autoping", "ping failed", "provider", provider, "conn", connectionID, "error", err)
}
