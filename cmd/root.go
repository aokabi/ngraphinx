package cmd

import (
	"github.com/aokabi/ngraphinx/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	userLicense string

	rootCmd = &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return lib.GenerateGraph([]string{}, nginxAccessLogFilepath)
		},
	}

	nginxAccessLogFilepath string
	aggregates             string
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	rootCmd.PersistentFlags().StringVar(&aggregates, "aggregates", "", "aggregate endpoint")
	rootCmd.PersistentFlags().StringVar(&nginxAccessLogFilepath, "path", "access.log", "nginx access log path")
	viper.SetDefault("license", "apache")
}
