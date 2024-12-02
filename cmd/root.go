package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	userLicense string

	rootCmd = &cobra.Command{
		Use:     "ngraphinx",
		Short:   "ngraphinx is a tool to generate graph from nginx access log",
		Long:    "ngraphinx is a tool to generate graph from nginx access log",
		Version: "2.0.4",
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
