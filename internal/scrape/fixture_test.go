package scrape

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func loadFixture(t *testing.T, name string) *goquery.Document {
	t.Helper()

	f, err := os.Open(filepath.Join("testdata", name))
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)
	return doc
}
