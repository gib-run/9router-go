package dashboard

import (
	json "encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postCustomModels(t *testing.T, router http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/models/custom", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

type batchSkip struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
	Code   string `json:"code"`
}

type batchResponse struct {
	Status  string      `json:"status"`
	Added   []string    `json:"added"`
	Skipped []batchSkip `json:"skipped"`
}

func decodeBatch(t *testing.T, rec *httptest.ResponseRecorder) batchResponse {
	t.Helper()
	var out batchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode batch response %s: %v", rec.Body.String(), err)
	}
	return out
}

func skipCode(skips []batchSkip, id string) string {
	for _, skip := range skips {
		if skip.ID == id {
			return skip.Code
		}
	}
	return ""
}

// A batch is one request for a pasted list, so it has to say which lines it kept
// and which it did not: an id that is silently dropped is indistinguishable from
// a typo the operator will keep pasting.
func TestSaveCustomModelsBatch_AddsNewIdsAndReportsTheRest(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	router := setupTestRouter(repo)

	// One model already on this provider, with metadata the batch must not clobber.
	if err := repo.SetKV("customModels", "nara|glm-4.7|llm",
		`{"providerAlias":"nara","id":"glm-4.7","type":"llm","name":"GLM 4.7","contextWindow":128000}`); err != nil {
		t.Fatalf("seed existing model: %v", err)
	}

	rec := postCustomModels(t, router, `{"providerAlias":"nara","models":[
		{"id":"glm-5.3","vision":true,"reasoning":true,"contextWindow":200000,"maxOutput":8192},
		"glm-4.7",
		{"id":"glm-5.3"}]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("bulk save = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	out := decodeBatch(t, rec)
	if out.Status != "ok" {
		t.Errorf("status = %q, want ok", out.Status)
	}
	if len(out.Added) != 1 || out.Added[0] != "glm-5.3" {
		t.Fatalf("added = %v, want exactly [glm-5.3]", out.Added)
	}
	if got := skipCode(out.Skipped, "glm-4.7"); got != "ALREADY_EXISTS" {
		t.Errorf("existing id skipped as %q, want ALREADY_EXISTS", got)
	}
	if got := skipCode(out.Skipped, "glm-5.3"); got != "DUPLICATE_IN_BATCH" {
		t.Errorf("repeated id skipped as %q, want DUPLICATE_IN_BATCH", got)
	}

	scope, err := repo.GetKVScope("customModels")
	if err != nil {
		t.Fatalf("GetKVScope: %v", err)
	}
	if len(scope) != 2 {
		t.Fatalf("stored keys = %d (%v), want the existing row plus the new one", len(scope), scope)
	}
	if !strings.Contains(scope["nara|glm-4.7|llm"], `"GLM 4.7"`) {
		t.Errorf("the existing model was overwritten: %s", scope["nara|glm-4.7|llm"])
	}
	added := scope["nara|glm-5.3|llm"]
	for _, want := range []string{`"contextWindow":200000`, `"maxOutput":8192`, `"vision":true`, `"providerAlias":"nara"`} {
		if !strings.Contains(added, want) {
			t.Errorf("stored model %s does not carry %s", added, want)
		}
	}
}

// The batch applies the same name rules a single save does — asked without
// writing, so a refused line is reported instead of answering the whole paste
// with one 409.
func TestSaveCustomModelsBatch_AppliesTheSameNameRulesAsASingleSave(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	router := setupTestRouter(repo)

	if err := repo.CreateCombo("combo-wombo", "combo-wombo", "llm", `["nara/glm-5.3"]`, "fallback"); err != nil {
		t.Fatalf("seed combo: %v", err)
	}

	rec := postCustomModels(t, router, `{"providerAlias":"nara","models":["combo-wombo","glm-5.3"]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("bulk save = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	out := decodeBatch(t, rec)
	if got := skipCode(out.Skipped, "combo-wombo"); got != "COMBO_NAME_CONFLICT" {
		t.Errorf("colliding id skipped as %q, want COMBO_NAME_CONFLICT", got)
	}
	if len(out.Added) != 1 || out.Added[0] != "glm-5.3" {
		t.Fatalf("added = %v, want the non-colliding id to land anyway", out.Added)
	}

	scope, err := repo.GetKVScope("customModels")
	if err != nil {
		t.Fatalf("GetKVScope: %v", err)
	}
	if _, stored := scope["nara|combo-wombo|llm"]; stored {
		t.Error("the colliding id was stored by the batch path")
	}

	// One rule, two callers: the single save must refuse the same id.
	single := postCustomModels(t, router,
		`{"providerAlias":"nara","id":"combo-wombo","type":"llm"}`)
	if single.Code != http.StatusConflict {
		t.Fatalf("single save of the colliding id = %d, want 409 (body: %s)", single.Code, single.Body.String())
	}
	if !strings.Contains(single.Body.String(), "COMBO_NAME_CONFLICT") {
		t.Errorf("single save 409 body = %s, want it to name the same code", single.Body.String())
	}
}

func TestSaveCustomModelsBatch_RejectsMalformedBodies(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	router := setupTestRouter(repo)

	tests := []struct {
		name string
		body string
	}{
		{"models is not a list", `{"providerAlias":"nara","models":"glm-5.3"}`},
		{"entry is not an id or an object", `{"providerAlias":"nara","models":[42]}`},
		{"alias missing", `{"models":["glm-5.3"]}`},
		{"alias empty", `{"providerAlias":"  ","models":["glm-5.3"]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postCustomModels(t, router, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("bulk save = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
			}
		})
	}
}

// The single-model body is the shape that shipped first; adding the array form
// must not have moved it.
func TestSaveCustomModel_SingleBodyStillWorks(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	router := setupTestRouter(repo)

	byKey := postCustomModels(t, router,
		`{"key":"nara|glm-5.3|llm","value":{"providerAlias":"nara","id":"glm-5.3","type":"llm","name":"glm-5.3"}}`)
	if byKey.Code != http.StatusOK {
		t.Fatalf("save by key = %d, want 200 (body: %s)", byKey.Code, byKey.Body.String())
	}
	byID := postCustomModels(t, router, `{"providerAlias":"nara","id":"glm-4.7","type":"llm"}`)
	if byID.Code != http.StatusOK {
		t.Fatalf("save by id = %d, want 200 (body: %s)", byID.Code, byID.Body.String())
	}

	scope, err := repo.GetKVScope("customModels")
	if err != nil {
		t.Fatalf("GetKVScope: %v", err)
	}
	for _, key := range []string{"nara|glm-5.3|llm", "nara|glm-4.7|llm"} {
		if _, ok := scope[key]; !ok {
			t.Errorf("key %s missing after a single save: %v", key, scope)
		}
	}
}
