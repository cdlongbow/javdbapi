package scrape

import (
	"errors"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReviews(t *testing.T) {
	got, err := ParseReviews(loadFixture(t, "reviews.html"))
	require.NoError(t, err)
	require.Len(t, got.Value, 1)
	assert.Equal(t, "218256471", got.Value[0].ID)
	assert.Equal(t, "wo***2", got.Value[0].Author)
	require.NotNil(t, got.Value[0].Score)
	assert.Equal(t, 4, *got.Value[0].Score, "score must count non-gray i.icon-star only")
	assert.Equal(t, 186, got.Value[0].Likes)
	require.NotNil(t, got.Value[0].PublishedAt)
	assert.Equal(t, "2026-03-10", got.Value[0].PublishedAt.Format("2006-01-02"))
	assert.Equal(t, "Great video, highly recommended.", got.Value[0].Content)
}

func TestParseReviewsEmpty(t *testing.T) {
	got, err := ParseReviews(loadFixture(t, "reviews-empty.html"))
	require.NoError(t, err)
	assert.Empty(t, got.Value)
	assert.NotNil(t, got.Value)
}

func TestParseReviewsRejectsUnrecognizedDocument(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body><div class=\"unrelated\"></div></body></html>"))
	require.NoError(t, err)
	_, err = ParseReviews(doc)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrParse))
}
