package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyDefaultsFillsMissingValues(t *testing.T) {
	var cfg Config
	applyDefaults(&cfg)

	if cfg.App.Name != defaultAppName {
		t.Fatalf("app name default mismatch: got %q", cfg.App.Name)
	}
	if len(cfg.App.Versions) != 1 || cfg.App.Versions[0] != defaultVersion {
		t.Fatalf("app versions default mismatch: %#v", cfg.App.Versions)
	}
	if cfg.App.Secret != defaultAppSecret {
		t.Fatalf("app secret default mismatch: got %q", cfg.App.Secret)
	}
	if cfg.App.Locale != defaultLocale {
		t.Fatalf("app locale default mismatch: got %q", cfg.App.Locale)
	}
	if cfg.HTTP.Port != defaultPort {
		t.Fatalf("http port default mismatch: got %q", cfg.HTTP.Port)
	}
	if cfg.Database.Port != defaultDBPort {
		t.Fatalf("db port default mismatch: got %d", cfg.Database.Port)
	}
	if cfg.Database.LogQueries {
		t.Fatalf("expected log queries to default to false")
	}
	if cfg.External.JSONPlaceholderURL != defaultJSONPlaceholderURL {
		t.Fatalf("external placeholder url default mismatch: got %q", cfg.External.JSONPlaceholderURL)
	}
	if cfg.Auth.AccessTokenTTLMinutes != defaultAccessTokenTTLMinutes {
		t.Fatalf("access ttl default mismatch: got %d", cfg.Auth.AccessTokenTTLMinutes)
	}
	if cfg.Auth.RefreshTokenTTLMinutes != defaultRefreshTokenTTLMinutes {
		t.Fatalf("refresh ttl default mismatch: got %d", cfg.Auth.RefreshTokenTTLMinutes)
	}
	if cfg.Redis.Port != defaultRedisPort {
		t.Fatalf("redis port default mismatch: got %d", cfg.Redis.Port)
	}
	if cfg.Redis.Host != "localhost" {
		t.Fatalf("redis host default mismatch: got %q", cfg.Redis.Host)
	}
	if cfg.Logging.MaxSizeMB != defaultLogSizeMB {
		t.Fatalf("log size default mismatch: got %d", cfg.Logging.MaxSizeMB)
	}
	if cfg.Logging.MaxBackups != defaultLogBackups {
		t.Fatalf("log backups default mismatch: got %d", cfg.Logging.MaxBackups)
	}
	if cfg.Logging.MaxAgeDays != defaultLogAgeDays {
		t.Fatalf("log age default mismatch: got %d", cfg.Logging.MaxAgeDays)
	}
}

func TestApplyDefaultsPreservesExistingValues(t *testing.T) {
	cfg := Config{
		App: AppConfig{
			Name:     "Custom",
			Versions: []string{"v2"},
			Secret:   "secret",
			Locale:   "en",
		},
		HTTP: HTTPConfig{Port: "9090"},
		Database: DatabaseConfig{
			Port:       3307,
			LogQueries: true,
		},
		External: ExternalConfig{JSONPlaceholderURL: "https://example.com/api"},
		Auth: AuthConfig{
			AccessTokenTTLMinutes:  15,
			RefreshTokenTTLMinutes: 120,
		},
		Redis: RedisConfig{
			Host:     "redis.internal",
			Port:     6380,
			Password: "redispass",
			DB:       2,
		},
		Logging: LoggingConfig{
			MaxSizeMB:  20,
			MaxBackups: 7,
			MaxAgeDays: 60,
		},
	}

	applyDefaults(&cfg)

	if cfg.App.Name != "Custom" || cfg.App.Versions[0] != "v2" || cfg.App.Secret != "secret" || cfg.App.Locale != "en" {
		t.Fatalf("app config overwritten: %#v", cfg.App)
	}
	if cfg.HTTP.Port != "9090" {
		t.Fatalf("http port overwritten: got %q", cfg.HTTP.Port)
	}
	if cfg.Database.Port != 3307 || !cfg.Database.LogQueries {
		t.Fatalf("database config overwritten: %#v", cfg.Database)
	}
	if cfg.Logging.MaxSizeMB != 20 || cfg.Logging.MaxBackups != 7 || cfg.Logging.MaxAgeDays != 60 {
		t.Fatalf("logging config overwritten: %#v", cfg.Logging)
	}
	if cfg.External.JSONPlaceholderURL != "https://example.com/api" {
		t.Fatalf("external config overwritten: %#v", cfg.External)
	}
	if cfg.Auth.AccessTokenTTLMinutes != 15 || cfg.Auth.RefreshTokenTTLMinutes != 120 {
		t.Fatalf("auth config overwritten: %#v", cfg.Auth)
	}
	if cfg.Redis.Host != "redis.internal" || cfg.Redis.Port != 6380 || cfg.Redis.Password != "redispass" || cfg.Redis.DB != 2 {
		t.Fatalf("redis config overwritten: %#v", cfg.Redis)
	}
}

