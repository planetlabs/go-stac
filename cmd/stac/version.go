package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var (
	version = "development"
	commit  = "none"
	date    = "unknown"
)

var versionCommand = &cli.Command{
	Name:        "version",
	Usage:       "Print build information",
	Description: "Prints the build information for the executable.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    flagVerbose,
			Usage:   "Include additional metadata with the build version",
			EnvVars: []string{toEnvVar(flagVerbose)},
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Bool(flagVerbose) {
			fmt.Printf("%s %s %s\n", version, commit, date)
			return nil
		}
		fmt.Println(version)
		return nil
	},
}
