package fetch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeNetError lets tests simulate transport-level failures (net.Error) with
// a controlled Timeout/Temporary classification, without depending on real
// network conditions.
type fakeNetError struct {
	msg       string
	timeout   bool
	temporary bool
}

func (e fakeNetError) Error() string   { return e.msg }
func (e fakeNetError) Timeout() bool   { return e.timeout }
func (e fakeNetError) Temporary() bool { return e.temporary }

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return u
}

func newTestClient(t *testing.T, cfg Config) *Client {
	t.Helper()
	if cfg.UserAgent == "" {
		cfg.UserAgent = "javdbapi-test"
	}
	if cfg.MaxResponseBytes == 0 {
		cfg.MaxResponseBytes = 1 << 20
	}
	if cfg.sleep == nil {
		cfg.sleep = func(context.Context, time.Duration) error { return nil }
	}
	if cfg.jitter == nil {
		cfg.jitter = func() float64 { return 0 }
	}
	client, err := New(cfg)
	require.NoError(t, err)
	return client
}

// roundTripFunc lets tests redirect every request to a local httptest server
// regardless of the requested host, so Get can be exercised with realistic
// site URLs instead of the server's own address.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func redirectingClient(t *testing.T, server *httptest.Server) *http.Client {
	t.Helper()
	serverURL := mustURL(t, server.URL)
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.URL.Scheme = serverURL.Scheme
			req.URL.Host = serverURL.Host
			return http.DefaultTransport.RoundTrip(req)
		}),
	}
}

func clientReturningStatus(t *testing.T, status int) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)

	return newTestClient(t, Config{
		HTTPClient:        redirectingClient(t, server),
		RetryDisabled:     true,
		RateLimitDisabled: true,
	})
}

func TestGetRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "12345")
	}))
	t.Cleanup(server.Close)
	client := newTestClient(t, Config{MaxResponseBytes: 4, RetryDisabled: true, RateLimitDisabled: true})
	_, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.ErrorIs(t, err, ErrResponseTooLarge)
}

func TestGetMapsStatus(t *testing.T) {
	client := clientReturningStatus(t, http.StatusTooManyRequests)
	_, err := client.Get(context.Background(), mustURL(t, "https://javdb.com/v/x"))
	var statusErr *StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusTooManyRequests, statusErr.StatusCode)
}

func TestGetRetriesOnTooManyRequestsThenSucceeds(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(server.Close)

	var slept []time.Duration
	client := newTestClient(t, Config{
		RateLimitDisabled: true,
		RetryMaxAttempts:  3,
		sleep: func(_ context.Context, d time.Duration) error {
			slept = append(slept, d)
			return nil
		},
	})

	body, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
	assert.EqualValues(t, 2, atomic.LoadInt32(&attempts))
	assert.Len(t, slept, 1)
}

func TestRetryAfterSecondsOverridesExponentialDelay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	var delays []time.Duration
	client := newTestClient(t, Config{
		RateLimitDisabled: true,
		RetryMaxAttempts:  2,
		RetryBaseDelay:    time.Millisecond,
		sleep: func(_ context.Context, d time.Duration) error {
			delays = append(delays, d)
			return nil
		},
	})

	_, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.Error(t, err)
	require.Len(t, delays, 1)
	assert.Equal(t, 7*time.Second, delays[0])
}

func TestRetryAfterHTTPDateOverridesExponentialDelay(t *testing.T) {
	fixedNow := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	retryAt := fixedNow.Add(5 * time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", retryAt.Format(http.TimeFormat))
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	var delays []time.Duration
	client := newTestClient(t, Config{
		RateLimitDisabled: true,
		RetryMaxAttempts:  2,
		RetryBaseDelay:    time.Millisecond,
		now:               func() time.Time { return fixedNow },
		sleep: func(_ context.Context, d time.Duration) error {
			delays = append(delays, d)
			return nil
		},
	})

	_, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.Error(t, err)
	require.Len(t, delays, 1)
	assert.Equal(t, 5*time.Second, delays[0])
}

func TestGetStopsAfterOneAttemptOnNotFound(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, Config{RateLimitDisabled: true, RetryMaxAttempts: 3})
	_, err := client.Get(context.Background(), mustURL(t, server.URL))

	var statusErr *StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusNotFound, statusErr.StatusCode)
	assert.EqualValues(t, 1, atomic.LoadInt32(&attempts))
}

func TestGetStopsBeforeNextAttemptWhenContextCanceled(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithCancel(context.Background())
	client := newTestClient(t, Config{
		RateLimitDisabled: true,
		RetryMaxAttempts:  3,
		sleep: func(ctx context.Context, _ time.Duration) error {
			cancel()
			return ctx.Err()
		},
	})

	_, err := client.Get(ctx, mustURL(t, server.URL))
	require.Error(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt32(&attempts))
}

