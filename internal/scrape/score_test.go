package scrape

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseScore(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		value float64
		count int
	}{
		{raw: "4.55分, 由3080人評價", value: 4.55, count: 3080},
		{raw: "0.0分, 由2人評價", value: 0, count: 2},
		{raw: " 5分, 由1人評價 ", value: 5, count: 1},
	} {
		got, err := parseScore(tc.raw)
		require.NoError(t, err)
		assert.InDelta(t, tc.value, got.Value, 0.0001)
		assert.Equal(t, tc.count, got.Count)
	}
}

func TestParseScoreRejectsMalformed(t *testing.T) {
	for _, raw := range []string{"", "not a score", "4.55分"} {
		_, err := parseScore(raw)
		require.Error(t, err)
	}
}
