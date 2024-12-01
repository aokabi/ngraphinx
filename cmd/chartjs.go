package cmd

import (
	"regexp"
	"strings"

	"github.com/aokabi/ngraphinx/lib"
	chartjs "github.com/aokabi/ngraphinx/lib/chatjs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chartjsCmd)
}

var chartjsCmd = &cobra.Command{
	Use:   "chartjs",
	Short: "Chartjs commands",
	Long:  `Chartjs commands`,
	RunE: func(cmd *cobra.Command, args []string) error {
		regStrs := strings.Split(aggregates, ",")
		regexps := make(lib.Regexps, len(regStrs))

		for i, aggregate := range regStrs {
			regexps[i] = regexp.MustCompile(aggregate)
		}

		return chartjs.GenerateGraph(regexps, nginxAccessLogFilepath, nil)
	},
}
