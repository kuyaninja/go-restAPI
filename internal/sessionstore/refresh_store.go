package sessionstore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrTokenNotFound indicates the refresh token is missing or expired.
	ErrTokenNotFound = errors.New("refresh token not found")
)

// RefreshTokenStore persists refresh tokens and enforces single sessions.
type RefreshTokenStore interface {
	Save(ctx context.Context, userID, token string, ttl time.Duration) error
	Validate(ctx context.Context, token string) (string, error)
	Delete(ctx context.Context, userID string) error
}

// RedisRefreshTokenStore stores refresh tokens inside Redis.
type RedisRefreshTokenStore struct {
	client *redis.Client
	prefix string
}

// NewRedisRefreshTokenStore constructs a Redis-backed refresh token store.
func NewRedisRefreshTokenStore(client *redis.Client) RefreshTokenStore {
	return &RedisRefreshTokenStore{client: client, prefix: "auth"}
}

func (s *RedisRefreshTokenStore) Save(ctx context.Context, userID, token string, ttl time.Duration) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(token) == "" {
		return errors.New("user id and token required")
	}

	userKey := s.userKey(userID)
	existing, err := s.client.Get(ctx, userKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := s.client.TxPipeline()
	if existing != "" {
		pipe.Del(ctx, s.tokenKey(existing))
	}
	pipe.Set(ctx, userKey, token, ttl)
	pipe.Set(ctx, s.tokenKey(token), userID, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisRefreshTokenStore) Validate(ctx context.Context, token string) (string, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return "", ErrTokenNotFound
	}

	userID, err := s.client.Get(ctx, s.tokenKey(trimmed)).Result()
	if err == redis.Nil {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", err
	}

	stored, err := s.client.Get(ctx, s.userKey(userID)).Result()
	if err == redis.Nil || stored != trimmed {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (s *RedisRefreshTokenStore) Delete(ctx context.Context, userID string) error {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return nil
	}

	token, err := s.client.Get(ctx, s.userKey(trimmed)).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.userKey(trimmed))
	pipe.Del(ctx, s.tokenKey(token))
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisRefreshTokenStore) userKey(userID string) string {
	return fmt.Sprintf("%s:user:%s", s.prefix, userID)
}

func (s *RedisRefreshTokenStore) tokenKey(token string) string {
	return fmt.Sprintf("%s:token:%s", s.prefix, token)
}
