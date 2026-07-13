package javdbapi_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	javdbapi "github.com/RPbro/javdbapi"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("internal", "scrape", "testdata", name))
	require.NoError(t, err)
	return data
}

// newFixtureServer serves the given fixture body for every request and
// records each requested path (without query) in request order.
func newFixtureServer(t *testing.T, body []byte, onRequest func(path string)) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		onRequest(r.URL.Path)
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server
}

func newDetailFixtureServer(t *testing.T, onRequest func(path string)) *httptest.Server {
	t.Helper()
	return newFixtureServer(t, readFixture(t, "detail-complete.html"), onRequest)
}

func newReviewsFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	return newFixtureServer(t, readFixture(t, "reviews.html"), func(string) {})
}

func newClientForServer(t *testing.T, server *httptest.Server, opts ...func(*javdbapi.ClientConfig)) *javdbapi.Client {
	t.Helper()
	cfg := javdbapi.ClientConfig{
		BaseURL:   server.URL,
		Retry:     javdbapi.RetryPolicy{Disabled: true},
		RateLimit: javdbapi.RateLimitPolicy{Disabled: true},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	client, err := javdbapi.NewClient(cfg)
	require.NoError(t, err)
	return client
}

func TestClientDetailDoesNotFetchReviews(t *testing.T) {
	var paths []string
	server := newDetailFixtureServer(t, func(path string) { paths = append(paths, path) })
	client := newClientForServer(t, server)
	detail, err := client.Detail(context.Background(), "P9Jkq9")
	require.NoError(t, err)
	assert.Equal(t, javdbapi.VideoID("P9Jkq9"), detail.Summary.ID)
	assert.Equal(t, []string{"/v/P9Jkq9"}, paths)
}

func TestClientReviewsAreIndependent(t *testing.T) {
	client := newClientForServer(t, newReviewsFixtureServer(t))
	reviews, err := client.Reviews(context.Background(), "P9Jkq9")
	require.NoError(t, err)
	require.NotEmpty(t, reviews)
}

func TestClientHomeMapsQueryAndHasNext(t *testing.T) {
	var captured *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := *r.URL
		captured = &u
		_, _ = w.Write(readFixture(t, "list-pagination.html"))
	}))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server)
	page, err := client.Home(context.Background(), javdbapi.HomeQuery{
		Type:   javdbapi.HomeTypeCensored,
		Filter: javdbapi.HomeFilterDownload,
		Sort:   javdbapi.HomeSortMagnetDate,
		Page:   2,
	})
	require.NoError(t, err)
	require.NotNil(t, captured)

	assert.Equal(t, "/censored", captured.Path)
	q := captured.Query()
	assert.Equal(t, "1", q.Get("vft"))
	assert.Equal(t, "2", q.Get("vst"))
	assert.Equal(t, "2", q.Get("page"))
	assert.Equal(t, "zh", q.Get("locale"))
	assert.True(t, page.HasNext)
	assert.Equal(t, 2, page.Number)
	assert.NotEmpty(t, page.Items)
}

func TestClientSearchUsesFixedLocaleAndCustomUserAgent(t *testing.T) {
	var gotUA, gotLocale string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotLocale = r.URL.Query().Get("locale")
		_, _ = w.Write(readFixture(t, "list.html"))
	}))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server, func(cfg *javdbapi.ClientConfig) {
		cfg.UserAgent = "javdbapi-test-agent/1.0"
	})
	_, err := client.Search(context.Background(), javdbapi.SearchQuery{Keyword: "VR", Page: 1})
	require.NoError(t, err)
	assert.Equal(t, "javdbapi-test-agent/1.0", gotUA)
	assert.Equal(t, string(javdbapi.LocaleZH), gotLocale)
}

func TestClientMapsEmptyResult(t *testing.T) {
	client := newClientForServer(t, newFixtureServer(t, readFixture(t, "list-empty.html"), func(string) {}))
	_, err := client.Search(context.Background(), javdbapi.SearchQuery{Keyword: "nonexistent", Page: 1})
	require.ErrorIs(t, err, javdbapi.ErrEmptyResult)
}

func TestClientMapsNotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server)
	_, err := client.Detail(context.Background(), "P9Jkq9")
	require.ErrorIs(t, err, javdbapi.ErrNotFound)

	var httpErr *javdbapi.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusNotFound, httpErr.StatusCode)
}

func TestClientMapsRateLimitedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server)
	_, err := client.Detail(context.Background(), "P9Jkq9")
	require.ErrorIs(t, err, javdbapi.ErrRateLimited)
}

func TestClientErrorsDoNotLeakSearchKeyword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server)
	_, err := client.Search(context.Background(), javdbapi.SearchQuery{Keyword: "top-secret-query", Page: 1})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "top-secret-query")
}

