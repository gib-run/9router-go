//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/samber/lo"
)

// accessObject is the `access` field the Node dashboard reads and writes.
type accessObject struct {
	Restricted bool     `json:"restricted"`
	Allow      []string `json:"allow"`
}

type allowlistWithAccess struct {
	ID     string       `json:"id"`
	Models []string     `json:"models"`
	Access accessObject `json:"access"`
}

type keyRowWithAccess struct {
	ID     string       `json:"id"`
	Access accessObject `json:"access"`
}

// newKeyForAccess mints a key and returns its id, reusing the harness helper
// that already knows the POST /api/keys contract.
func newKeyForAccess(t *testing.T, env *Env, name string) string {
	t.Helper()
	id, _ := createKey(t, env, name)
	return id
}

func accessOfKey(t *testing.T, env *Env, id string) accessObject {
	t.Helper()
	res := env.Get(t, "/api/keys")
	if res.Status != http.StatusOK {
		t.Fatalf("GET /api/keys = %d, want 200 (body: %s)", res.Status, truncate(res.Body))
	}
	var rows []keyRowWithAccess
	res.Decode(t, &rows)
	row, ok := lo.Find(rows, func(r keyRowWithAccess) bool { return r.ID == id })
	if !ok {
		t.Fatalf("api key %s is not in GET /api/keys", id)
	}
	return row.Access
}

func modelsOfKey(t *testing.T, env *Env, id string) allowlistWithAccess {
	t.Helper()
	res := env.Get(t, "/api/keys/"+id+"/models")
	if res.Status != http.StatusOK {
		t.Fatalf("GET /api/keys/%s/models = %d, want 200 (body: %s)", id, res.Status, truncate(res.Body))
	}
	var out allowlistWithAccess
	res.Decode(t, &out)
	return out
}

func putAccess(t *testing.T, env *Env, id string, access map[string]any) {
	t.Helper()
	res := env.Put(t, "/api/keys/"+id, map[string]any{"access": access})
	if res.Status != http.StatusOK {
		t.Fatalf("PUT access = %d, want 200 (body: %s)", res.Status, truncate(res.Body))
	}
}

func samePatterns(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for _, w := range want {
		if !lo.Contains(got, w) {
			return false
		}
	}
	return true
}

// (a) writes the `access` form and reads the flat /models list back. Both must
// describe the same restriction — the point of the alias is that a client
// speaking either shape reaches one policy, not two.
func TestKeyAccessShapeIsTheSamePolicyAsModels(t *testing.T) {
	env := newEnv(t)
	id := newKeyForAccess(t, env, "access-a")

	want := []string{"cc/claude-opus-5", "gpt-4o"}
	putAccess(t, env, id, map[string]any{"restricted": true, "allow": want})

	got := modelsOfKey(t, env, id)
	if !samePatterns(got.Models, want) {
		t.Errorf("GET /models after writing access = %v, want %v", got.Models, want)
	}
	if !got.Access.Restricted {
		t.Errorf("access.restricted = false after writing a restricted list")
	}
	if !samePatterns(got.Access.Allow, want) {
		t.Errorf("access.allow = %v, want %v", got.Access.Allow, want)
	}
}

// (b) writes the flat `models` form and reads `access` back — the reverse
// direction, which is what the Node dashboard needs from a Go-written key.
func TestKeyModelsShapeIsTheSamePolicyAsAccess(t *testing.T) {
	env := newEnv(t)
	id := newKeyForAccess(t, env, "access-b")

	want := []string{"claude/*", "oc/*"}
	res := env.Put(t, "/api/keys/"+id+"/models", map[string]any{"models": want})
	if res.Status != http.StatusOK {
		t.Fatalf("PUT /models = %d, want 200 (body: %s)", res.Status, truncate(res.Body))
	}

	got := accessOfKey(t, env, id)
	if !got.Restricted {
		t.Fatalf("GET /api/keys access.restricted = false after writing an allowlist")
	}
	if !samePatterns(got.Allow, want) {
		t.Errorf("GET /api/keys access.allow = %v, want %v", got.Allow, want)
	}
}

// (c) restricted with an empty list is deny-all, not "no restriction". The
// stored form is the deny-all sentinel so the chat gate cannot read the empty
// list as allow-everything.
func TestKeyAccessRestrictedEmptyDeniesEverything(t *testing.T) {
	env := newEnv(t)
	id := newKeyForAccess(t, env, "access-c")

	putAccess(t, env, id, map[string]any{"restricted": true, "allow": []string{}})

	got := accessOfKey(t, env, id)
	if !got.Restricted {
		t.Fatalf("access.restricted = false for a restricted key with an empty allow list")
	}
	if len(got.Allow) != 0 {
		t.Errorf("access.allow = %v, want empty", got.Allow)
	}

	// The stored form must be the deny-all sentinel, not an empty table: an
	// empty table reads as unrestricted to the chat gate.
	models := modelsOfKey(t, env, id)
	if !samePatterns(models.Models, []string{"!*"}) {
		t.Errorf("/models = %v, want the deny-all sentinel [!*]", models.Models)
	}

	// And the unrestricted form must clear it again, or the key can never be
	// opened from the dashboard.
	putAccess(t, env, id, map[string]any{"restricted": false, "allow": []string{}})
	if after := accessOfKey(t, env, id); after.Restricted {
		t.Errorf("access.restricted = true after writing restricted:false")
	}
	if after := modelsOfKey(t, env, id); len(after.Models) != 0 {
		t.Errorf("/models = %v after clearing, want empty", after.Models)
	}
}

// (d) a corrupt `access.allow` is refused, and the key keeps the policy it
// already had. Turning a restricted key into an unrestricted one on a bad
// payload is the one failure this shape must never have.
func TestKeyAccessCorruptPayloadKeepsRestriction(t *testing.T) {
	env := newEnv(t)
	id := newKeyForAccess(t, env, "access-d")

	seed := []string{"cc/claude-opus-5"}
	putAccess(t, env, id, map[string]any{"restricted": true, "allow": seed})

	corrupt := []struct {
		name string
		body map[string]any
	}{
		{"allow is not an array", map[string]any{"restricted": true, "allow": "gpt-4o"}},
		{"allow holds a non-string", map[string]any{"restricted": true, "allow": []any{"ok", 42}}},
		{"restricted flag is missing", map[string]any{"allow": []any{"gpt-4o"}}},
		{"restricted is not a boolean", map[string]any{"restricted": "yes", "allow": []any{"gpt-4o"}}},
	}
	for _, tc := range corrupt {
		res := env.Put(t, "/api/keys/"+id, map[string]any{"access": tc.body})
		if res.Status != http.StatusBadRequest {
			t.Errorf("%s: PUT access = %d, want 400 (body: %s)", tc.name, res.Status, truncate(res.Body))
		}
	}

	got := accessOfKey(t, env, id)
	if !got.Restricted {
		t.Fatalf("access.restricted = false after four refused payloads; the key was widened")
	}
	if !samePatterns(got.Allow, seed) {
		t.Errorf("access.allow = %v, want the seeded %v", got.Allow, seed)
	}
	if models := modelsOfKey(t, env, id); !samePatterns(models.Models, seed) {
		t.Errorf("/models = %v, want the seeded %v", models.Models, seed)
	}
}
