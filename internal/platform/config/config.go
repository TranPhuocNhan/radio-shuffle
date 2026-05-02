package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env                    string
	HTTPAddr               string
	ShutdownTimeoutSeconds int
	ReadTimeoutSeconds    int
	WriteTimeoutSeconds   int

	DatabaseURL         string
	MigrationsPath      string
	RunMigrationsOnBoot bool

	JWTSigningKey              string
	JWTAccessTokenTTLMinutes int
	JWTRefreshTokenTTLDays    int
	JWTIssuer                 string

	CORSAllowedOrigins string

	RateLimitRPS float64
	BurstTokens  int

	LogLevel string

	CookieSecure      bool
	CookieSameSite    string
	CookieRefreshName string

	SyncIntervalStr     string
	RadioBrowserBaseURL string
}

func (c Config) ShutdownTimeout() time.Duration {
	sec := c.ShutdownTimeoutSeconds
	if sec <= 0 {
		sec = 15
	}
	return time.Duration(sec) * time.Second
}

func (c Config) ReadTimeout() time.Duration {
	sec := c.ReadTimeoutSeconds
	if sec <= 0 {
		sec = 10
	}
	return time.Duration(sec) * time.Second
}

func (c Config) WriteTimeout() time.Duration {
	sec := c.WriteTimeoutSeconds
	if sec <= 0 {
		sec = 30
	}
	return time.Duration(sec) * time.Second
}

func (c Config) AccessTTL() time.Duration {
	m := c.JWTAccessTokenTTLMinutes
	if m <= 0 {
		m = 15
	}
	return time.Duration(m) * time.Minute
}

func (c Config) RefreshTTL() time.Duration {
	d := c.JWTRefreshTokenTTLDays
	if d <= 0 {
		d = 7
	}
	return time.Duration(d) * 24 * time.Hour
}

func Load() (Config, error) {
	cfg := loadBase()
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSigningKey == "" {
		return Config{}, errors.New("JWT_SIGNING_KEY is required")
	}
	return cfg, nil
}

