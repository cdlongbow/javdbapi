package javdbapi_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	javdbapi "github.com/RPbro/javdbapi"
)

func TestParseVideoID(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		raw     string
		want    javdbapi.VideoID
		wantErr bool
	}{
		{raw: "P9Jkq9", want: "P9Jkq9"},
		{raw: "", wantErr: true},
		{raw: "/v/P9Jkq9", wantErr: true},
		{raw: "../admin", wantErr: true},
		{raw: "abc?q=1", wantErr: true},
	} {
		got, err := javdbapi.ParseVideoID(tc.raw)
		if tc.wantErr {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
		assert.Equal(t, tc.want, got)
	}
}

func TestPageItemsEncodeAsArray(t *testing.T) {
	body, err := json.Marshal(javdbapi.Page[javdbapi.VideoSummary]{Number: 1})
	require.NoError(t, err)
	assert.JSONEq(t, `{"items":[],"number":1,"has_next":false}`, string(body))
}
