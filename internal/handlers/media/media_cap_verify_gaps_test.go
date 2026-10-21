package media

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"sync/atomic"

	"9router/proxy/internal/db"
	"testing"
)

// The tests here pin the ways the first pass at G20c could be walked around:
// a media part with a media field name but no filename, bare base64 in the
// image fields, a request over the sum cap but under every other one, and an
// upstream allowed to rewrite the usage numbers the gateway itself measured.

// noFilenameMultipart builds a body whose media part is named like a media
// field but carries no `filename=` attribute. antigravity_stt selects its audio
// by FormName() == "file", so this is the audio it forwards.
func noFilenameMultipart(t *testing.T, model string, payload []byte) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	if err := mw.WriteField("model", model); err != nil {
		t.Fatalf("write model field: %v", err)
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"`)
	h.Set("Content-Type", "audio/mpeg")
	fw, err := mw.CreatePart(h)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := fw.Write(payload); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	return buf, mw.FormDataContentType()
}

// TestMediaInputCap_CountsMediaPartWithoutFilename is red before the fix: a
// part named "file" with no filename measured as zero media, so any number of
// them — and any size — reached upstream unbilled by the cap.
func TestMediaInputCap_CountsMediaPartWithoutFilename(t *testing.T) {
	lowerMediaCaps(t, 2, 1<<20)
	var hits int32
	upstream := newCountingUpstream(t, &hits)
	defer upstream.Close()

	bodies := make([][]byte, 3)
	for i := range bodies {
		bodies[i] = []byte("tiny")
	}
	// Three separate parts, each named "file" without a filename.
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	_ = mw.WriteField("model", "whisper-1")
	for range bodies {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="file"`)
		fw, err := mw.CreatePart(h)
		if err != nil {
			t.Fatalf("create part: %v", err)
		}
		_, _ = fw.Write([]byte("tiny"))
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	rec := runSTT(t, upstream.URL, buf, mw.FormDataContentType())

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("3 filename-less file parts with cap 2: expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := atomic.LoadInt32(&hits); got != 0 {
		t.Errorf("the part reached upstream %d time(s); the cap must count it before any upstream call", got)
	}
}

// TestMediaInputCap_RefusesOversizedFilenamelessPart is the size half of the
// same bypass: one part over the per-item ceiling, still no filename.
func TestMediaInputCap_RefusesOversizedFilenamelessPart(t *testing.T) {
	lowerMediaCaps(t, 16, 4096)
	var hits int32
	upstream := newCountingUpstream(t, &hits)
	defer upstream.Close()

	body, ct := noFilenameMultipart(t, "whisper-1", bytes.Repeat([]byte("a"), 8192))
	rec := runSTT(t, upstream.URL, body, ct)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("8KB filename-less file part with 4KB cap: expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := atomic.LoadInt32(&hits); got != 0 {
		t.Errorf("oversized part reached upstream %d time(s)", got)
	}
}

// TestMediaInputCap_CountsBareBase64ImageFields is red before the fix: the
// image lanes accept bare base64 and wrap it into a data URI themselves, but
// the measurement only counted strings starting with "data:", so a multi-megabyte
// bare-base64 image measured as zero media.
func TestMediaInputCap_CountsBareBase64ImageFields(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("a"), 8<<10))
	if len(payload) < 1024 {
		t.Fatalf("fixture too small to be treated as base64: %d bytes", len(payload))
	}

	use, err := measureMediaInput("application/json", []byte(`{"model":"x","image":"`+payload+`"}`))
	if err != nil {
		t.Fatalf("measure: %v", err)
	}
	if use.Items != 1 || use.Bytes != int64(len(payload)) {
		t.Errorf("bare base64 in \"image\": measured %d item(s) / %d bytes, want 1 / %d",
			use.Items, use.Bytes, len(payload))
	}

	// The list form the image edit lanes accept.
	use, err = measureMediaInput("application/json", []byte(`{"images":["`+payload+`","`+payload+`"]}`))
	if err != nil {
		t.Fatalf("measure: %v", err)
	}
	if use.Items != 2 {
		t.Errorf("bare base64 in \"images\": measured %d items, want 2", use.Items)
	}
}

// TestMediaInputCap_IgnoresProseInImageField guards the other direction: the
// base64 sniff must not treat an ordinary sentence in a media-named field as
// media, or every request mentioning an image would count as one.
func TestMediaInputCap_IgnoresProseInImageField(t *testing.T) {
	prose := strings.Repeat("please make the sky a softer blue, ", 60) // >1024 bytes
	use, err := measureMediaInput("application/json", []byte(`{"image":"`+prose+`"}`))
	if err != nil {
		t.Fatalf("measure: %v", err)
	}
	if use.Items != 0 {
		t.Errorf("prose in \"image\" measured as %d media item(s), want 0", use.Items)
	}
}

// TestMediaInputCap_RefusesTotalBytesOverCeiling covers MediaMaxTotalBytes,
// which had no test at all: each item fits, the count fits, the sum does not.
func TestMediaInputCap_RefusesTotalBytesOverCeiling(t *testing.T) {
	prevTotal := MediaMaxTotalBytes
	MediaMaxTotalBytes = 2048
	t.Cleanup(func() { MediaMaxTotalBytes = prevTotal })
	lowerMediaCaps(t, 16, 1<<20)

	var hits int32
	upstream := newCountingUpstream(t, &hits)
	defer upstream.Close()

	body, ct := sttMultipart(t, "whisper-1", [][]byte{
		bytes.Repeat([]byte("a"), 1500),
		bytes.Repeat([]byte("b"), 1500),
	})
	rec := runSTT(t, upstream.URL, body, ct)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("3KB total with a 2KB sum cap: expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := atomic.LoadInt32(&hits); got != 0 {
		t.Errorf("over-sum request reached upstream %d time(s)", got)
	}
	if !strings.Contains(rec.Body.String(), "of media") {
		t.Errorf("413 does not name the sum cap: %s", rec.Body.String())
	}
}

// TestMediaUsage_UpstreamCannotOverwriteUsageHeaders is red before the fix:
// every upstream header was copied into the response map unconditionally, so a
// provider emitting X-Media-Items replaced the number the gateway measured.
func TestMediaUsage_UpstreamCannotOverwriteUsageHeaders(t *testing.T) {
	var hits int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Media-Items", "999")
		w.Header().Set("X-Media-Bytes", "999999")
		w.Header().Set("X-Upstream-Trace", "trace-1")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"text":"ok"}`))
	}))
	defer upstream.Close()

	body, ct := sttMultipart(t, "whisper-1", [][]byte{[]byte("abc")})
	rec := runSTT(t, upstream.URL, body, ct)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Media-Items"); got != "1" {
		t.Errorf("X-Media-Items = %q, want the measured 1 — the upstream overwrote it", got)
	}
	if got := rec.Header().Get("X-Media-Bytes"); got != "3" {
		t.Errorf("X-Media-Bytes = %q, want the measured 3 — the upstream overwrote it", got)
	}
	// The fix must not be "drop every upstream header".
	if got := rec.Header().Get("X-Upstream-Trace"); got != "trace-1" {
		t.Errorf("X-Upstream-Trace = %q, want trace-1: an unrelated upstream header was dropped", got)
	}
}

