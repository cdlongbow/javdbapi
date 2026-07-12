package javdbapi

import (
	"context"
	"fmt"

	"github.com/RPbro/javdbapi/internal/scrape"
	"github.com/RPbro/javdbapi/internal/siteurl"
)

func toPublicNamedRef[ID ~string](ref *scrape.NamedRef) *NamedRef[ID] {
	if ref == nil {
		return nil
	}
	return &NamedRef[ID]{ID: ID(ref.ID), Name: ref.Name}
}

func toPublicDetail(d scrape.Detail) *VideoDetail {
	out := NewVideoDetail(toPublicSummary(d.Summary))
	out.DurationMinutes = d.DurationMinutes
	out.Director = toPublicNamedRef[DirectorID](d.Director)
	out.Maker = toPublicNamedRef[MakerID](d.Maker)
	out.Series = toPublicNamedRef[SeriesID](d.Series)
	for _, a := range d.Actors {
		out.Actors = append(out.Actors, Actor{ID: ActorID(a.ID), Name: a.Name, Gender: Gender(a.Gender)})
	}
	for _, t := range d.Tags {
		out.Tags = append(out.Tags, Tag{Name: t.Name, Path: t.Path})
	}
	for _, s := range d.Screenshots {
		out.Screenshots = append(out.Screenshots, Screenshot{FullURL: s.FullURL, ThumbnailURL: s.ThumbnailURL})
	}
	for _, m := range d.Magnets {
		out.Magnets = append(out.Magnets, Magnet{
			URI:         m.URI,
			Name:        m.Name,
			SizeText:    m.SizeText,
			SizeBytes:   m.SizeBytes,
			FileCount:   m.FileCount,
			HasSubtitle: m.HasSubtitle,
			IsHD:        m.IsHD,
			PublishedAt: m.PublishedAt,
		})
	}
	out.WantCount = d.WantCount
	out.WatchedCount = d.WatchedCount
	return &out
}

// Detail fetches only the video's detail page. Reviews are an independent
// request via Client.Reviews and are never embedded here.
func (c *Client) Detail(ctx context.Context, id VideoID) (*VideoDetail, error) {
	target, err := siteurl.Video(c.baseURL, string(id), string(c.locale))
	if err != nil {
		return nil, &OpError{Op: "detail.build_url", Err: err}
	}

	body, err := c.fetcher.Get(ctx, target)
	if err != nil {
		return nil, mapFetchErr("detail.fetch", target, err)
	}

	doc, err := newDocument(body)
	if err != nil {
		return nil, &OpError{Op: "detail.document", URL: sanitizedURL(target), Err: fmt.Errorf("%w: %v", ErrParse, err)}
	}

	parsed, err := scrape.ParseDetail(doc, string(id))
	if err != nil {
		return nil, mapScrapeErr("detail.parse", target, err)
	}
	c.logWarnings("detail.parse", parsed.Warnings)
	return toPublicDetail(parsed.Value), nil
}
