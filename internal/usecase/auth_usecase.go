package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"skeleton-go/internal/repository"
	"skeleton-go/internal/sessionstore"
	"skeleton-go/pkg/token"
)

var errSessionStoreUnavailable = errors.New("session store unavailable")

// AuthUsecase handles login/token issuance.
type AuthUsecase interface {
	Login(ctx context.Context, input LoginInput) (AuthTokens, error)
	Refresh(ctx context.Context, input RefreshInput) (AuthTokens, error)
	Logout(ctx context.Context, userID string) error
}

// LoginInput collects credentials for authentication.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshInput carries the refresh token.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthTokens groups the response tokens returned to clients.
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type authUsecase struct {
	repo       repository.UserRepository
	tokens     *token.Manager
	sessions   sessionstore.RefreshTokenStore
	refreshTTL time.Duration
}

// NewAuthUsecase constructs the authentication service.
func NewAuthUsecase(repo repository.UserRepository, tokens *token.Manager, sessions sessionstore.RefreshTokenStore, refreshTTL time.Duration) AuthUsecase {
	return &authUsecase{repo: repo, tokens: tokens, sessions: sessions, refreshTTL: refreshTTL}
}

func (a *authUsecase) Login(ctx context.Context, input LoginInput) (AuthTokens, error) {
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)
	if email == "" || password == "" {
		return AuthTokens{}, errors.New("email and password are required")
	}

	user, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		return AuthTokens{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthTokens{}, errors.New("invalid credentials")
	}

	if a.sessions == nil {
		return AuthTokens{}, errSessionStoreUnavailable
	}

	access, err := a.tokens.Generate(user.ID.String())
	if err != nil {
		return AuthTokens{}, err
	}
	refresh, err := a.generateRefreshToken()
	if err != nil {
		return AuthTokens{}, err
	}
	if err := a.sessions.Save(ctx, user.ID.String(), refresh, a.refreshTTL); err != nil {
		return AuthTokens{}, err
	}

	return AuthTokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: a.accessTTLSeconds()}, nil
}

func (a *authUsecase) Refresh(ctx context.Context, input RefreshInput) (AuthTokens, error) {
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return AuthTokens{}, errors.New("refresh token required")
	}

	if a.sessions == nil {
		return AuthTokens{}, errSessionStoreUnavailable
	}

	userID, err := a.sessions.Validate(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, sessionstore.ErrTokenNotFound) {
			return AuthTokens{}, errors.New("invalid refresh token")
		}
		return AuthTokens{}, err
	}

	access, err := a.tokens.Generate(userID)
	if err != nil {
		return AuthTokens{}, err
	}

	newRefresh, err := a.generateRefreshToken()
	if err != nil {
		return AuthTokens{}, err
	}
	if err := a.sessions.Save(ctx, userID, newRefresh, a.refreshTTL); err != nil {
		return AuthTokens{}, err
	}

	return AuthTokens{AccessToken: access, RefreshToken: newRefresh, ExpiresIn: a.accessTTLSeconds()}, nil
}

func (a *authUsecase) Logout(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user id required")
	}
	if a.sessions == nil {
		return errSessionStoreUnavailable
	}
	return a.sessions.Delete(ctx, userID)
}

func (a *authUsecase) generateRefreshToken() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func (a *authUsecase) accessTTLSeconds() int64 {
	if a.tokens == nil {
		return 0
	}
	return int64(a.tokens.TTL().Seconds())
}
