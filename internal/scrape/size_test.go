package scrape

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMagnetMeta(t *testing.T) {
	got, err := parseMagnetMeta("5.76GB, 2個文件")
	require.NoError(t, err)
	assert.Equal(t, "5.76GB, 2個文件", got.SizeText)
	require.NotNil(t, got.SizeBytes)
	assert.EqualValues(t, 6184752906, *got.SizeBytes)
	require.NotNil(t, got.FileCount)
	assert.Equal(t, 2, *got.FileCount)
}

func TestParseMagnetMetaWithoutFileCount(t *testing.T) {
	got, err := parseMagnetMeta("1.2TB")
	require.NoError(t, err)
	require.NotNil(t, got.SizeBytes)
	assert.Nil(t, got.FileCount)
}

func TestParseMagnetMetaRejectsMalformedSize(t *testing.T) {
	_, err := parseMagnetMeta("abcGB, 2個文件")
	require.Error(t, err)
}

func TestParseMagnetMetaAllowsMissingSize(t *testing.T) {
	got, err := parseMagnetMeta("2個文件")
	require.NoError(t, err)
	assert.Nil(t, got.SizeBytes)
	require.NotNil(t, got.FileCount)
	assert.Equal(t, 2, *got.FileCount)
}
