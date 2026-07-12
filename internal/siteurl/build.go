package siteurl

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const reviewsRoute = "reviews/lastest"

var segmentPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func validateSegment(kind, value string) error {
	if !segmentPattern.MatchString(value) {
		return fmt.Errorf("invalid %s %q: must contain only letters and digits", kind, value)
	}
	return nil
}

func cloneBase(base *url.URL) (*url.URL, error) {
	if base == nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid base url")
	}
	clone := *base
	return &clone, nil
}

// appendSegments joins pre-validated, alphanumeric-only path segments onto
// the base path. It never accepts a caller-supplied raw path or query string.
func appendSegments(u *url.URL, segments ...string) {
	escaped := strings.TrimRight(u.EscapedPath(), "/")
	for _, s := range segments {
		if s == "" {
			continue
		}
		escaped += "/" + s
	}
	u.RawPath = escaped
	if unescaped, err := url.PathUnescape(escaped); err == nil {
		u.Path = unescaped
	}
}

func setQuery(u *url.URL, values url.Values) {
	q := u.Query()
	for k, vs := range values {
		for _, v := range vs {
			if v == "" {
				continue
			}
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
}

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func Video(base *url.URL, id, locale string) (*url.URL, error) {
	if err := validateSegment("video id", id); err != nil {
		return nil, err
	}
	u, err := cloneBase(base)
	if err != nil {
		return nil, err
	}
	appendSegments(u, "v", id)
	setQuery(u, url.Values{"locale": {locale}})
	return u, nil
}

// Reviews keeps the site's real "reviews/lastest" route spelling internal;
// it is not part of the SDK's public API.
func Reviews(base *url.URL, id, locale string) (*url.URL, error) {
	if err := validateSegment("video id", id); err != nil {
		return nil, err
	}
	u, err := cloneBase(base)
	if err != nil {
		return nil, err
	}
	appendSegments(u, "v", id, reviewsRoute)
	setQuery(u, url.Values{"locale": {locale}})
	return u, nil
}

func Search(base *url.URL, keyword string, page int, locale string) (*url.URL, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, fmt.Errorf("missing search keyword")
	}
	u, err := cloneBase(base)
	if err != nil {
		return nil, err
	}
	appendSegments(u, "search")
	setQuery(u, url.Values{
		"q":      {keyword},
		"f":      {"all"},
		"page":   {strconv.Itoa(normalizePage(page))},
		"locale": {locale},
	})
	return u, nil
}

func Home(base *url.URL, typeSegment, filter, sort string, page int, locale string) (*url.URL, error) {
	u, err := cloneBase(base)
	if err != nil {
		return nil, err
	}
	if typeSegment != "" {
		if err := validateSegment("home type", typeSegment); err != nil {
			return nil, err
		}
		appendSegments(u, typeSegment)
	}
	setQuery(u, url.Values{
		"vft":    {filter},
		"vst":    {sort},
		"page":   {strconv.Itoa(normalizePage(page))},
		"locale": {locale},
	})
	return u, nil
}

func MakerVideos(base *url.URL, makerID, filter string, sortType, page int, locale string) (*url.URL, error) {
	if err := validateSegment("maker id", makerID); err != nil {
		return nil, err
	}
	u, err := cloneBase(base)
	if err != nil {
		return nil, err
	}
	appendSegments(u, "makers", makerID)
	setQuery(u, url.Values{
		"f":         {filter},
		"sort_type": {strconv.Itoa(sortType)},
		"page":      {strconv.Itoa(normalizePage(page))},
		"locale":    {locale},
	})
	return u, nil
}

func ActorVideos(base *url.URL, actorID string, filters []string, sortType, page int, locale string) (*url.URL, error) {
	if err := validateSegment("actor id", actorID); err != nil {
		return nil, err
	}
	u, err := cloneBase(base)
	if err != nil {
		return nil, err
	}
	appendSegments(u, "actors", actorID)

	values := url.Values{
		"sort_type": {strconv.Itoa(sortType)},
		"page":      {strconv.Itoa(normalizePage(page))},
		"locale":    {locale},
	}
	nonEmpty := make([]string, 0, len(filters))
	for _, f := range filters {
		if f != "" {
			nonEmpty = append(nonEmpty, f)
		}
	}
	if len(nonEmpty) > 0 {
		values["t"] = []string{strings.Join(nonEmpty, ",")}
	}
	setQuery(u, values)
	return u, nil
}

func Ranking(base *url.URL, period, rankingType string, page int, locale string) (*url.URL, error) {
	u, err := cloneBase(base)
	if err != nil {
		return nil, err
	}
	appendSegments(u, "rankings", "movies")
	setQuery(u, url.Values{
		"p":      {period},
		"t":      {rankingType},
		"page":   {strconv.Itoa(normalizePage(page))},
		"locale": {locale},
	})
	return u, nil
}

// SameHostRedirect mirrors http.Client.CheckRedirect's (req, via) shape but
// takes the upcoming request's URL directly so it can be unit tested without
// constructing a full *http.Request for the redirect target.
func SameHostRedirect(next *url.URL, via []*http.Request) error {
	if len(via) == 0 || next == nil {
		return nil
	}
	original := via[0].URL
	if original == nil || !strings.EqualFold(next.Host, original.Host) {
		return fmt.Errorf("refusing cross-host redirect to %q", next.Host)
	}
	return nil
}
