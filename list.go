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
