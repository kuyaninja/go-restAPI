package token

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParse(t *testing.T) {
	manager := NewManager("secret", time.Minute)
	tokenStr, err := manager.Generate("user-1")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	claims, err := manager.Parse(tokenStr)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("expected user id, got %q", claims.UserID)
	}
}

func TestParseRejectsInvalidSigningMethod(t *testing.T) {
	manager := NewManager("secret", time.Minute)
	badToken := strings.Join([]string{
		"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9",
		"eyJ1c2VyX2lkIjoiMSJ9",
		"signature",
	}, ".")
	if _, err := manager.Parse(badToken); err == nil {
		t.Fatalf("expected error for RS token")
	}
}

func TestParseRejectsMalformedToken(t *testing.T) {
	manager := NewManager("secret", time.Minute)
	if _, err := manager.Parse("invalid"); err == nil {
		t.Fatalf("expected parse error for malformed token")
	}
}

func TestParseRejectsEmptyUserID(t *testing.T) {
	manager := NewManager("secret", time.Minute)
	tokenStr, err := manager.Generate("")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if _, err := manager.Parse(tokenStr); err == nil {
		t.Fatalf("expected error for empty user id")
	}
}

func TestParseRejectsRS256Token(t *testing.T) {
	manager := NewManager("secret", time.Minute)
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"user_id": "abc"})
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign rsa token: %v", err)
	}
	if _, err := manager.Parse(signed); err == nil {
		t.Fatalf("expected error for RSA token")
	}
}

func TestManagerTTL(t *testing.T) {
	expected := 2 * time.Hour
	manager := NewManager("secret", expected)
	if manager.TTL() != expected {
		t.Fatalf("ttl mismatch: got %v want %v", manager.TTL(), expected)
	}
	var nilManager *Manager
	if got := nilManager.TTL(); got != 0 {
		t.Fatalf("expected zero ttl for nil manager, got %v", got)
	}
}
