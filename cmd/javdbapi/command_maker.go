package main

import (
	"io"

	cli "github.com/urfave/cli/v3"

	javdbapi "github.com/RPbro/javdbapi"
	"github.com/RPbro/javdbapi/internal/cliapp"
)

func newMakerCommand(buildFetcher fetcherBuilder, stdout io.Writer, stderr io.Writer) *cli.Command {
	return newListCommand("maker", "list videos from a maker", "javdbapi maker --id 7R --filter playable", []cli.Flag{
		&cli.StringFlag{Name: "id", Required: true, Usage: "maker ID"},
		newStringFlag("filter", makerFilterSpec, false),
	}, buildFetcher, stdout, stderr, func(cmd *cli.Command) (cliapp.ListRequest, error) {
		makerID, err := javdbapi.ParseMakerID(cmd.String("id"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		filter, err := parseMakerFilter(cmd.String("filter"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		return cliapp.ListRequest{
			Command:  cliapp.CommandMaker,
			Page:     cmd.Int("page"),
			MaxPages: cmd.Int("max-pages"),
			Maker: &javdbapi.MakerVideosQuery{
				MakerID: makerID,
				Filter:  filter,
			},
		}, nil
	})
}
