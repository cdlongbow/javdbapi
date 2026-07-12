package cliapp

import (
	"context"

	javdbapi "github.com/RPbro/javdbapi"
)

// Fetcher is the narrow surface RunListCommand and RunVideoCommand depend on.
// ClientFetcher adapts a *javdbapi.Client to it; tests supply their own fake.
type Fetcher interface {
	Home(context.Context, javdbapi.HomeQuery) (javdbapi.Page[javdbapi.VideoSummary], error)
	Search(context.Context, javdbapi.SearchQuery) (javdbapi.Page[javdbapi.VideoSummary], error)
	MakerVideos(context.Context, javdbapi.MakerVideosQuery) (javdbapi.Page[javdbapi.VideoSummary], error)
	ActorVideos(context.Context, javdbapi.ActorVideosQuery) (javdbapi.Page[javdbapi.VideoSummary], error)
	Ranking(context.Context, javdbapi.RankingQuery) (javdbapi.Page[javdbapi.VideoSummary], error)
	Detail(context.Context, javdbapi.VideoID) (*javdbapi.VideoDetail, error)
	Reviews(context.Context, javdbapi.VideoID) ([]javdbapi.Review, error)
}

type ClientFetcher struct {
	Client *javdbapi.Client
}

func (f ClientFetcher) Home(ctx context.Context, q javdbapi.HomeQuery) (javdbapi.Page[javdbapi.VideoSummary], error) {
	return f.Client.Home(ctx, q)
}

func (f ClientFetcher) Search(ctx context.Context, q javdbapi.SearchQuery) (javdbapi.Page[javdbapi.VideoSummary], error) {
	return f.Client.Search(ctx, q)
}

func (f ClientFetcher) MakerVideos(ctx context.Context, q javdbapi.MakerVideosQuery) (javdbapi.Page[javdbapi.VideoSummary], error) {
	return f.Client.MakerVideos(ctx, q)
}

func (f ClientFetcher) ActorVideos(ctx context.Context, q javdbapi.ActorVideosQuery) (javdbapi.Page[javdbapi.VideoSummary], error) {
	return f.Client.ActorVideos(ctx, q)
}

func (f ClientFetcher) Ranking(ctx context.Context, q javdbapi.RankingQuery) (javdbapi.Page[javdbapi.VideoSummary], error) {
	return f.Client.Ranking(ctx, q)
}

func (f ClientFetcher) Detail(ctx context.Context, id javdbapi.VideoID) (*javdbapi.VideoDetail, error) {
	return f.Client.Detail(ctx, id)
}

func (f ClientFetcher) Reviews(ctx context.Context, id javdbapi.VideoID) ([]javdbapi.Review, error) {
	return f.Client.Reviews(ctx, id)
}
