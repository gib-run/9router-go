package chat

import (
	"strings"
	"testing"

	"9router/proxy/internal/db"
	"9router/proxy/internal/providers"
)

// TestHandleModels_HiddenProviderIsNotAdvertised pins the routing half of the
// Hidden flag on the one path that would otherwise publish it. A no-auth
// registry provider with no connection row is added to /v1/models so a caller
// can use the free tier without connecting anything; that rule is right for
// providers a person picks between and wrong for one the platform runs on its
// own. Hidden removes it from the catalog without unregistering it: a stored
// connection pointing at mimo-free must still resolve and route.
func TestHandleModels_HiddenProviderIsNotAdvertised(t *testing.T) {
	cfg, ok := providers.KnownProviders["mimo-free"]
	if !ok || !cfg.Hidden || !cfg.NoAuth {
		t.Skip("fixture assumes mimo-free is a hidden no-auth provider")
	}

	database, cleanup := setupChatTestDB(t)
	defer cleanup()
	h := NewChatHandler(db.NewRepo(database))

	// Connected mode is where the no-auth catalog is published at all: the
	// default listing only dumps it when there are no connections, and the
	// fixture seeds one. Asking for the mode that reaches the code under test
	// is what makes the assertions below mean anything.
	ids := modelsIDsQuery(t, h, "?connected=1")

	// hiddenPrefixes is decided from the registry, not from the two names this
	// fixture happens to know: whatever is marked Hidden must be absent from the
	// listing, under every spelling that resolves to it.
	hiddenPrefixes := make(map[string]bool, 2)
	for id := range providers.KnownProviders {
		if providers.KnownProviders[id].Hidden {
			hiddenPrefixes[id] = true
		}
	}
	for alias := range providers.ProviderModels {
		if canon := providers.ResolveAlias(alias); hiddenPrefixes[canon] {
			hiddenPrefixes[alias] = true
		}
	}
	for _, id := range ids {
		prefix, _, ok := strings.Cut(id, "/")
		if ok && hiddenPrefixes[prefix] {
			t.Errorf("hidden provider reached /v1/models: %q", id)
		}
	}
	// The control. The absence above is only meaningful if the no-auth catalog
	// pass ran and published its visible members — otherwise a listing that
	// never reached appendNoAuthCatalogModels would pass the assertion for the
	// wrong reason. The fixture seeds connections, so the visible no-auth
	// providers in the result must come from that pass and nowhere else.
	visiblePresent := false
	for _, id := range ids {
		prefix, _, ok := strings.Cut(id, "/")
		if !ok || hiddenPrefixes[prefix] || prefix == "ds" || prefix == "groq" {
			continue
		}
		if providers.IsNoAuthProvider(prefix) {
			visiblePresent = true
			break
		}
	}
	if !visiblePresent {
		t.Fatalf("no visible no-auth provider in /v1/models (%d ids); the no-auth catalog pass did not run, so the absence of %v proves nothing", len(ids), hiddenPrefixes)
	}

	// Hidden is not disabled: the resolver still knows the provider, which is
	// what routing uses for a connection that already points at it.
	if got := providers.ResolveAlias("mmf"); got != "mimo-free" {
		t.Errorf("ResolveAlias(mmf) = %q, want mimo-free: hidden must not unregister the provider", got)
	}
	if providers.GetProviderAlias("mimo-free") == "" {
		t.Error("mimo-free lost its published alias; hidden is about listing, not identity")
	}
}
