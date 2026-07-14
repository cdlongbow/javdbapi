package main

import (
	"io"

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