func TestGetRetryDelayCappedAtMaxDelay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	var delays []time.Duration
	client := newTestClient(t, Config{
		RateLimitDisabled: true,
		RetryMaxAttempts:  5,
		RetryBaseDelay:    time.Second,
		RetryMaxDelay:     2 * time.Second,
		jitter:            func() float64 { return 1 },
		sleep: func(_ context.Context, d time.Duration) error {
			delays = append(delays, d)
			return nil
		},
	})

	_, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.Error(t, err)
	require.NotEmpty(t, delays)
	for _, d := range delays {
		assert.LessOrEqual(t, d, 2*time.Second)
	}
}

// trackedBody wraps a response body so tests can verify every retry attempt's
// body is fully drained (read to EOF) and closed before the next attempt.
type trackedBody struct {
	io.Reader
	closer io.Closer
	closed *int32
}

func (b trackedBody) Close() error {
	atomic.AddInt32(b.closed, 1)
	return b.closer.Close()
}

type eofTrackingReader struct {
	r       io.Reader
	drained *int32
}

func (r *eofTrackingReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if err == io.EOF {
		atomic.AddInt32(r.drained, 1)
	}
	return n, err
}

func TestGetRetryResponseBodiesAreDrainedAndClosed(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, "try again later")
			return
		}
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(server.Close)

	var closed, drained int32
	base := redirectingClient(t, server)
	baseTransport := base.Transport
	trackingTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		resp, err := baseTransport.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		resp.Body = trackedBody{
			Reader: &eofTrackingReader{r: resp.Body, drained: &drained},
			closer: resp.Body,
			closed: &closed,
		}
		return resp, nil
	})

	client := newTestClient(t, Config{
		HTTPClient:        &http.Client{Transport: trackingTransport},
		RateLimitDisabled: true,
		RetryMaxAttempts:  2,
		RetryBaseDelay:    time.Millisecond,
	})

	body, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
	assert.EqualValues(t, 2, atomic.LoadInt32(&closed))
	assert.GreaterOrEqual(t, atomic.LoadInt32(&drained), int32(2))
}

func TestGetWaitsOnLimiterEveryAttempt(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&attempts, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(server.Close)

	var waits int32
	client := newTestClient(t, Config{
		HTTPClient:       redirectingClient(t, server),
		RetryMaxAttempts: 3,
		rateWait: func(context.Context) error {
			atomic.AddInt32(&waits, 1)
			return nil
		},
	})

	body, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
	assert.EqualValues(t, 3, atomic.LoadInt32(&waits), "limiter must be awaited before every attempt, including retries")
}

func TestGetRetriesTemporaryNetworkErrorThenSucceeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(server.Close)

	base := redirectingClient(t, server)
	baseTransport := base.Transport
	var attempts int32
	flaky := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			return nil, fakeNetError{msg: "connection reset", temporary: true}
		}
		return baseTransport.RoundTrip(req)
	})

	client := newTestClient(t, Config{
		HTTPClient:        &http.Client{Transport: flaky},
		RateLimitDisabled: true,
		RetryMaxAttempts:  2,
		RetryBaseDelay:    time.Millisecond,
	})

	body, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
	assert.EqualValues(t, 2, atomic.LoadInt32(&attempts))
}

func TestGetDoesNotRetryPermanentNetworkError(t *testing.T) {
	var attempts int32
	permanent := roundTripFunc(func(*http.Request) (*http.Response, error) {
		atomic.AddInt32(&attempts, 1)
		return nil, fakeNetError{msg: "no route to host"}
	})

	client := newTestClient(t, Config{
		HTTPClient:        &http.Client{Transport: permanent},
		RateLimitDisabled: true,
		RetryMaxAttempts:  3,
	})

	_, err := client.Get(context.Background(), mustURL(t, "https://javdb.com/v/x"))
	require.Error(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt32(&attempts), "a permanent network error must not be retried")
}

func TestGetDoesNotRetryOnContextCanceled(t *testing.T) {
	var attempts int32
	canceling := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&attempts, 1)
		return nil, req.Context().Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(t, Config{
		HTTPClient:        &http.Client{Transport: canceling},
		RateLimitDisabled: true,
		RetryMaxAttempts:  3,
	})

	_, err := client.Get(ctx, mustURL(t, "https://javdb.com/v/x"))
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.EqualValues(t, 1, atomic.LoadInt32(&attempts), "context cancellation must not be retried even though it satisfies net.Error")
}

func TestRetryAfterCappedWhenExceedsMaxDelay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "100")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	var delays []time.Duration
	client := newTestClient(t, Config{
		RateLimitDisabled: true,
		RetryMaxAttempts:  2,
		RetryMaxDelay:     3 * time.Second,
		sleep: func(_ context.Context, d time.Duration) error {
			delays = append(delays, d)
			return nil
		},
	})

	_, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.Error(t, err)
	require.Len(t, delays, 1)
	assert.Equal(t, 3*time.Second, delays[0], "Retry-After above RetryMaxDelay must be truncated to RetryMaxDelay")
}
