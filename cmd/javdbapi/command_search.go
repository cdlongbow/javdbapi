package main

import (
	"io"

	cli "github.com/urfave/cli/v3"

	javdbapi "github.com/RPbro/javdbapi"
	"github.com/RPbro/javdbapi/internal/cliapp"
)

func newSearchCommand(buildFetcher fetcherBuilder, stdout io.Writer, stderr io.Writer) *cli.Command {
	return newListCommand("search", "search videos by keyword", "javdbapi search --keyword VR", []cli.Flag{
		&cli.StringFlag{Name: "keyword", Required: true, Usage: "search keyword"},
	}, buildFetcher, stdout, stderr, func(cmd *cli.Command) (cliapp.ListRequest, error) {
		return cliapp.ListRequest{
			Command:  cliapp.CommandSearch,
			Page:     cmd.Int("page"),
			MaxPages: cmd.Int("max-pages"),
			Search:   &javdbapi.SearchQuery{Keyword: cmd.String("keyword")},
		}, nil
	}, func(cmd *cli.Command) string {
		return cmd.String("keyword")
	})
}
