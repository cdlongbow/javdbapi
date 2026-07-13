package clioutput

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// idPattern mirrors the SDK's own resource-ID validation (alnum only), so a
// caller-supplied ID can never escape outputDir via "/" or "..".
var idPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// codePattern allows alphanumeric, hyphens, and underscores for video codes
// like "DLDSS-271" or "HEY-001".
var codePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9\-_]*$`)

// FilePath returns the cache file path for a validated video ID within
// outputDir. The cache file name is the ID itself, per the v2 design: no
// path-derived key is needed since the CLI now takes --id directly.
func FilePath(outputDir, id string) (string, error) {
	if !idPattern.MatchString(id) {
		return "", fmt.Errorf("invalid video id %q", id)
	}
	return filepath.Join(outputDir, id+".json"), nil
}

// FilePathWithCode returns the cache file path, preferring the video code
// (番号) over the internal ID for the filename. Codes may contain hyphens.
func FilePathWithCode(outputDir, code, id string) (string, error) {
	name := code
	if name == "" {
		name = id
	}
	name = sanitizeFileName(name)
	return filepath.Join(outputDir, name+".json"), nil
}

// sanitizeFileName removes characters that are unsafe for filenames.
func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	if len(name) > 100 {
		name = name[:100]
	}
	return name
}
