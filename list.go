package javdbapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/RPbro/javdbapi/internal/scrape"
	"github.com/RPbro/javdbapi/internal/siteurl"
)

func (c *Client) Home(ctx context.Context, query HomeQuery) (Page[VideoSummary], error) {
	if !query.Type.valid() {
		return Page[VideoSummary]{}, fmt.Errorf("%w: invalid home type %q", ErrInvalidQuery, query.Type)
	}
	if !query.Filter.valid() {
		return Page[VideoSummary]{}, fmt.Errorf("%w: invalid home filter %q", ErrInvalidQuery, query.Filter)
	}
	if !query.Sort.valid() {
		return Page[VideoSummary]{}, fmt.Errorf("%w: invalid home sort %q", ErrInvalidQuery, query.Sort)
	}
	page, err := normalizeQueryPage(query.Page)
	if err != nil {
		return Page[VideoSummary]{}, err
	}

	target, err := siteurl.Home(c.baseURL, string(query.Type), string(query.Filter), string(query.Sort), page, string(c.locale))
	if err != nil {
		return Page[VideoSummary]{}, &OpError{Op: "home.build_url", Err: err}
	}
	return c.fetchList(ctx, "home", target, page)
}

func (c *Client) Search(ctx context.Context, query SearchQuery) (Page[VideoSummary], error) {
	if strings.TrimSpace(query.Keyword) == "" {
		return Page[VideoSummary]{}, fmt.Errorf("%w: search keyword must not be empty", ErrInvalidQuery)
	}
	page, err := normalizeQueryPage(query.Page)
	if err != nil {
		return Page[VideoSummary]{}, err
	}

	target, err := siteurl.Search(c.baseURL, query.Keyword, page, string(c.locale))
	if err != nil {
		return Page[VideoSummary]{}, &OpError{Op: "search.build_url", Err: err}
	}
	return c.fetchList(ctx, "search", target, page)
}

// SearchActors searches for actors by keyword via the site's actor filter.
func (c *Client) SearchActors(ctx context.Context, keyword string, page int) (Page[ActorSearchItem], error) {
	if strings.TrimSpace(keyword) == "" {
		return Page[ActorSearchItem]{}, fmt.Errorf("%w: search keyword must not be empty", ErrInvalidQuery)
	}

	target, err := siteurl.SearchActors(c.baseURL, keyword, page, string(c.locale))
	if err != nil {
		return Page[ActorSearchItem]{}, &OpError{Op: "search_actors.build_url", Err: err}
	}

	body, err := c.fetcher.Get(ctx, target)
	if err != nil {
		return Page[ActorSearchItem]{}, mapFetchErr("search_actors.fetch", target, err)
	}

	doc, err := newDocument(body)
	if err != nil {
		return Page[ActorSearchItem]{}, &OpError{Op: "search_actors.document", URL: sanitizedURL(target), Err: fmt.Errorf("%w: %v", ErrParse, err)}
	}

	parsed, err := scrape.ParseActorSearch(doc, 1)
	if err != nil {
		return Page[ActorSearchItem]{}, mapScrapeErr("search_actors.parse", target, err)
	}

	items := make([]ActorSearchItem, len(parsed.Value.Items))
	for i, item := range parsed.Value.Items {
		items[i] = ActorSearchItem{
			ID:      ActorID(item.ID),
			Name:    item.Name,
			Aliases: item.Aliases,
		}
	}
	return NewPage(page, items, parsed.Value.HasNext), nil
}