// LoadSyncer loads config for the syncer binary. JWT fields are not required.
func LoadSyncer() (Config, error) {
	cfg := loadBase()
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func loadBase() Config {
	v := viper.New()
	v.SetConfigType("env")
	v.SetConfigName(".env")
	v.AddConfigPath(".")
	_ = v.ReadInConfig()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := Config{
		HTTPAddr:                 ":8080",
		ShutdownTimeoutSeconds:  15,
		ReadTimeoutSeconds:      10,
		WriteTimeoutSeconds:     30,
		MigrationsPath:           "./db/migrations",
		RunMigrationsOnBoot:     false,
		JWTAccessTokenTTLMinutes: 15,
		JWTRefreshTokenTTLDays:  7,
		JWTIssuer:               "radio-shuffle",
		CORSAllowedOrigins:      "*",
		RateLimitRPS:             20,
		BurstTokens:              60,
		LogLevel:                 "info",
		CookieSameSite:           "Lax",
		CookieRefreshName:       "refresh_token",
		SyncIntervalStr:         "6h",
	}

	cfg.Env = firstNonEmpty(v.GetString("APP_ENV"), getenv("APP_ENV"), "development")
	cfg.DatabaseURL = firstNonEmpty(v.GetString("DATABASE_URL"), getenv("DATABASE_URL"))
	cfg.HTTPAddr = firstNonEmpty(v.GetString("HTTP_ADDR"), getenv("HTTP_ADDR"), cfg.HTTPAddr)
	cfg.RunMigrationsOnBoot = parseBool(firstNonEmpty(v.GetString("RUN_MIGRATIONS_ON_BOOT"), getenv("RUN_MIGRATIONS_ON_BOOT")), cfg.RunMigrationsOnBoot)
	cfg.MigrationsPath = firstNonEmpty(v.GetString("MIGRATIONS_PATH"), getenv("MIGRATIONS_PATH"), cfg.MigrationsPath)
	cfg.LogLevel = firstNonEmpty(v.GetString("LOG_LEVEL"), getenv("LOG_LEVEL"), cfg.LogLevel)

	cfg.JWTSigningKey = firstNonEmpty(v.GetString("JWT_SIGNING_KEY"), getenv("JWT_SIGNING_KEY"))
	cfg.JWTAccessTokenTTLMinutes = atoiDef(firstNonEmpty(v.GetString("JWT_ACCESS_TOKEN_TTL_MINUTES"), getenv("JWT_ACCESS_TOKEN_TTL_MINUTES")), cfg.JWTAccessTokenTTLMinutes)
	cfg.JWTRefreshTokenTTLDays = atoiDef(firstNonEmpty(v.GetString("JWT_REFRESH_TOKEN_TTL_DAYS"), getenv("JWT_REFRESH_TOKEN_TTL_DAYS")), cfg.JWTRefreshTokenTTLDays)
	cfg.JWTIssuer = firstNonEmpty(v.GetString("JWT_ISSUER"), getenv("JWT_ISSUER"), cfg.JWTIssuer)

	cfg.CORSAllowedOrigins = firstNonEmpty(v.GetString("CORS_ALLOWED_ORIGINS"), getenv("CORS_ALLOWED_ORIGINS"), cfg.CORSAllowedOrigins)
	cfg.RateLimitRPS = parseFloat(firstNonEmpty(v.GetString("RATE_LIMIT_RPS"), getenv("RATE_LIMIT_RPS")), cfg.RateLimitRPS)
	cfg.BurstTokens = atoiDef(firstNonEmpty(v.GetString("RATE_LIMIT_BURST"), getenv("RATE_LIMIT_BURST")), cfg.BurstTokens)

	cfg.ShutdownTimeoutSeconds = atoiDef(firstNonEmpty(v.GetString("HTTP_SHUTDOWN_TIMEOUT_SECONDS"), getenv("HTTP_SHUTDOWN_TIMEOUT_SECONDS")), cfg.ShutdownTimeoutSeconds)
	cfg.ReadTimeoutSeconds = atoiDef(firstNonEmpty(v.GetString("HTTP_READ_TIMEOUT_SECONDS"), getenv("HTTP_READ_TIMEOUT_SECONDS")), cfg.ReadTimeoutSeconds)
	cfg.WriteTimeoutSeconds = atoiDef(firstNonEmpty(v.GetString("HTTP_WRITE_TIMEOUT_SECONDS"), getenv("HTTP_WRITE_TIMEOUT_SECONDS")), cfg.WriteTimeoutSeconds)

	if s := firstNonEmpty(v.GetString("AUTH_COOKIE_SECURE"), getenv("AUTH_COOKIE_SECURE")); s != "" {
		cfg.CookieSecure = parseBool(s, cfg.CookieSecure)
	}
	cfg.CookieSameSite = firstNonEmpty(v.GetString("AUTH_COOKIE_SAMESITE"), getenv("AUTH_COOKIE_SAMESITE"), cfg.CookieSameSite)
	cfg.CookieRefreshName = firstNonEmpty(v.GetString("AUTH_REFRESH_COOKIE_NAME"), getenv("AUTH_REFRESH_COOKIE_NAME"), cfg.CookieRefreshName)

	cfg.SyncIntervalStr = firstNonEmpty(v.GetString("SYNC_INTERVAL"), getenv("SYNC_INTERVAL"), cfg.SyncIntervalStr)
	cfg.RadioBrowserBaseURL = firstNonEmpty(v.GetString("RADIO_BROWSER_BASE_URL"), getenv("RADIO_BROWSER_BASE_URL"))

	return cfg
}

func (c Config) SyncInterval() time.Duration {
	d, err := time.ParseDuration(c.SyncIntervalStr)
	if err != nil || d <= 0 {
		return 6 * time.Hour
	}
	return d
}

func (c Config) GinMode() string {
	if strings.EqualFold(strings.TrimSpace(c.Env), "production") {
		return "release"
	}
	return "debug"
}

// CORSAllowedOriginsList splits comma-separated origins for gin-contrib/cors.
func (c Config) CORSAllowedOriginsList() []string {
	if strings.TrimSpace(c.CORSAllowedOrigins) == "" || c.CORSAllowedOrigins == "*" {
		return []string{"*"}
	}
	var out []string
	for _, part := range strings.Split(c.CORSAllowedOrigins, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func getenv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func atoiDef(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func parseFloat(raw string, def float64) float64 {
	if raw == "" {
		return def
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return def
	}
	return n
}

func parseBool(raw string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	default:
		return def
	}
}
