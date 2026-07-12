package main

import (
	"io"

	cli "github.com/urfave/cli/v3"

	javdbapi "github.com/RPbro/javdbapi"
	"github.com/RPbro/javdbapi/internal/cliapp"
)

func newHomeCommand(buildFetcher fetcherBuilder, stdout io.Writer, stderr io.Writer) *cli.Command {
	return newListCommand("home", "browse home page listings", "javdbapi home --type censored --filter all --sort publish", []cli.Flag{
		newStringFlag("type", homeTypeSpec, false),
		newStringFlag("filter", homeFilterSpec, false),
		newStringFlag("sort", homeSortSpec, false),
	}, buildFetcher, stdout, stderr, func(cmd *cli.Command) (cliapp.ListRequest, error) {
		homeType, err := parseHomeType(cmd.String("type"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		homeFilter, err := parseHomeFilter(cmd.String("filter"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		homeSort, err := parseHomeSort(cmd.String("sort"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		return cliapp.ListRequest{
			Command:  cliapp.CommandHome,
			Page:     cmd.Int("page"),
			MaxPages: cmd.Int("max-pages"),
			Home: &javdbapi.HomeQuery{
				Type:   homeType,
				Filter: homeFilter,
				Sort:   homeSort,
			},
		}, nil
	})
}
