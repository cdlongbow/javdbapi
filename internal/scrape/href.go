package scrape

import (
	"fmt"
	"net/url"
	"strings"
)

// refIDFromHref extracts the trailing ID segment from a site-relative path
// such as "/actors/vd5z", enforcing that it is exactly "/{routeSegment}/{id}".
func refIDFromHref(href, routeSegment string) (string, error) {
	href = strings.TrimSpace(href)
	u, err := url.Parse(href)
	if err != nil {
		return "", fmt.Errorf("%w: invalid %s href %q: %w", ErrParse, routeSegment, href, err)
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(segments) != 2 || segments[0] != routeSegment || segments[1] == "" {
		return "", fmt.Errorf("%w: unexpected %s path %q", ErrParse, routeSegment, href)
	}
	return segments[1], nil
}
