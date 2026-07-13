package main

import (
	"context"
	"io"
	"strings"

	cli "github.com/urfave/cli/v3"

	javdbapi "github.com/RPbro/javdbapi"
	"github.com/RPbro/javdbapi/internal/cliapp"
)

func newActorCommand(buildFetcher fetcherBuilder, stdout io.Writer, stderr io.Writer) *cli.Command {
	return newListCommand("actor", "list videos from an actor", "javdbapi actor --id neRNX --filter cnsub,download", []cli.Flag{
		&cli.StringFlag{Name: "id", Required: true, Usage: "actor ID"},
		newStringFlag("filter", actorFilterSpec, false),
	}, buildFetcher, stdout, stderr, func(cmd *cli.Command) (cliapp.ListRequest, error) {
		actorID, err := javdbapi.ParseActorID(cmd.String("id"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		filters, err := parseActorFilters(cmd.String("filter"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		return cliapp.ListRequest{
			Command:  cliapp.CommandActor,
			Page:     cmd.Int("page"),
			MaxPages: cmd.Int("max-pages"),
			Actor: &javdbapi.ActorVideosQuery{
				ActorID: actorID,
				Filters: filters,
			},
		}, nil
	}, func(cmd *cli.Command) string {
		return cmd.String("id")
	})
}

// resolveActorName fetches the first page of actor's videos, finds a single
// work (单体作品) video and extracts the actor's name from its detail. If no
// single work exists, it falls back to the first video's detail. Female actors
// with multiple matches are joined by commas.
func resolveActorName(ctx context.Context, fetcher cliapp.Fetcher, actorID string) (string, error) {
	actor, err := javdbapi.ParseActorID(actorID)
	if err != nil {
		return "", err
	}
	page, err := fetcher.ActorVideos(ctx, javdbapi.ActorVideosQuery{
		ActorID: actor,
		Filters: []javdbapi.ActorFilter{javdbapi.ActorFilterAll},
		Page:    1,
	})
	if err != nil || len(page.Items) == 0 {
		return "", err
	}

	// Try single work first
	for _, item := range page.Items {
		detail, err := fetcher.Detail(ctx, item.ID)
		if err != nil || detail == nil {
			continue
		}
		for _, tag := range detail.Tags {
			if tag.Name == "單體作品" || tag.Name == "单体作品" {
				return extractFemaleNames(detail.Actors, actorID), nil
			}
		}
	}

	// Fallback: use the first video's detail
	detail, err := fetcher.Detail(ctx, page.Items[0].ID)
	if err != nil || detail == nil {
		return "", err
	}
	return extractFemaleNames(detail.Actors, actorID), nil
}

// extractFemaleNames returns the names of female actors whose ID matches the
// given actorID, joined by commas if there are multiple.
func extractFemaleNames(actors []javdbapi.Actor, targetID string) string {
	var names []string
	for _, a := range actors {
		if string(a.ID) == targetID && a.Gender == "female" {
			names = append(names, a.Name)
		}
	}
	if len(names) == 0 {
		return ""
	}
	return strings.Join(names, ",")
}
