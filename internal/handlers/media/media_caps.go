package media

import (
	"bytes"
	json "encoding/json/v2"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"9router/proxy/internal/handlerutil"
	"9router/proxy/internal/middleware"
)

// Caps for a single media request's input. They are package vars, not consts,
// so a test can lower them instead of allocating a real 25MB part.
var (
	// MediaMaxItems caps how many media items (file parts, or data: URIs in a
	// JSON body) one request may carry. 8 matches the largest fan-out the
	// OpenAI-compatible clients send in practice; beyond it the request is a
	// batching mistake, not a user intent.
	MediaMaxItems = 8
	// MediaMaxItemBytes caps one item. 25MB is OpenAI's own per-file ceiling
	// for vision and audio input.
	MediaMaxItemBytes int64 = 25 << 20
	// MediaMaxTotalBytes caps the summed item size of one request. It sits
	// below the 128MB transport ceiling (middleware.MediaMaxBodySize) on
	// purpose: the transport protects the process, this protects the bill.
	MediaMaxTotalBytes int64 = 64 << 20
)

// mediaInputUse is how much media one request actually carried. It is the
// request-level usage the media lanes report back to the caller, and the same
// measurement the caps are checked against — one measurement, two consumers.
type mediaInputUse struct {
	Items   int
	Bytes   int64
	MaxItem int64
}

// measureMediaInput counts the media items in a request body.
//
// A media item is a multipart part with a filename, or a JSON string value that
// starts with "data:" (an inline data URI). Everything else — the model field,
// a prompt, JSON scaffolding — is not media and is not counted, so a large text
// body is never refused by a media cap.
func measureMediaInput(contentType string, body []byte) (mediaInputUse, error) {
	if strings.Contains(contentType, "multipart/form-data") {
		return measureMultipartInput(contentType, body)
	}
	return measureJSONInput(body), nil
}

func measureMultipartInput(contentType string, body []byte) (mediaInputUse, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil || params["boundary"] == "" {
		return mediaInputUse{}, fmt.Errorf("invalid multipart content type: %w", err)
	}
	var use mediaInputUse
	mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return mediaInputUse{}, err
		}
		// Read every part to completion so the reader advances; only file
		// parts count as media.
		n, err := io.Copy(io.Discard, part)
		// A file part is identified the way the lanes themselves identify it,
		// which is not the same as "has a filename". antigravity_stt selects
		// its audio by `FormName() == "file"` and reads FileName() only for a
		// label, so a part named "file" with no `filename=` attribute is the
		// audio sent upstream — and counting by FileName() alone let it
		// through every cap.
		hasFilename := part.FileName() != ""
		formName := part.FormName()
		_ = part.Close()
		if err != nil {
			return mediaInputUse{}, err
		}
		if !hasFilename && !formNameNamesMedia(formName) {
			continue
		}
		use.recordItem(n)
	}
	return use, nil
}

// mediaFormNames are the form field names the media lanes read binary input
// from. A part carrying one of these is media whether or not the client set a
// filename on it.
var mediaFormNames = map[string]bool{
	"file":   true, // stt, tts reference audio, image edit
	"image":  true,
	"audio":  true,
	"video":  true,
	"input":  true,
	"source": true,
}

func formNameNamesMedia(name string) bool {
	return mediaFormNames[strings.ToLower(name)]
}

func measureJSONInput(body []byte) mediaInputUse {
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		// Not JSON (or not a shape we understand): the lane's own parsing
		// rejects it. Reporting no media is honest; inventing a cap failure
		// would blame the size for a syntax error.
		return mediaInputUse{}
	}
	var use mediaInputUse
	walkInlineMedia(decoded, &use)
	return use
}

// walkInlineMedia counts inline media anywhere in a JSON body.
func walkInlineMedia(v any, use *mediaInputUse) {
	switch node := v.(type) {
	case string:
		if strings.HasPrefix(node, "data:") {
			// The inline payload as sent: the whole data URI. Counting the
			// decoded bytes would understate what this process had to hold.
			use.recordItem(int64(len(node)))
		}
	case []any:
		for _, item := range node {
			walkInlineMedia(item, use)
		}
	case map[string]any:
		walkInlineMediaFields(node, use)
	}
}