func TestLoadReadsEnvironmentConfigFile(t *testing.T) {
	env := "testconfig"
	writeConfigFile(t, env, `app:
  name: Test Service
  versions: ["v3"]
  secret: super-secret
  locale: en
http:
  port: "8088"
database:
  host: db.example.com
  port: 3308
  user: user
  password: pass
  name: testdb
  log_queries: true
logging:
  file: logs/test.log
  max_size_mb: 12
  max_backups: 4
  max_age_days: 15
  compress: true
external:
  jsonplaceholder_url: https://placeholder.test
auth:
  access_token_ttl_minutes: 15
  refresh_token_ttl_minutes: 1440
redis:
  host: redis.example.com
  port: 6380
  password: redispass
  db: 1
`)
	t.Setenv("APP_ENV", env)

	cfg := Load()

	if cfg.App.Name != "Test Service" || cfg.App.Versions[0] != "v3" || cfg.App.Secret != "super-secret" || cfg.App.Locale != "en" {
		t.Fatalf("unexpected app config: %#v", cfg.App)
	}
	if cfg.HTTP.Port != "8088" {
		t.Fatalf("unexpected http port: %q", cfg.HTTP.Port)
	}
	if cfg.Database.Host != "db.example.com" || cfg.Database.Port != 3308 || cfg.Database.User != "user" || cfg.Database.Password != "pass" || cfg.Database.Name != "testdb" || !cfg.Database.LogQueries {
		t.Fatalf("unexpected db config: %#v", cfg.Database)
	}
	if cfg.Logging.File != "logs/test.log" || cfg.Logging.MaxSizeMB != 12 || cfg.Logging.MaxBackups != 4 || cfg.Logging.MaxAgeDays != 15 || !cfg.Logging.Compress {
		t.Fatalf("unexpected logging config: %#v", cfg.Logging)
	}
	if cfg.External.JSONPlaceholderURL != "https://placeholder.test" {
		t.Fatalf("unexpected external config: %#v", cfg.External)
	}
	if cfg.Auth.AccessTokenTTLMinutes != 15 || cfg.Auth.RefreshTokenTTLMinutes != 1440 {
		t.Fatalf("unexpected auth config: %#v", cfg.Auth)
	}
	if cfg.Redis.Host != "redis.example.com" || cfg.Redis.Port != 6380 || cfg.Redis.Password != "redispass" || cfg.Redis.DB != 1 {
		t.Fatalf("unexpected redis config: %#v", cfg.Redis)
	}
}

func TestLoadUsesDefaultEnvironmentFile(t *testing.T) {
	t.Setenv("APP_ENV", "")
	writeConfigFile(t, "local", `app:
  name: Local
  versions: ["v1"]
  secret: s
  locale: id
http:
  port: "9090"
database:
  host: host
  port: 3307
  user: user
  password: pass
  name: localdb
  log_queries: true
logging:
  file: logs/log.log
`)
	cfg := Load()
	if cfg.App.Name == "" || len(cfg.App.Versions) == 0 {
		t.Fatalf("expected default config to load app info")
	}
	if cfg.Database.Host == "" || !cfg.Database.LogQueries {
		t.Fatalf("expected database config to be populated")
	}
}

func TestLoadPanicsWhenFileMissing(t *testing.T) {
	t.Setenv("APP_ENV", "missing-env")
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when config file missing")
		}
	}()
	_ = Load()
}

func writeConfigFile(t *testing.T, env, data string) string {
	t.Helper()
	if err := os.MkdirAll("resources", 0o755); err != nil {
		t.Fatalf("failed to ensure resources dir: %v", err)
	}
	filename := filepath.Join("resources", fmt.Sprintf("config.%s.yaml", env))
	if err := os.WriteFile(filename, []byte(data), 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	t.Cleanup(func() { os.Remove(filename) })
	return filename
}
