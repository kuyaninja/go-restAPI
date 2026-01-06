package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config aggregates runtime configuration values grouped by concern.
type Config struct {
	App      AppConfig      `yaml:"app"`
	HTTP     HTTPConfig     `yaml:"http"`
	Database DatabaseConfig `yaml:"database"`
	External ExternalConfig `yaml:"external"`
	Auth     AuthConfig     `yaml:"auth"`
	Redis    RedisConfig    `yaml:"redis"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// AppConfig holds identifying metadata for the service.
type AppConfig struct {
	Name     string   `yaml:"name"`
	Versions []string `yaml:"versions"`
	Secret   string   `yaml:"secret"`
	Locale   string   `yaml:"locale"`
}

// HTTPConfig describes the HTTP server settings.
type HTTPConfig struct {
	Port string `yaml:"port"`
}

// DatabaseConfig stores connection details for the primary database.
type DatabaseConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	Name       string `yaml:"name"`
	SSLMode    string `yaml:"sslmode"`
	LogQueries bool   `yaml:"log_queries"`
}

// ExternalConfig stores third-party API settings.
type ExternalConfig struct {
	JSONPlaceholderURL string `yaml:"jsonplaceholder_url"`
}

// AuthConfig stores token expiration settings.
type AuthConfig struct {
	AccessTokenTTLMinutes  int `yaml:"access_token_ttl_minutes"`
	RefreshTokenTTLMinutes int `yaml:"refresh_token_ttl_minutes"`
}

// RedisConfig stores Redis connection settings.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// LoggingConfig stores log destinations.
type LoggingConfig struct {
	File       string `yaml:"file"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Compress   bool   `yaml:"compress"`
}

const (
	defaultEnv                    = "local"
	configDir                     = "resources"
	configTpl                     = "config.%s.yaml"
	defaultPort                   = "8080"
	defaultAppName                = "Skeleton API"
	defaultVersion                = "v1"
	defaultLocale                 = "id"
	defaultAppSecret              = "change-me"
	defaultDBPort                 = 5432
	defaultJSONPlaceholderURL     = "https://jsonplaceholder.typicode.com"
	defaultAccessTokenTTLMinutes  = 60
	defaultRefreshTokenTTLMinutes = 60 * 24 * 7
	defaultRedisPort              = 6379
	defaultLogSizeMB              = 10
	defaultLogBackups             = 5
	defaultLogAgeDays             = 30
)

// Load resolves the environment config file and parses it into Config.
// The target file is resources/config.<env>.yaml where env defaults to "local".
func Load() *Config {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = defaultEnv
	}

	filename := fmt.Sprintf("%s/%s", configDir, fmt.Sprintf(configTpl, env))

	data, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Errorf("config: failed to read %s: %w", filename, err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Errorf("config: failed to decode %s: %w", filename, err))
	}

	applyDefaults(&cfg)

	return &cfg
}

func applyDefaults(cfg *Config) {
	if cfg.App.Name == "" {
		cfg.App.Name = defaultAppName
	}

	if len(cfg.App.Versions) == 0 {
		cfg.App.Versions = []string{defaultVersion}
	}

	if cfg.App.Secret == "" {
		cfg.App.Secret = defaultAppSecret
	}

	if cfg.App.Locale == "" {
		cfg.App.Locale = defaultLocale
	}

	if cfg.HTTP.Port == "" {
		cfg.HTTP.Port = defaultPort
	}

	if cfg.Database.Port == 0 {
		cfg.Database.Port = defaultDBPort
	}

	if cfg.External.JSONPlaceholderURL == "" {
		cfg.External.JSONPlaceholderURL = defaultJSONPlaceholderURL
	}

	if cfg.Auth.AccessTokenTTLMinutes == 0 {
		cfg.Auth.AccessTokenTTLMinutes = defaultAccessTokenTTLMinutes
	}

	if cfg.Auth.RefreshTokenTTLMinutes == 0 {
		cfg.Auth.RefreshTokenTTLMinutes = defaultRefreshTokenTTLMinutes
	}

	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = defaultRedisPort
	}

	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}

	if cfg.Logging.MaxSizeMB == 0 {
		cfg.Logging.MaxSizeMB = defaultLogSizeMB
	}

	if cfg.Logging.MaxBackups == 0 {
		cfg.Logging.MaxBackups = defaultLogBackups
	}

	if cfg.Logging.MaxAgeDays == 0 {
		cfg.Logging.MaxAgeDays = defaultLogAgeDays
	}
}
