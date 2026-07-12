package scrape

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var leadingIntPattern = regexp.MustCompile(`[0-9]+`)

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// parseCount extracts the first integer run from free-form label text,
// e.g. "1,234人想看" or "56 人看過".
func parseCount(raw string) (int, error) {
	cleaned := strings.ReplaceAll(raw, ",", "")
	m := leadingIntPattern.FindString(cleaned)
	if m == "" {
		return 0, fmt.Errorf("%w: no count in %q", ErrParse, raw)
	}
	n, err := strconv.Atoi(m)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid count %q: %w", ErrParse, raw, err)
	}
	return n, nil
}
