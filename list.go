package javdbapi

import (
	"context"
	"fmt"
	"strings"

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

// ActorByName resolves an actor ID from their name by searching and examining
// the first video's detail page. It prefers a single-work video and picks the
// closest matching female actor name.
func (c *Client) ActorByName(ctx context.Context, name string) (ActorID, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("%w: actor name is required", ErrInvalidQuery)
	}

	// Step 1: search for the name
	searchPage, err := c.Search(ctx, SearchQuery{Keyword: name, Page: 1})
	if err != nil || len(searchPage.Items) == 0 {
		return "", err
	}

	// Step 2: find a single-work video first, fallback to first video
	var candidateID VideoID
	for _, item := range searchPage.Items {
		detail, err := c.Detail(ctx, item.ID)
		if err != nil || detail == nil {
			continue
		}
		for _, tag := range detail.Tags {
			if tag.Name == "單體作品" || tag.Name == "单体作品" {
				candidateID = item.ID
				goto found
			}
		}
	}
	// Fallback: use first video
	candidateID = searchPage.Items[0].ID

found:
	detail, err := c.Detail(ctx, candidateID)
	if err != nil || detail == nil {
		return "", err
	}

	// Step 3: find the best matching female actor
	return findBestMatchingFemaleActor(detail.Actors, name)
}

// findBestMatchingFemaleActor filters female actors and picks the one whose
// name is closest to the search keyword. Exact match > contains > prefix > first.
func findBestMatchingFemaleActor(actors []Actor, keyword string) (ActorID, error) {
	var females []Actor
	for _, a := range actors {
		if a.Gender == GenderFemale {
			females = append(females, a)
		}
	}
	if len(females) == 0 {
		return "", fmt.Errorf("no female actor found")
	}

	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return females[0].ID, nil
	}

	// Priority 1: exact match
	for _, a := range females {
		if strings.TrimSpace(a.Name) == keyword {
			return a.ID, nil
		}
	}

	// Priority 2: keyword is contained in actor name (or vice versa)
	for _, a := range females {
		aname := strings.TrimSpace(a.Name)
		if strings.Contains(aname, keyword) || strings.Contains(keyword, aname) {
			return a.ID, nil
		}
	}

	// Priority 3: first character match (handles simplified/traditional differences)
	for _, a := range females {
		aname := strings.TrimSpace(a.Name)
		if len(aname) > 0 && len(keyword) > 0 {
			kFirst := string([]rune(keyword)[:1])
			aFirst := string([]rune(aname)[:1])
			if kFirst == aFirst {
				return a.ID, nil
			}
		}
	}

	// Fallback: return first female actor
	return females[0].ID, nil
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
