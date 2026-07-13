package main

import (
	"io"

	cli "github.com/urfave/cli/v3"

	javdbapi "github.com/RPbro/javdbapi"
	"github.com/RPbro/javdbapi/internal/cliapp"
)

func newRankingCommand(buildFetcher fetcherBuilder, stdout io.Writer, stderr io.Writer) *cli.Command {
	return newListCommand("ranking", "fetch ranked videos", "javdbapi ranking --period weekly --type censored", []cli.Flag{
		newStringFlag("period", rankingPeriodSpec, true),
		newStringFlag("type", rankingTypeSpec, true),
	}, buildFetcher, stdout, stderr, func(cmd *cli.Command) (cliapp.ListRequest, error) {
		period, err := parseRankingPeriod(cmd.String("period"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		rankingType, err := parseRankingType(cmd.String("type"))
		if err != nil {
			return cliapp.ListRequest{}, err
		}
		return cliapp.ListRequest{
			Command:  cliapp.CommandRanking,
			Page:     cmd.Int("page"),
			MaxPages: cmd.Int("max-pages"),
			Ranking: &javdbapi.RankingQuery{
				Period: period,
				Type:   rankingType,
			},
		}, nil
	}, nil)
}
