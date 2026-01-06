package cache

import (
	"net"
	"strconv"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"

	"skeleton-go/pkg/config"
)

func TestNewRedisConnectsSuccessfully(t *testing.T) {
	server := miniredis.RunT(t)
	host, portStr, err := net.SplitHostPort(server.Addr())
	if err != nil {
		t.Fatalf("failed to parse miniredis addr: %v", err)
	}
	port, _ := strconv.Atoi(portStr)

	cfg := config.RedisConfig{Host: host, Port: port}
	client, err := NewRedis(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	t.Cleanup(func() { client.Close() })
}

func TestNewRedisFailsWhenPingFails(t *testing.T) {
	server := miniredis.RunT(t)
	host, portStr, _ := net.SplitHostPort(server.Addr())
	port, _ := strconv.Atoi(portStr)
	server.Close()

	cfg := config.RedisConfig{Host: host, Port: port}
	if _, err := NewRedis(cfg); err == nil {
		t.Fatalf("expected error when redis unreachable")
	}
}
