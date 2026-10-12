package oauth

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestMusePendingMintGet_roundtrip(t *testing.T) {
	const code = "test-device-code-roundtrip"
	musePendingMints.Delete(code)
	musePendingMints.Store(code, musePendingMint{accountToken: "tok123", createdAt: time.Now()})
	defer musePendingMints.Delete(code)

	pm, ok := musePendingMintGet(code)
	if !ok {
		t.Fatal("expected stored entry to be found")
	}
	if pm.accountToken != "tok123" {
		t.Fatalf("expected accountToken tok123, got %q", pm.accountToken)
	}
}

func TestMusePendingMintGet_expired(t *testing.T) {
	const code = "test-device-code-expired"
	musePendingMints.Store(code, musePendingMint{
		accountToken: "tok",
		createdAt:    time.Now().Add(-musePendingMintTTL - time.Minute),
	})

	if _, ok := musePendingMintGet(code); ok {
		t.Fatal("expected expired entry to be dropped")
	}
	if _, ok := musePendingMints.Load(code); ok {
		t.Fatal("expected expired entry to be deleted on read")
	}
}

func TestMusePendingMintGet_missing(t *testing.T) {
	if _, ok := musePendingMintGet("test-device-code-no-such"); ok {
		t.Fatal("expected miss for unknown code")
	}
}

func TestErrMuseMintTransient_wrapping(t *testing.T) {
	wrapped := fmt.Errorf("Muse Code key mint failed (%w): boom", errMuseMintTransient)
	if !errors.Is(wrapped, errMuseMintTransient) {
		t.Fatal("expected errors.Is to match the transient sentinel")
	}
	if errors.Is(errors.New("plain failure"), errMuseMintTransient) {
		t.Fatal("plain error must not match the transient sentinel")
	}
}
