package scrape

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ActorSearchItem is a single actor result from an actor-filtered search.
type ActorSearchItem struct {
	ID      string
	Name    string
	Aliases []string
}

// ActorSearchPage holds the parsed results of an actor search.
type ActorSearchPage struct {
	Items   []ActorSearchItem
	Number  int
	HasNext bool
}

var categoryActorPaths = map[string]bool{
	"/actors/censored":   true,
	"/actors/uncensored": true,
	"/actors/western":    true,
}

var actorIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// actorIDFromPath extracts the actor ID from a path like "/actors/pRMq".
func actorIDFromPath(href string) (string, error) {
	segments := strings.Split(strings.Trim(strings.TrimSpace(href), "/"), "/")
	if len(segments) != 2 || segments[0] != "actors" {
		return "", fmt.Errorf("%w: unexpected actor path %q", ErrParse, href)
	}
	id := segments[1]
	if !actorIDPattern.MatchString(id) {
		return "", fmt.Errorf("%w: invalid actor id %q in path %q", ErrParse, id, href)
	}
	return id, nil
}

// ParseActorSearch parses an actor-filtered search result page.
// It finds all actor links (<a href="/actors/{id}">) that are not category
// navigation links (/actors/censored, /actors/uncensored, /actors/western).
func ParseActorSearch(doc *goquery.Document, page int) (Result[ActorSearchPage], error) {
	result := Result[ActorSearchPage]{
		Value: ActorSearchPage{Items: []ActorSearchItem{}, Number: page},
	}

	var items []ActorSearchItem
	doc.Find("a[href*='/actors/']").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		if categoryActorPaths[href] {
			return
		}
		id, err := actorIDFromPath(href)
		if err != nil {
			return
		}

		name := strings.TrimSpace(s.Find("strong").Text())
		if name == "" {
			return
		}

		var aliases []string
		if title, ok := s.Attr("title"); ok {
			for _, p := range strings.Split(title, ",") {
				p = strings.TrimSpace(p)
				if p != "" && p != name {
					aliases = append(aliases, p)
				}
			}
		}

		items = append(items, ActorSearchItem{
			ID:      id,
			Name:    name,
			Aliases: aliases,
		})
	})

	if len(items) == 0 {
		return result, fmt.Errorf("%w: no actors found in search results", ErrEmpty)
	}

	result.Value.Items = items
	result.Value.HasNext = doc.Find("a.pagination-next[rel=next]").Length() > 0
	return result, nil
}
