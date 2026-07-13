package fetch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrawlTransportSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req trawlRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "request.get", req.Cmd)
		assert.Equal(t, "https://javdb.com/v/test123", req.URL)

		resp := trawlResponse{
			Status: "ok",
			Solution: trawlSolution{
				URL:      req.URL,
				Status:   200,
				Response: "<html><body>hello trawl</body></html>",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, Config{
		TrawlURL:         server.URL,
		RateLimitDisabled: true,
		RetryDisabled:    true,
	})

	body, err := client.Get(context.Background(), mustURL(t, "https://javdb.com/v/test123"))
	require.NoError(t, err)
	assert.Equal(t, "<html><body>hello trawl</body></html>", string(body))
}

func TestTrawlTransportMapsErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := trawlResponse{
			Status: "ok",
			Solution: trawlSolution{
				URL:      "https://javdb.com/v/test404",
				Status:   404,
				Response: "<html>not found</html>",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, Config{
		TrawlURL:         server.URL,
		RateLimitDisabled: true,
		RetryDisabled:    true,
	})

	_, err := client.Get(context.Background(), mustURL(t, "https://javdb.com/v/test404"))
	require.Error(t, err)
	var statusErr *StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 404, statusErr.StatusCode)
}

func TestTrawlTransportErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := trawlResponse{
			Status:  "error",
			Message: "browser pool exhausted",
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, Config{
		TrawlURL:         server.URL,
		RateLimitDisabled: true,
		RetryDisabled:    true,
	})

	_, err := client.Get(context.Background(), mustURL(t, "https://javdb.com/v/x"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "browser pool exhausted")
}

func TestTrawlTransportRetriesOn429(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := int(attempts)
		attempts++
		if count == 0 {
			resp := trawlResponse{
				Status: "ok",
				Solution: trawlSolution{
					URL:      "https://javdb.com/v/x",
					Status:   429,
					Response: "",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(resp))
			return
		}
		resp := trawlResponse{
			Status: "ok",
			Solution: trawlSolution{
				URL:      "https://javdb.com/v/x",
				Status:   200,
				Response: "<html>ok</html>",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, Config{
		TrawlURL:        server.URL,
		RetryMaxAttempts: 3,
		RetryBaseDelay:   time.Millisecond,
	})

	body, err := client.Get(context.Background(), mustURL(t, "https://javdb.com/v/x"))
	require.NoError(t, err)
	assert.Equal(t, "<html>ok</html>", string(body))
	assert.Equal(t, 2, attempts)
}

func TestTrawlTransportBodyIsDrainedAndClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := trawlResponse{
			Status: "ok",
			Solution: trawlSolution{
				URL:      "https://javdb.com/v/x",
				Status:   200,
				Response: "<html>drain test</html>",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, Config{
		TrawlURL:         server.URL,
		RateLimitDisabled: true,
		RetryDisabled:    true,
	})

	body, err := client.Get(context.Background(), mustURL(t, "https://javdb.com/v/x"))
	require.NoError(t, err)
	assert.Equal(t, "<html>drain test</html>", string(body))
}

func TestTrawlTransportWithoutTrawlDoesNotIntercept(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "direct")
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, Config{
		TrawlURL:         "",
		RateLimitDisabled: true,
		RetryDisabled:    true,
	})

	body, err := client.Get(context.Background(), mustURL(t, server.URL))
	require.NoError(t, err)
	assert.Equal(t, "direct", string(body))
}
