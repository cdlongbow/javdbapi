package clioutput

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	javdbapi "github.com/RPbro/javdbapi"
)

func fixedNow() time.Time {
	return time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
}

func TestLoadTreatsV1AsMiss(t *testing.T) {
	store := NewStore(t.TempDir(), fixedNow)
	require.NoError(t, os.WriteFile(filepath.Join(store.outputDir, "P9Jkq9.json"), []byte(`{"metadata":{}}`), 0o644))
	state, err := store.Load("P9Jkq9", 24*time.Hour)
	require.NoError(t, err)
	assert.False(t, state.DetailFresh)
	assert.Nil(t, state.Document)
}

func TestFailedReviewsRemainRetryable(t *testing.T) {
	state := CacheState{DetailFresh: true, ReviewsFresh: false}
	assert.True(t, state.NeedsReviews())
}

func TestLoadReportsMissingFile(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir(), fixedNow)
	state, err := store.Load("P9Jkq9", 24*time.Hour)
	require.NoError(t, err)
	assert.Nil(t, state.Document)
	assert.False(t, state.DetailFresh)
	assert.False(t, state.ReviewsFresh)
}

func TestWriteFileThenLoadReportsIndependentFreshness(t *testing.T) {
	t.Parallel()

	now := fixedNow()
	store := NewStore(t.TempDir(), func() time.Time { return now })

	reviewsUpdatedAt := now.Add(-72 * time.Hour)
	doc := Document{
		SchemaVersion: SchemaVersion,
		Metadata: Metadata{
			LastUpdated:      now,
			DetailUpdatedAt:  now,
			ReviewsUpdatedAt: &reviewsUpdatedAt,
			Sources:          []Source{NewSearchSource("VR", 1)},
		},
		Detail: javdbapi.NewVideoDetail(javdbapi.VideoSummary{ID: "P9Jkq9", Code: "ABC-123"}),
	}
	require.NoError(t, store.WriteFile(doc))

	state, err := store.Load("P9Jkq9", 24*time.Hour)
	require.NoError(t, err)
	require.NotNil(t, state.Document)
	assert.True(t, state.DetailFresh)
	assert.False(t, state.ReviewsFresh, "reviews were updated 72h ago, past the 24h staleAfter window")
	assert.False(t, state.NeedsDetail())
	assert.True(t, state.NeedsReviews())
}

func TestWriteFileMergesSourcesWithoutDuplicates(t *testing.T) {
	t.Parallel()

	now := fixedNow()
	store := NewStore(t.TempDir(), func() time.Time { return now })

	base := javdbapi.NewVideoDetail(javdbapi.VideoSummary{ID: "P9Jkq9", Code: "ABC-123"})

	first := Document{
		SchemaVersion: SchemaVersion,
		Metadata: Metadata{
			LastUpdated:     now.Add(-48 * time.Hour),
			DetailUpdatedAt: now.Add(-48 * time.Hour),
			Sources:         []Source{NewSearchSource("VR", 1)},
		},
		Detail: base,
	}
	require.NoError(t, store.WriteFile(first))

	second := Document{
		SchemaVersion: SchemaVersion,
		Metadata: Metadata{
			LastUpdated:     now,
			DetailUpdatedAt: now,
			Sources: []Source{
				NewSearchSource("VR", 1),
				NewVideoSource("P9Jkq9"),
			},
		},
		Detail: base,
	}
	require.NoError(t, store.WriteFile(second))

	loaded, err := store.Load("P9Jkq9", 24*time.Hour)
	require.NoError(t, err)
	require.NotNil(t, loaded.Document)
	assert.Len(t, loaded.Document.Metadata.Sources, 2)
}

func TestWriteFileSerializesEmptySlicesNotNull(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir(), fixedNow)
	doc := Document{
		SchemaVersion: SchemaVersion,
		Metadata:      Metadata{LastUpdated: fixedNow(), DetailUpdatedAt: fixedNow()},
		Detail:        javdbapi.NewVideoDetail(javdbapi.VideoSummary{ID: "P9Jkq9"}),
	}
	require.NoError(t, store.WriteFile(doc))

	filePath, err := store.FilePath("P9Jkq9")
	require.NoError(t, err)
	body, err := os.ReadFile(filePath)
	require.NoError(t, err)

	assert.Contains(t, string(body), `"reviews": []`)
	assert.Contains(t, string(body), `"partial_errors": []`)
	assert.NotContains(t, string(body), "null")
}

func TestWriteFileAtomicRenameFailureLeavesPreviousDocumentIntact(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := NewStore(dir, fixedNow)

	first := Document{
		SchemaVersion: SchemaVersion,
		Metadata:      Metadata{LastUpdated: fixedNow(), DetailUpdatedAt: fixedNow()},
		Detail:        javdbapi.NewVideoDetail(javdbapi.VideoSummary{ID: "P9Jkq9", Code: "FIRST"}),
	}
	require.NoError(t, store.WriteFile(first))

	store.rename = func(oldpath, newpath string) error { return errors.New("simulated rename failure") }

	second := Document{
		SchemaVersion: SchemaVersion,
		Metadata:      Metadata{LastUpdated: fixedNow(), DetailUpdatedAt: fixedNow()},
		Detail:        javdbapi.NewVideoDetail(javdbapi.VideoSummary{ID: "P9Jkq9", Code: "SECOND"}),
	}
	err := store.WriteFile(second)
	require.Error(t, err)

	state, err := store.Load("P9Jkq9", 24*time.Hour)
	require.NoError(t, err)
	require.NotNil(t, state.Document)
	assert.Equal(t, "FIRST", state.Document.Detail.Summary.Code)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, strings.HasPrefix(entry.Name(), ".tmp-"), "temp file %s was not cleaned up", entry.Name())
	}
}
