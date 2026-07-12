package javdbapi

import (
	"context"
	"fmt"

	"github.com/RPbro/javdbapi/internal/scrape"
	"github.com/RPbro/javdbapi/internal/siteurl"
)

func toPublicReviews(reviews []scrape.Review) []Review {
	out := make([]Review, len(reviews))
	for i, r := range reviews {
		out[i] = Review{
			ID:          ReviewID(r.ID),
			Author:      r.Author,
			Score:       r.Score,
			PublishedAt: r.PublishedAt,
			Likes:       r.Likes,
			Content:     r.Content,
		}
	}
	return out
}

// Reviews fetches a video's reviews independently of Detail, with its own
// failure and refresh cycle.
func (c *Client) Reviews(ctx context.Context, id VideoID) ([]Review, error) {
	target, err := siteurl.Reviews(c.baseURL, string(id), string(c.locale))
	if err != nil {
		return nil, &OpError{Op: "reviews.build_url", Err: err}
	}

	body, err := c.fetcher.Get(ctx, target)
	if err != nil {
		return nil, mapFetchErr("reviews.fetch", target, err)
	}

	doc, err := newDocument(body)
	if err != nil {
		return nil, &OpError{Op: "reviews.document", URL: sanitizedURL(target), Err: fmt.Errorf("%w: %v", ErrParse, err)}
	}

	parsed, err := scrape.ParseReviews(doc)
	if err != nil {
		return nil, mapScrapeErr("reviews.parse", target, err)
	}
	c.logWarnings("reviews.parse", parsed.Warnings)
	return toPublicReviews(parsed.Value), nil
}
