package main

import (
	"context"
	"io"
	"time"

	cli "github.com/urfave/cli/v3"

	javdbapi "github.com/RPbro/javdbapi"
	"github.com/RPbro/javdbapi/internal/cliapp"
	"github.com/RPbro/javdbapi/internal/clioutput"
)

func newActorDetailCommand(buildFetcher fetcherBuilder, stdout io.Writer, stderr io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "actor-detail",
		Usage:     "fetch actor name and aliases",
		UsageText: "javdbapi actor-detail --id WE4e",
		Flags: append(sharedFlags(),
			&cli.StringFlag{Name: "id", Required: true, Usage: "actor ID, e.g. WE4e"},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			actorID, err := javdbapi.ParseActorID(cmd.String("id"))
			if err != nil {
				return err
			}

			logger := loggerFromCommand(cmd, stderr)
			shared, err := sharedOptionsFromCommand(cmd, stdout, logger)
			if err != nil {
				return err
			}

			store := clioutput.NewStore(shared.OutputDir, time.Now)

			req := cliapp.ListRequest{
				Command: cliapp.CommandActorDetail,
				ActorDetail: &javdbapi.ActorDetailQuery{
					ActorID: actorID,
				},
			}
			req.Shared = shared

			summary, err := cliapp.RunActorDetailCommand(ctx, store, req)
			logSummary(shared.Logger, summary, err)
			return err
		},
	}
}