// walkInlineMediaFields counts media-bearing object members, including the ones
// that carry raw base64 with no data: prefix.
//
// The image lanes accept bare base64 in `image` / `images` and wrap it into a
// data URL themselves (codex_image.go, antigravity_image.go). Those payloads
// are exactly as large as the data-URI form, but the prefix check missed them,
// so a 5MB bare-base64 image measured as zero media and no cap ever fired.
func walkInlineMediaFields(node map[string]any, use *mediaInputUse) {
	for name, value := range node {
		if isBareMediaField(name) {
			switch v := value.(type) {
			case string:
				if !strings.HasPrefix(v, "data:") && looksLikeBase64Payload(v) {
					use.recordItem(int64(len(v)))
					continue
				}
			case []any:
				// {"image": ["<b64>", "<b64>"]}
				for _, item := range v {
					if s, ok := item.(string); ok && !strings.HasPrefix(s, "data:") && looksLikeBase64Payload(s) {
						use.recordItem(int64(len(s)))
					}
				}
			}
		}
		walkInlineMedia(value, use)
	}
}

// bareMediaFields are the members the image lanes read base64 out of directly.
var bareMediaFields = map[string]bool{
	"image":     true,
	"images":    true,
	"mask":      true,
	"reference": true,
}

func isBareMediaField(name string) bool {
	return bareMediaFields[strings.ToLower(name)]
}

// looksLikeBase64Payload reports whether a string is plausibly raw base64
// rather than ordinary text. It has to be long enough to matter and free of
// whitespace and punctuation that base64 never contains — a prompt mentioning
// an image must not be counted as one.
func looksLikeBase64Payload(s string) bool {
	if len(s) < 1024 {
		return false
	}
	const sample = 256
	for i := 0; i < sample; i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '+', c == '/', c == '=':
		default:
			return false
		}
	}
	return true
}

func (use *mediaInputUse) recordItem(n int64) {
	use.Items++
	use.Bytes += n
	if n > use.MaxItem {
		use.MaxItem = n
	}
}

// capViolation returns the 413 message when the request exceeds a cap, or ""
// when it is within them. The message names the cap and the actual figure: a
// caller told only "too large" cannot tell which limit it hit or by how much.
func (use mediaInputUse) capViolation() string {
	switch {
	case use.Items > MediaMaxItems:
		return fmt.Sprintf("media input too large: request carries %d items, the limit is %d",
			use.Items, MediaMaxItems)
	case use.MaxItem > MediaMaxItemBytes:
		return fmt.Sprintf("media input too large: one item is %s, the per-item limit is %s",
			middleware.HumanBytes(use.MaxItem), middleware.HumanBytes(MediaMaxItemBytes))
	case use.Bytes > MediaMaxTotalBytes:
		return fmt.Sprintf("media input too large: request carries %s of media, the limit is %s",
			middleware.HumanBytes(use.Bytes), middleware.HumanBytes(MediaMaxTotalBytes))
	}
	return ""
}

// copyUpstreamHeaders relays the provider's response headers to the client
// without letting the provider overwrite what the gateway measured.
//
// X-Media-Items / X-Media-Bytes are set by guardMediaInput from the request the
// gateway actually parsed. Copying every upstream header straight into the
// response map let a provider that emits either name replace the measured usage
// with its own number — meaning the usage the gateway claims to own was
// writable by the party it measures.
func copyUpstreamHeaders(dst, src http.Header) {
	for k, v := range src {
		if isGatewayOwnedHeader(k) {
			continue
		}
		dst[k] = v
	}
}

func isGatewayOwnedHeader(name string) bool {
	switch http.CanonicalHeaderKey(name) {
	case "X-Media-Items", "X-Media-Bytes":
		return true
	}
	return false
}

// guardMediaInput measures the request, reports the usage on the response, and
// answers 413 itself when a cap is exceeded — reporting whether it answered.
// Callers run it at the handler boundary, before model resolution and before
// any upstream call, so an oversized request is never billed.
//
// Usage headers are set here, before any write, because the media lanes later
// stream the upstream body straight to the client and headers can no longer
// change once the status line is out.
func guardMediaInput(w http.ResponseWriter, contentType string, body []byte) (mediaInputUse, bool) {
	use, err := measureMediaInput(contentType, body)
	if err != nil {
		// Unmeasurable (a malformed multipart): the lane's own parse rejects
		// it with its usual error. Do not turn that into a cap failure.
		return mediaInputUse{}, false
	}
	if msg := use.capViolation(); msg != "" {
		handlerutil.WriteJSONError(w, http.StatusRequestEntityTooLarge, msg)
		return use, true
	}
	w.Header().Set("X-Media-Items", strconv.Itoa(use.Items))
	w.Header().Set("X-Media-Bytes", strconv.FormatInt(use.Bytes, 10))
	return use, false
}
