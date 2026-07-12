package javdbapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Locale string

const LocaleZH Locale = "zh"

type ClientConfig struct {
	BaseURL   string
	Locale    Locale
	UserAgent string
	HTTP      HTTPConfig
	Retry     RetryPolicy
	RateLimit RateLimitPolicy
	Logger    *slog.Logger
}

type HTTPConfig struct {
	Client           *http.Client
	Timeout          time.Duration
	ProxyURL         string
	MaxResponseBytes int64
}

type RetryPolicy struct {
	Disabled    bool
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

type RateLimitPolicy struct {
	Disabled          bool
	RequestsPerSecond float64
	Burst             int
}

const (
	defaultBaseURL   = "https://javdb.com"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"

	defaultHTTPTimeout      = 30 * time.Second
	defaultMaxResponseBytes = 8 << 20
	defaultRetryMaxAttempts = 3
	defaultRetryBaseDelay   = 1 * time.Second
	defaultRetryMaxDelay    = 30 * time.Second
	defaultRateLimitRPS     = 1.0
	defaultRateLimitBurst   = 1
)

// normalizeClientConfig validates the raw config, then fills in defaults for
// zero values. Custom-client conflicts are checked against the raw input
// before defaults are applied, so a caller-supplied Timeout/ProxyURL is
// always rejected when HTTP.Client is also set.
func normalizeClientConfig(cfg ClientConfig) (ClientConfig, error) {
	if cfg.HTTP.Client != nil && (cfg.HTTP.Timeout != 0 || cfg.HTTP.ProxyURL != "") {
		return ClientConfig{}, fmt.Errorf("%w: HTTP.Client cannot be combined with Timeout or ProxyURL", ErrInvalidConfig)
	}

	out := cfg

	if out.BaseURL == "" {
		out.BaseURL = defaultBaseURL
	}
	if out.Locale == "" {
		out.Locale = LocaleZH
	}
	if out.Locale != LocaleZH {
		return ClientConfig{}, fmt.Errorf("%w: unsupported locale %q", ErrInvalidConfig, out.Locale)
	}
	if out.UserAgent == "" {
		out.UserAgent = defaultUserAgent
	}

	if out.HTTP.Client == nil {
		switch {
		case out.HTTP.Timeout == 0:
			out.HTTP.Timeout = defaultHTTPTimeout
		case out.HTTP.Timeout < 0:
			return ClientConfig{}, fmt.Errorf("%w: negative HTTP timeout", ErrInvalidConfig)
		}
	}
	switch {
	case out.HTTP.MaxResponseBytes == 0:
		out.HTTP.MaxResponseBytes = defaultMaxResponseBytes
	case out.HTTP.MaxResponseBytes < 0:
		return ClientConfig{}, fmt.Errorf("%w: negative MaxResponseBytes", ErrInvalidConfig)
	}

	if !out.Retry.Disabled {
		switch {
		case out.Retry.MaxAttempts == 0:
			out.Retry.MaxAttempts = defaultRetryMaxAttempts
		case out.Retry.MaxAttempts < 1:
			return ClientConfig{}, fmt.Errorf("%w: retry MaxAttempts must be >= 1", ErrInvalidConfig)
		}
		switch {
		case out.Retry.BaseDelay == 0:
			out.Retry.BaseDelay = defaultRetryBaseDelay
		case out.Retry.BaseDelay < 0:
			return ClientConfig{}, fmt.Errorf("%w: negative retry BaseDelay", ErrInvalidConfig)
		}
		switch {
		case out.Retry.MaxDelay == 0:
			out.Retry.MaxDelay = defaultRetryMaxDelay
		case out.Retry.MaxDelay < 0:
			return ClientConfig{}, fmt.Errorf("%w: negative retry MaxDelay", ErrInvalidConfig)
		}
	}

	if !out.RateLimit.Disabled {
		switch {
		case out.RateLimit.RequestsPerSecond == 0:
			out.RateLimit.RequestsPerSecond = defaultRateLimitRPS
		case out.RateLimit.RequestsPerSecond <= 0:
			return ClientConfig{}, fmt.Errorf("%w: RequestsPerSecond must be > 0", ErrInvalidConfig)
		}
		switch {
		case out.RateLimit.Burst == 0:
			out.RateLimit.Burst = defaultRateLimitBurst
		case out.RateLimit.Burst < 1:
			return ClientConfig{}, fmt.Errorf("%w: Burst must be >= 1", ErrInvalidConfig)
		}
	}

	return out, nil
}
