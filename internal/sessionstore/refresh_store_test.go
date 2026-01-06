package sessionstore

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	server := miniredis.RunT(t)
	host, portStr, _ := net.SplitHostPort(server.Addr())
	port, _ := strconv.Atoi(portStr)
	client := redis.NewClient(&redis.Options{Addr: net.JoinHostPort(host, strconv.Itoa(port))})
	t.Cleanup(func() {
		client.Close()
		server.Close()
	})
	return client
}

func TestRedisRefreshTokenStoreSaveAndValidate(t *testing.T) {
	client := newRedisClient(t)
	store := NewRedisRefreshTokenStore(client)

	ctx := context.Background()
	if err := store.Save(ctx, "user1", "token-1", time.Minute); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	userID, err := store.Validate(ctx, "token-1")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if userID != "user1" {
		t.Fatalf("unexpected user id: %s", userID)
	}
}

func TestRedisRefreshTokenStoreOverwriteRemovesOldToken(t *testing.T) {
	client := newRedisClient(t)
	store := NewRedisRefreshTokenStore(client)
	ctx := context.Background()

	if err := store.Save(ctx, "user1", "token-a", time.Minute); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if err := store.Save(ctx, "user1", "token-b", time.Minute); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if _, err := store.Validate(ctx, "token-a"); err != ErrTokenNotFound {
		t.Fatalf("expected token-a invalid, got %v", err)
	}
	if _, err := store.Validate(ctx, "token-b"); err != nil {
		t.Fatalf("expected token-b valid: %v", err)
	}
}

func TestRedisRefreshTokenStoreDelete(t *testing.T) {
	client := newRedisClient(t)
	store := NewRedisRefreshTokenStore(client)
	ctx := context.Background()

	if err := store.Save(ctx, "user1", "token-a", time.Minute); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if err := store.Delete(ctx, "user1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := store.Validate(ctx, "token-a"); err != ErrTokenNotFound {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}
