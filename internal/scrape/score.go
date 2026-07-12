package scrape

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var scorePattern = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*分\s*,\s*由\s*([0-9]+)\s*人評價`)

func parseScore(raw string) (Score, error) {
	m := scorePattern.FindStringSubmatch(strings.TrimSpace(raw))
	if m == nil {
		return Score{}, fmt.Errorf("%w: invalid score %q", ErrParse, raw)
	}

	value, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return Score{}, fmt.Errorf("%w: invalid score value %q: %w", ErrParse, m[1], err)
	}
	count, err := strconv.Atoi(m[2])
	if err != nil {
		return Score{}, fmt.Errorf("%w: invalid score count %q: %w", ErrParse, m[2], err)
	}

	return Score{Value: value, Count: count}, nil
}
