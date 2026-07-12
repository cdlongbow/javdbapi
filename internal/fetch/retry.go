package fetch

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"
)

func retryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// temporaryError mirrors the deprecated net.Error.Temporary method via duck
// typing so checking it here does not trip a deprecated-symbol lint warning,
// while still recognizing errors (e.g. connection resets) that only report
// temporariness through Temporary(), not Timeout().
type temporaryError interface {
	Temporary() bool
}

// isRetryableRequestErr classifies a doAttempt transport-level error (not a
// *StatusError, which is classified by retryableStatus instead). Context
// cancellation and deadlines are never retried, even though they otherwise
// satisfy net.Error. A network error is retried only when it reports itself
// as a timeout or as temporary; any other error is treated as permanent.
func isRetryableRequestErr(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var netErr net.Error
	if !errors.As(err, &netErr) {
		return false
	}
	if netErr.Timeout() {
		return true
	}
	var temp temporaryError
	return errors.As(err, &temp) && temp.Temporary()
}

func exponentialDelay(base, max time.Duration, retry int, jitter float64) time.Duration {
	delay := base << retry
	if delay > max || delay <= 0 {
		delay = max
	}
	return time.Duration(float64(delay) * (0.5 + 0.5*jitter))
}

// nextDelay honors a valid Retry-After header before falling back to the
// exponential backoff schedule. The result is always capped at
// c.retryMaxDelay, whether it came from the server or from backoff.
func (c *Client) nextDelay(header http.Header, attempt int) time.Duration {
	if header != nil {
		if raw := header.Get("Retry-After"); raw != "" {
			if secs, err := strconv.Atoi(raw); err == nil && secs >= 0 {
				return capDelay(time.Duration(secs)*time.Second, c.retryMaxDelay)
			}
			if at, err := http.ParseTime(raw); err == nil {
				if d := at.Sub(c.now()); d > 0 {
					return capDelay(d, c.retryMaxDelay)
				}
			}
		}
	}
	return exponentialDelay(c.retryBaseDelay, c.retryMaxDelay, attempt, c.jitter())
}

func capDelay(d, max time.Duration) time.Duration {
	if d > max {
		return max
	}
	return d
}
