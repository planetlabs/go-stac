package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/planetlabs/go-stac/internal/proxy"
	"github.com/planetlabs/go-stac/internal/view"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var viewCommand = &cli.Command{
	Name:        "view",
	Usage:       "Browse STAC resources",
	Description: "Opens a viewer for browsing catalogs, collections, and items.",
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

		proxyHandler, proxyErr := proxy.New(&proxy.Options{
			Base:   "proxy",
			Entry:  entry,
			Logger: logger,
			NoCORS: true,
		})
		if proxyErr != nil {
			return fmt.Errorf("failed to create proxy handler: %w", proxyErr)
		}

		viewHandler, viewErr := view.New(&view.Options{
			Base:   "view",
			Logger: logger,
		})
		if viewErr != nil {
			return fmt.Errorf("failed to create view handler: %w", viewErr)
		}

		mux := http.NewServeMux()
		mux.Handle("/proxy/", proxyHandler)
		mux.Handle("/view/", viewHandler)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		}

		viewUrl := fmt.Sprintf("http://localhost:%d%s", port, viewHandler.GetPath(entry))
		logger.V(1).Info("Listening", "url", viewUrl)

		go func() {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.Error(err, "Server failed")
			}
		}()

		openErr := browser.OpenURL(viewUrl)
		if openErr != nil {
			logger.Error(openErr, "Failed to open browser")
		}

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)

		<-quit
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}

		return nil
	},
}
