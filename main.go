package main

import (
	"strings"

	"github.com/planetlabs/go-stac/cmd"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	logger, _ = zap.NewDevelopment()

	viper.SetEnvPrefix("stac")
	viper.AutomaticEnv()                                   // read in environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_")) // eg --my-var = STAC_MY_VAR

	var stac = &cobra.Command{
		Use:   "stac",
		Short: "stac is a cli tool for managing a SpatioTemporal Assec Catalog service",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// viper.BindPFlags allows flags to be set by environment variables
			if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
				return err
			}
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			if viper.GetBool("production") {
				p, _ := zap.NewProduction()
				*logger = *p
			}
			return nil
		},
	}

	stac.PersistentFlags().Bool("production", false, "use production logging presets (default is dev)")

	schema := cmd.Schema(logger)
	schema.AddCommand(cmd.List(logger))
	schema.AddCommand(cmd.Get(logger))
	schema.AddCommand(cmd.Validate(logger))
	schema.AddCommand(cmd.Render(logger))
	schema.AddCommand(cmd.Generate(logger))
	stac.AddCommand(schema)

	if err := stac.Execute(); err != nil {
		logger.Fatal(err.Error())
	}
}