// ActorByName searches for an actor directly via the site's actor search tab
// (f=actor), extracts the actor ID from the result link, and returns it.
func (c *Client) ActorByName(ctx context.Context, name string) (ActorID, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("%w: actor name is required", ErrInvalidQuery)
	}

	normName := NormalizeName(name)

	// Try with normalized name first
	page, err := c.SearchActors(ctx, normName, 1)
	if err != nil {
		// If normalized search fails and differs from original, try original
		if normName != name {
			page, err = c.SearchActors(ctx, name, 1)
		}
		if err != nil {
			return "", err
		}
	}

	if len(page.Items) == 0 {
		return "", fmt.Errorf("%w: no actor found for %q", ErrNotFound, name)
	}

	// Pick the best match: exact > alias exact > contains > contains alias > char match
	for _, item := range page.Items {
		if item.Name == name || item.Name == normName {
			return item.ID, nil
		}
		for _, alias := range item.Aliases {
			if alias == name || alias == normName {
				return item.ID, nil
			}
		}
	}
	for _, item := range page.Items {
		if strings.Contains(item.Name, name) || strings.Contains(item.Name, normName) {
			return item.ID, nil
		}
		for _, alias := range item.Aliases {
			if strings.Contains(alias, name) || strings.Contains(alias, normName) {
				return item.ID, nil
			}
		}
	}
	for _, item := range page.Items {
		if charsOverlap(item.Name, name, normName) {
			return item.ID, nil
		}
		for _, alias := range item.Aliases {
			if charsOverlap(alias, name, normName) {
				return item.ID, nil
			}
		}
	}
	return page.Items[0].ID, nil
}

// charsOverlap reports whether a and b (or a and c) share at least one character.
func charsOverlap(a, b, c string) bool {
	set := make(map[rune]struct{})
	for _, r := range a {
		set[r] = struct{}{}
	}
	for _, r := range b {
		if _, ok := set[r]; ok {
			return true
		}
	}
	for _, r := range c {
		if _, ok := set[r]; ok {
			return true
		}
	}
	return false
}

func (c *Client) MakerVideos(ctx context.Context, query MakerVideosQuery) (Page[VideoSummary], error) {
	if !query.Filter.valid() {
		return Page[VideoSummary]{}, fmt.Errorf("%w: invalid maker filter %q", ErrInvalidQuery, query.Filter)
	}
	sortType, err := query.Sort.sortTypeValue()
	if err != nil {
		return Page[VideoSummary]{}, err
	}
	page, err := normalizeQueryPage(query.Page)
	if err != nil {
		return Page[VideoSummary]{}, err
	}

	target, err := siteurl.MakerVideos(c.baseURL, string(query.MakerID), string(query.Filter), sortType, page, string(c.locale))
	if err != nil {
		return Page[VideoSummary]{}, &OpError{Op: "maker_videos.build_url", Err: err}
	}
	return c.fetchList(ctx, "maker_videos", target, page)
}

func (c *Client) ActorVideos(ctx context.Context, query ActorVideosQuery) (Page[VideoSummary], error) {
	filters := make([]string, 0, len(query.Filters))
	for _, f := range query.Filters {
		if !f.valid() {
			return Page[VideoSummary]{}, fmt.Errorf("%w: invalid actor filter %q", ErrInvalidQuery, f)
		}
		filters = append(filters, string(f))
	}
	sortType, err := query.Sort.sortTypeValue()
	if err != nil {
		return Page[VideoSummary]{}, err
	}
	page, err := normalizeQueryPage(query.Page)
	if err != nil {
		return Page[VideoSummary]{}, err
	}

	target, err := siteurl.ActorVideos(c.baseURL, string(query.ActorID), filters, sortType, page, string(c.locale))
	if err != nil {
		return Page[VideoSummary]{}, &OpError{Op: "actor_videos.build_url", Err: err}
	}
	return c.fetchList(ctx, "actor_videos", target, page)
}

func (c *Client) Ranking(ctx context.Context, query RankingQuery) (Page[VideoSummary], error) {
	if !query.Period.valid() {
		return Page[VideoSummary]{}, fmt.Errorf("%w: invalid ranking period %q", ErrInvalidQuery, query.Period)
	}
	if !query.Type.valid() {
		return Page[VideoSummary]{}, fmt.Errorf("%w: invalid ranking type %q", ErrInvalidQuery, query.Type)
	}
	page, err := normalizeQueryPage(query.Page)
	if err != nil {
		return Page[VideoSummary]{}, err
	}

	target, err := siteurl.Ranking(c.baseURL, string(query.Period), string(query.Type), page, string(c.locale))
	if err != nil {
		return Page[VideoSummary]{}, &OpError{Op: "ranking.build_url", Err: err}
	}
	return c.fetchList(ctx, "ranking", target, page)
}
