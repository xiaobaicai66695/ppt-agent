package auth

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

func TestValidateSessionDoesNotRequireDatabase(t *testing.T) {
	previousDB := db.DB
	db.DB = nil
	t.Cleanup(func() { db.DB = previousDB })

	want := &db.User{ID: 42, Email: "user@example.com", IsAdmin: true}
	token, err := createToken(want)
	if err != nil {
		t.Fatalf("createToken() error = %v", err)
	}
	got, err := ValidateSession(token)
	if err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}
	if got.ID != want.ID || got.Email != want.Email || got.IsAdmin != want.IsAdmin {
		t.Fatalf("ValidateSession() = %#v, want %#v", got, want)
	}
}

func TestValidateSessionRejectsTamperedToken(t *testing.T) {
	token, err := createToken(&db.User{ID: 7, Email: "user@example.com"})
	if err != nil {
		t.Fatalf("createToken() error = %v", err)
	}
	token = token[:len(token)-1] + "x"
	if _, err := ValidateSession(token); err == nil {
		t.Fatal("ValidateSession() accepted a tampered token")
	}
}

func TestWithUserCarriesAuthorizationClaims(t *testing.T) {
	want := &db.User{ID: 9, Email: "admin@example.com", IsAdmin: true}
	ctx := WithUser(context.Background(), want)
	if id, ok := UserIDFromContext(ctx); !ok || id != int(want.ID) {
		t.Fatalf("UserIDFromContext() = %d, %v", id, ok)
	}
	if isAdmin, ok := IsAdminFromContext(ctx); !ok || !isAdmin {
		t.Fatalf("IsAdminFromContext() = %v, %v", isAdmin, ok)
	}
}

func TestGuestLoginEnabledDefaultsToOnAndHonorsDisableFlag(t *testing.T) {
	t.Setenv("GUEST_LOGIN_ENABLED", "")
	if !GuestLoginEnabled() {
		t.Fatal("guest login should be enabled by default")
	}
	t.Setenv("GUEST_LOGIN_ENABLED", "false")
	if GuestLoginEnabled() {
		t.Fatal("guest login should honor disable flag")
	}
}

func TestIsGuestEmail(t *testing.T) {
	if !IsGuestEmail("guest-0123abcd@guest.local") {
		t.Fatal("expected generated guest address to be recognized")
	}
	if IsGuestEmail("member@example.com") {
		t.Fatal("regular address must not be recognized as guest")
	}
}

func TestGuestIPFingerprintIsStableAndDoesNotExposeIPAddress(t *testing.T) {
	t.Setenv("GUEST_IP_HASH_SECRET", "test-only-secret")
	first := guestIPFingerprint("203.0.113.24")
	if first == "" || first != guestIPFingerprint("203.0.113.24") {
		t.Fatalf("guestIPFingerprint should be stable, got %q", first)
	}
	if first == guestIPFingerprint("203.0.113.25") {
		t.Fatal("different IPs must not produce the same fingerprint")
	}
	if strings.Contains(first, "203.0.113.24") {
		t.Fatalf("fingerprint must not contain the raw IP: %q", first)
	}
	if got := guestIPFingerprint("not-an-ip"); got != "" {
		t.Fatalf("invalid address fingerprint = %q, want empty", got)
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "compliant", password: "Deckform2026", wantErr: false},
		{name: "too short", password: "Deck1a", wantErr: true},
		{name: "missing uppercase", password: "deckform2026", wantErr: true},
		{name: "missing lowercase", password: "DECKFORM2026", wantErr: true},
		{name: "missing digit", password: "DeckformPassword", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePassword(tt.password); (err != nil) != tt.wantErr {
				t.Fatalf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}
