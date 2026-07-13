package javdbapi

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type VideoID string
type ActorID string
type MakerID string
type DirectorID string
type SeriesID string

var resourceIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func parseResourceID(kind, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if !resourceIDPattern.MatchString(raw) {
		return "", fmt.Errorf("%w: invalid %s id %q", ErrInvalidQuery, kind, raw)
	}
	return raw, nil
}

func ParseVideoID(raw string) (VideoID, error) {
	value, err := parseResourceID("video", raw)
	if err != nil {
		return "", err
	}
	return VideoID(value), nil
}

// ResolveVideoID first tries to parse the raw string as a video ID. If that
// fails (e.g. the caller supplied a code like "DLDSS-271"), it performs a
// search to look up the internal ID.
func ResolveVideoID(raw string) (VideoID, error) {
	id, err := ParseVideoID(raw)
	if err == nil {
		return id, nil
	}
	return resolveVideoIDByCode(raw)
}

func resolveVideoIDByCode(code string) (VideoID, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", fmt.Errorf("%w: empty code", ErrInvalidQuery)
	}
	client, err := NewClient(ClientConfig{})
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	page, err := client.Search(ctx, SearchQuery{Keyword: code, Page: 1})
	if err != nil {
		return "", err
	}
	if len(page.Items) == 0 {
		return "", fmt.Errorf("%w: no video found for code %q", ErrNotFound, code)
	}
	return page.Items[0].ID, nil
}

func ParseActorID(raw string) (ActorID, error) {
	value, err := parseResourceID("actor", raw)
	return ActorID(value), err
}

func ParseMakerID(raw string) (MakerID, error) {
	value, err := parseResourceID("maker", raw)
	return MakerID(value), err
}

func ParseDirectorID(raw string) (DirectorID, error) {
	value, err := parseResourceID("director", raw)
	return DirectorID(value), err
}

func ParseSeriesID(raw string) (SeriesID, error) {
	value, err := parseResourceID("series", raw)
	return SeriesID(value), err
}
