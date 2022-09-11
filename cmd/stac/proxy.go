package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/planetlabs/go-stac/internal/proxy"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var proxyCommand = &cli.Command{
	Name:        "proxy",
	Usage:       "Proxy STAC resources",
	Description: "Starts an HTTP proxy for a STAC entry point.",
	ArgsUsage:   "<entry> Path or URL for the STAC resource.",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    flagPort,
			Usage:   "Listen on this port",
			EnvVars: []string{toEnvVar(flagPort)},
		},
		&cli.GenericFlag{
			Name:  flagLogLevel,
			Usage: fmt.Sprintf("Log level (%s)", strings.Join(logLevelValues, ", ")),
			Value: &Enum{
				Values:  logLevelValues,
				Default: zap.ErrorLevel.String(),
			},
			EnvVars: []string{toEnvVar(flagLogLevel)},
		},
	},
	Action: func(ctx *cli.Context) error {
		args := ctx.Args()
		if args.Len() != 1 {
			cli.ShowCommandHelpAndExit(ctx, ctx.Command.Name, 1)
		}

		port := ctx.Int(flagPort)
		if port == 0 {
			freePort, err := getFreePort()
			if err != nil {
				return fmt.Errorf("failed to get free port: %w", err)
			}
			port = freePort
		}

		logger, sync, logErr := configureLogger(ctx)
		if logErr != nil {
			return logErr
		}
		defer sync()

		entry, err := url.Parse(args.Get(0))
		if err != nil {
			return fmt.Errorf("failed to parse path %q: %w", args.Get(0), err)
		}

		handler, err := proxy.New(&proxy.Options{
			Entry:  entry,
			Logger: logger,
		})
		if err != nil {
			return fmt.Errorf("failed to create proxy: %w", err)
		}

		addr := fmt.Sprintf(":%d", port)

		logger.V(1).Info("Listening", "url", fmt.Sprintf("http://localhost:%d%s", port, handler.GetPath(entry)))
		return http.ListenAndServe(addr, handler)
	},
}
