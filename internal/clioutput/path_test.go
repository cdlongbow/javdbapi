package clioutput

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilePathBuildsFromValidatedID(t *testing.T) {
	t.Parallel()

	got, err := FilePath("/tmp/output", "P9Jkq9")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/tmp/output", "P9Jkq9.json"), got)
}

func TestFilePathRejectsHostileID(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"", "../admin", "/v/x", "a/b", "a?b=1", "a b"} {
		_, err := FilePath("/tmp/output", id)
		require.Error(t, err, "expected error for id %q", id)
	}
}

func TestSourceBuildersMarshalStableQueries(t *testing.T) {
	t.Parallel()

	actor := NewActorSource("neRNX", []string{"d", "", "c", "d"}, 2)
	raw, err := json.Marshal(actor)
	require.NoError(t, err)
	assert.Equal(
		t,
		`{"command":"actor","query":{"id":"neRNX","filter":["c","d"],"page":2}}`,
		string(raw),
	)

	video := NewVideoSource("P9Jkq9")
	raw, err = json.Marshal(video)
	require.NoError(t, err)
	assert.Equal(
		t,
		`{"command":"video","query":{"id":"P9Jkq9"}}`,
		string(raw),
	)
}

func TestNewActorSourceSerializesNormalizedRequestValues(t *testing.T) {
	t.Parallel()

	actor := NewActorSource("neRNX", []string{"d", "", "c", "d"}, 2)
	raw, err := json.Marshal(actor)
	require.NoError(t, err)
	assert.Equal(
		t,
		`{"command":"actor","query":{"id":"neRNX","filter":["c","d"],"page":2}}`,
		string(raw),
	)
}

func TestNewHomeSourceOmitsDefaultRequestValues(t *testing.T) {
	t.Parallel()

	home := NewHomeSource("", "", "", 1)
	raw, err := json.Marshal(home)
	require.NoError(t, err)
	assert.Equal(
		t,
		`{"command":"home","query":{"page":1}}`,
		string(raw),
	)
}
