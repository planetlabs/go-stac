package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	// make-links-absolute flags
	flagUrl    = "url"
	flagOutput = "output"

	// common flags
	flagEntry       = "entry"
	flagConcurrency = "concurrency"
	flagRecursion   = "recursion"
	flagLogLevel    = "log-level"
	flagLogFormat   = "log-format"
)

type Enum struct {
	Values   []string
	Default  string
	selected string
}

func (e *Enum) Set(value string) error {
	for _, enum := range e.Values {
		if enum == value {
			e.selected = value
			return nil
		}
	}

	return fmt.Errorf("allowed values are %s", strings.Join(e.Values, ", "))
}

func (e *Enum) String() string {
	if e.selected == "" {
		return e.Default
	}
	return e.selected
}

func toEnvVar(flag string) string {
	return fmt.Sprintf("%s_%s", "STAC", strings.ToUpper(strings.Replace(flag, "-", "_", -1)))
}

var (
	logLevelValues = []string{
		logrus.PanicLevel.String(),
		logrus.FatalLevel.String(),
		logrus.ErrorLevel.String(),
		logrus.WarnLevel.String(),
		logrus.InfoLevel.String(),
		logrus.DebugLevel.String(),
		logrus.TraceLevel.String(),
	}

	logFormatJSON   = "json"
	logFormatText   = "text"
	logFormatValues = []string{logFormatJSON, logFormatText}
)

func configureLogger(ctx *cli.Context) error {
	level, levelErr := logrus.ParseLevel(ctx.String(flagLogLevel))
	if levelErr != nil {
		return fmt.Errorf("unsupported %s '%s'", flagLogLevel, ctx.String(flagLogLevel))
	}
	logrus.SetLevel(level)

	format := ctx.String(flagLogFormat)
	switch format {
	case logFormatJSON:
		logrus.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyLevel: "severity",
			},
		})
		logrus.SetReportCaller(true)
	case logFormatText:
		logrus.SetFormatter(&logrus.TextFormatter{})
	default:
		return fmt.Errorf("unsupported %s '%s'", flagLogFormat, format)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:        "stac",
		Usage:       "STAC Utilities",
		Description: "Utilities for working with Spatio-Temporal Asset Catalog (STAC) metadata.",
		Commands: []*cli.Command{
			validateCommand,
			statsCommand,
			absoluteLinksCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
