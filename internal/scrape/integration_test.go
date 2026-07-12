//go:build integration

package scrape_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	javdbapi "github.com/RPbro/javdbapi"
)

// TestLiveSearchContract is an opt-in structural contract check against the
// real site. It never runs in CI or the default test target; a 403, 429, or
// network error here is an environment/access failure, not a fixture parser
// regression, and must not be used to justify weakening fixture assertions.
func TestLiveSearchContract(t *testing.T) {
	if os.Getenv("JAVDB_INTEGRATION") != "1" {
		t.Skip("set JAVDB_INTEGRATION=1 to run live contract tests")
	}

	client, err := javdbapi.NewClient(javdbapi.ClientConfig{})
	require.NoError(t, err)

	page, err := client.Search(context.Background(), javdbapi.SearchQuery{Keyword: "VR", Page: 1})
	require.NoError(t, err)
	require.NotEmpty(t, page.Items)
	assert.NotEmpty(t, page.Items[0].ID)
	assert.NotEmpty(t, page.Items[0].Code)
}

// TestLiveDetailContract is a low-volume opt-in check that the detail parser
// still matches the live DOM for actors, magnets, and want/watched counts —
// the exact sections that broke silently under fixture-only testing.
func TestLiveDetailContract(t *testing.T) {
	if os.Getenv("JAVDB_INTEGRATION") != "1" {
		t.Skip("set JAVDB_INTEGRATION=1 to run live contract tests")
	}

	client, err := javdbapi.NewClient(javdbapi.ClientConfig{})
	require.NoError(t, err)

	id, err := javdbapi.ParseVideoID("P9Jkq9")
	require.NoError(t, err)

	detail, err := client.Detail(context.Background(), id)
	require.NoError(t, err)
	assert.NotEmpty(t, detail.Actors)
	assert.NotEmpty(t, detail.Magnets)
	assert.NotNil(t, detail.WantCount)
	assert.NotNil(t, detail.WatchedCount)
}

// TestLiveReviewsContract is a low-volume opt-in check that the reviews
// parser correctly strips report/likes controls from the author field and
// derives a score from the live DOM.
func TestLiveReviewsContract(t *testing.T) {
	if os.Getenv("JAVDB_INTEGRATION") != "1" {
		t.Skip("set JAVDB_INTEGRATION=1 to run live contract tests")
	}

	client, err := javdbapi.NewClient(javdbapi.ClientConfig{})
	require.NoError(t, err)

	id, err := javdbapi.ParseVideoID("P9Jkq9")
	require.NoError(t, err)

	reviews, err := client.Reviews(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, reviews)
	for _, r := range reviews {
		assert.NotContains(t, r.Author, "檢舉")
		assert.NotContains(t, r.Author, "讚")
		assert.NotContains(t, r.Author, "贊")
	}
	assert.NotNil(t, reviews[0].Score)
}
