package cmd

import (
	"regexp"
	"strings"

	"github.com/aokabi/ngraphinx/lib"
	graph "github.com/aokabi/ngraphinx/lib/graph"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(imageCmd)
}

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Image commands",
	Long:  `Image commands`,
	RunE: func(cmd *cobra.Command, args []string) error {
		regStrs := strings.Split(aggregates, ",")
		option := graph.NewOption(imageWidth, imageHeight, reqMinCountPerSec)

		regexps := make(lib.Regexps, len(regStrs))

		for i, aggregate := range regStrs {
			regexps[i] = regexp.MustCompile(aggregate)
		}

		return graph.GenerateGraph(regexps, nginxAccessLogFilepath, option)
	},
}
