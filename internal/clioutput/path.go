package clioutput

import (
	"fmt"
	"path/filepath"
	"regexp"
)

// idPattern mirrors the SDK's own resource-ID validation (alnum only), so a
// caller-supplied ID can never escape outputDir via "/" or "..".
var idPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// FilePath returns the cache file path for a validated video ID within
// outputDir. The cache file name is the ID itself, per the v2 design: no
// path-derived key is needed since the CLI now takes --id directly.
func FilePath(outputDir, id string) (string, error) {
	if !idPattern.MatchString(id) {
		return "", fmt.Errorf("invalid video id %q", id)
	}
	return filepath.Join(outputDir, id+".json"), nil
}
