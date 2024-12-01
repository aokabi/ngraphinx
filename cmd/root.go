package cmd

import (
	graph "github.com/aokabi/ngraphinx/lib/graph"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "ngraphinx",
		Short: "ngraphinx is a tool to generate graph from nginx access log",
		Long:  "ngraphinx is a tool to generate graph from nginx access log",
	}

	nginxAccessLogFilepath string
	aggregates             string
	imageWidth             graph.Inch
	imageHeight            graph.Inch
	reqMinCountPerSec      int
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	rootCmd.PersistentFlags().StringVar(&aggregates, "aggregates", "", "aggregate endpoint")
	rootCmd.PersistentFlags().StringVar(&nginxAccessLogFilepath, "path", "access.log", "nginx access log path")
	rootCmd.PersistentFlags().IntVar(&imageWidth, "width", 10, "image width(Inch)")
	rootCmd.PersistentFlags().IntVar(&imageHeight, "height", 10, "image height(Inch)")
	rootCmd.PersistentFlags().IntVar(&reqMinCountPerSec, "mincount", 20, "required min request count per sec")
	viper.SetDefault("license", "apache")
}
