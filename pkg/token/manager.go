package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Manager issues and validates JWT tokens used for API authentication.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager builds a token manager with the provided secret and TTL.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// TTL exposes the configured token lifetime.
func (m *Manager) TTL() time.Duration {
	if m == nil {
		return 0
	}
	return m.ttl
}

// Claims represents JWT claims embedded in tokens.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Generate issues a signed token for the provided user identifier.
func (m *Manager) Generate(userID string) (string, error) {
	now := time.Now().UTC()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse validates a token string and returns the embedded claims.
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims := parsed.Claims.(*Claims)
	if claims.UserID == "" {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
