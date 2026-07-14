package main

import (
	"context"
	"io"

	cli "github.com/urfave/cli/v3"

	javdbapi "github.com/RPbro/javdbapi"
	"github.com/RPbro/javdbapi/internal/cliapp"
)

func newActorCommand(buildFetcher fetcherBuilder, stdout io.Writer, stderr io.Writer) *cli.Command {
	return newListCommand("actor", "list videos from an actor", "javdbapi actor --id neRNX --filter cnsub,download", []cli.Flag{
		&cli.StringFlag{Name: "id", Required: true, Usage: "actor ID or name"},
		newStringFlag("filter", actorFilterSpec, false),
	}, buildFetcher, stdout, stderr, func(cmd *cli.Command) (cliapp.ListRequest, error) {
		rawID := cmd.String("id")
		actorID, err := javdbapi.ParseActorID(rawID)
		if err != nil {
			// Try resolving by name
			actorID, err = resolveActorIDByName(rawID)
			if err != nil {
				return cliapp.ListRequest{}, err
			}
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

// resolveActorIDByName uses the SDK's ActorByName to resolve an actor ID from their name.
func resolveActorIDByName(name string) (javdbapi.ActorID, error) {
	client, err := javdbapi.NewClient(javdbapi.ClientConfig{})
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	actorID, err := client.ActorByName(ctx, name)
	if err != nil {
		return "", err
	}
	return actorID, nil
}
