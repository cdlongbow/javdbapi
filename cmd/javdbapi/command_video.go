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

func newVideoCommand(buildFetcher fetcherBuilder, stdout io.Writer, stderr io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "video",
		Usage:     "fetch full video detail",
		UsageText: "javdbapi video --id ZNdEbV",
		Flags: append(sharedFlags(),
			&cli.StringFlag{Name: "id", Required: true, Usage: "video ID, e.g. ZNdEbV"},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			rawID := cmd.String("id")
			id, err := javdbapi.ResolveVideoID(rawID)
			if err != nil {
				return err
			}

			logger := loggerFromCommand(cmd, stderr)
			shared, err := sharedOptionsFromCommand(cmd, stdout, logger)
			if err != nil {
				return err
			}

			fetcher, err := buildFetcher(cmd, logger)
			if err != nil {
				return err
			}
			store := clioutput.NewStore(shared.OutputDir, time.Now)

			// Use the raw input as the code for naming; falls back to resolved ID
			// if the raw input wasn't a valid code.
			req := cliapp.VideoRequest{Shared: shared, ID: id, Code: rawID}
			summary, err := cliapp.RunVideoCommand(ctx, fetcher, store, req)
			logSummary(shared.Logger, summary, err)
			return err
		},
	}
}
