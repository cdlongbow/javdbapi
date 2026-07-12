package javdbapi

import (
	"fmt"
	"regexp"
	"strings"
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
	return VideoID(value), err
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