func TestClientLogsParseWarningsWithoutRawHTML(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	server := newDetailFixtureServer(t, func(string) {})
	client := newClientForServer(t, server, func(cfg *javdbapi.ClientConfig) {
		cfg.Logger = logger
	})

	_, err := client.Detail(context.Background(), "P9Jkq9")
	require.NoError(t, err)

	logged := logBuf.String()
	assert.Contains(t, logged, "magnets.size")
	assert.NotContains(t, logged, "<html")
	assert.NotContains(t, logged, "<!DOCTYPE")
}

func TestClientVideoURL(t *testing.T) {
	client, err := javdbapi.NewClient(javdbapi.ClientConfig{
		Retry:     javdbapi.RetryPolicy{Disabled: true},
		RateLimit: javdbapi.RateLimitPolicy{Disabled: true},
	})
	require.NoError(t, err)
	assert.Equal(t, "https://javdb573.com/v/P9Jkq9?locale=zh", client.VideoURL("P9Jkq9"))
}

func TestClientRankingBuildsFixedRoute(t *testing.T) {
	var captured *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := *r.URL
		captured = &u
		_, _ = w.Write(readFixture(t, "list.html"))
	}))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server)
	_, err := client.Ranking(context.Background(), javdbapi.RankingQuery{
		Period: javdbapi.RankingPeriodWeekly,
		Type:   javdbapi.RankingTypeCensored,
		Page:   1,
	})
	require.NoError(t, err)
	require.NotNil(t, captured)
	assert.Equal(t, "/rankings/movies", captured.Path)
	assert.Equal(t, "weekly", captured.Query().Get("p"))
	assert.Equal(t, "censored", captured.Query().Get("t"))
}

func TestClientActorVideosJoinsFilters(t *testing.T) {
	var captured *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := *r.URL
		captured = &u
		_, _ = w.Write(readFixture(t, "list.html"))
	}))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server)
	_, err := client.ActorVideos(context.Background(), javdbapi.ActorVideosQuery{
		ActorID: "vd5z",
		Filters: []javdbapi.ActorFilter{javdbapi.ActorFilterPlayable, javdbapi.ActorFilterDownload},
		Page:    1,
	})
	require.NoError(t, err)
	require.NotNil(t, captured)
	assert.Equal(t, "/actors/vd5z", captured.Path)
	assert.Equal(t, "p,d", captured.Query().Get("t"))
	assert.Equal(t, strconv.Itoa(0), captured.Query().Get("sort_type"))
}

func TestClientMakerVideosBuildsSortType(t *testing.T) {
	var captured *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := *r.URL
		captured = &u
		_, _ = w.Write(readFixture(t, "list.html"))
	}))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server)
	_, err := client.MakerVideos(context.Background(), javdbapi.MakerVideosQuery{
		MakerID: "ybW",
		Filter:  javdbapi.MakerFilterPlayable,
		Sort:    javdbapi.ListSortRating,
		Page:    1,
	})
	require.NoError(t, err)
	require.NotNil(t, captured)
	assert.Equal(t, "/makers/ybW", captured.Path)
	assert.Equal(t, "playable", captured.Query().Get("f"))
	assert.Equal(t, "1", captured.Query().Get("sort_type"))
}

func TestClientRejectsInvalidQueryWithoutRequest(t *testing.T) {
	var requested bool
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requested = true }))
	t.Cleanup(server.Close)

	client := newClientForServer(t, server)
	_, err := client.Search(context.Background(), javdbapi.SearchQuery{Keyword: "", Page: 1})
	require.Error(t, err)
	assert.False(t, requested)
}

func TestClientSearchRejectsEmptyKeyword(t *testing.T) {
	for name, keyword := range map[string]string{
		"empty":           "",
		"whitespace-only": "   ",
	} {
		t.Run(name, func(t *testing.T) {
			var requested bool
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requested = true }))
			t.Cleanup(server.Close)

			client := newClientForServer(t, server)
			_, err := client.Search(context.Background(), javdbapi.SearchQuery{Keyword: keyword, Page: 1})
			require.ErrorIs(t, err, javdbapi.ErrInvalidQuery)
			assert.False(t, requested)
		})
	}
}

func TestNewClientRejectsUnsupportedBaseURLScheme(t *testing.T) {
	for name, baseURL := range map[string]string{
		"ftp scheme":  "ftp://javdb.com",
		"no scheme":   "javdb.com",
		"file scheme": "file:///etc/passwd",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := javdbapi.NewClient(javdbapi.ClientConfig{BaseURL: baseURL})
			require.ErrorIs(t, err, javdbapi.ErrInvalidConfig)
		})
	}
}

func TestNewClientAcceptsHTTPAndHTTPSBaseURL(t *testing.T) {
	for _, baseURL := range []string{"http://javdb.com", "https://javdb.com"} {
		t.Run(baseURL, func(t *testing.T) {
			_, err := javdbapi.NewClient(javdbapi.ClientConfig{BaseURL: baseURL})
			require.NoError(t, err)
		})
	}
}
