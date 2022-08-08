package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// validate flags
	flagSchema = "schema"

	// make-links-absolute flags
	flagUrl = "url"

	// version flags
	flagVerbose = "verbose"

	// common flags
	flagLogLevel    = "log-level"
	flagEntry       = "entry"
	flagOutput      = "output"
	flagNoRecursion = "no-recursion"
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
		zap.DebugLevel.String(),
		zap.InfoLevel.String(),
		zap.ErrorLevel.String(),
	}
)

func configureLogger(ctx *cli.Context) (*logr.Logger, func(), error) {
	level, levelErr := zap.ParseAtomicLevel(ctx.String(flagLogLevel))
	if levelErr != nil {
		return nil, nil, levelErr
	}

	config := &zap.Config{
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: zapcore.LowercaseColorLevelEncoder,
			TimeKey:     "time",
			EncodeTime:  zapcore.RFC3339TimeEncoder,
		},
		Level:            level,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	zapLogger, configErr := config.Build()
	if configErr != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", configErr)
	}

	sync := func() {
		_ = zapLogger.Sync()
	}

	logger := zapr.NewLogger(zapLogger)
	return &logger, sync, nil
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
			formatCommand,
			versionCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
