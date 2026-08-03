package auth

import (
	"context"
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
