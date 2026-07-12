package scrape

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	sizePattern      = regexp.MustCompile(`(?i)([0-9]+(?:\.[0-9]+)?)\s*(KB|MB|GB|TB)`)
	sizeUnitPattern  = regexp.MustCompile(`(?i)(KB|MB|GB|TB)`)
	fileCountPattern = regexp.MustCompile(`([0-9]+)\s*個文件`)
)

var sizeMultiplier = map[string]float64{
	"KB": 1 << 10,
	"MB": 1 << 20,
	"GB": 1 << 30,
	"TB": 1 << 40,
}

type magnetMeta struct {
	SizeText  string
	SizeBytes *int64
	FileCount *int
}

// parseSize returns nil when no size unit appears in raw, and errors only
// when a size unit is present but the full "<number><unit>" shape does not
// match, so a genuinely malformed value is not silently treated as absent.
func parseSize(raw string) (*int64, error) {
	if m := sizePattern.FindStringSubmatch(raw); m != nil {
		value, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid size number %q: %w", ErrParse, m[1], err)
		}
		mult, ok := sizeMultiplier[strings.ToUpper(m[2])]
		if !ok {
			return nil, fmt.Errorf("%w: invalid size unit %q", ErrParse, m[2])
		}
		bytes := int64(math.Round(value * mult))
		return &bytes, nil
	}
	if sizeUnitPattern.MatchString(raw) {
		return nil, fmt.Errorf("%w: invalid size %q", ErrParse, raw)
	}
	return nil, nil
}

func parseFileCount(raw string) (*int, error) {
	m := fileCountPattern.FindStringSubmatch(raw)
	if m == nil {
		return nil, nil
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid file count %q: %w", ErrParse, m[1], err)
	}
	return &n, nil
}

func parseMagnetMeta(raw string) (magnetMeta, error) {
	text := normalizeSpace(raw)

	sizeBytes, err := parseSize(text)
	if err != nil {
		return magnetMeta{}, err
	}
	fileCount, err := parseFileCount(text)
	if err != nil {
		return magnetMeta{}, err
	}

	return magnetMeta{SizeText: text, SizeBytes: sizeBytes, FileCount: fileCount}, nil
}