// TestMediaUsage_SystemoneAndMiMoCarryUsage pins that the two lanes that leave
// forwardMediaRequest still measure their request. They were the lanes the "one
// funnel" claim silently skipped.
func TestMediaUsage_SystemoneAndMiMoCarryUsage(t *testing.T) {
	var hits int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"audio":{"data":"AAAA","format":"wav"}}}]}`))
	}))
	defer upstream.Close()

	t.Run("Systemone", func(t *testing.T) {
		database, cleanup := setupMultimodalTestDB(t)
		defer cleanup()
		seedOpenAIConn(t, database, upstream.URL)
		handler := newTestMediaHandler(db.NewRepo(database))

		body := `{"model":"openai/gpt-4o","state":"s","questions":{"q":"a"}}`
		req := httptest.NewRequest(http.MethodPost, "/v1/systemone", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.HandleSystemone(rec, req)

		if rec.Code >= 500 {
			t.Fatalf("systemone lane failed: %d %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("X-Media-Items") == "" {
			t.Error("the systemone lane reports no request-level usage; it bypasses the funnel")
		}
	})

	t.Run("MiMoSpeech", func(t *testing.T) {
		database, cleanup := setupMultimodalTestDB(t)
		defer cleanup()
		seedMiMoConn(t, database, upstream.URL)
		handler := newTestMediaHandler(db.NewRepo(database))

		body := `{"model":"xiaomi-mimo/tts","input":"hello"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/audio/speech", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.HandleAudioSpeech(rec, req)

		if rec.Code >= 500 {
			t.Fatalf("MiMo speech lane failed: %d %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("X-Media-Items") == "" {
			t.Error("the MiMo speech lane reports no request-level usage; it bypasses the funnel")
		}
	})
}
