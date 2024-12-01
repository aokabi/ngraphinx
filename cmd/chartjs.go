package cmd

import (
	"regexp"
	"strings"

	"github.com/aokabi/ngraphinx/v2/lib"
	chartjs "github.com/aokabi/ngraphinx/v2/lib/chatjs"
	"github.com/spf13/cobra"
)

var (
	maxDatasetNum int
)

func init() {
	rootCmd.AddCommand(chartjsCmd)

	// define flags
	chartjsCmd.PersistentFlags().IntVar(&maxDatasetNum, "maxdataset", 10, "max dataset num")
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

		option := chartjs.NewOption(maxDatasetNum)

		return chartjs.GenerateGraph(regexps, nginxAccessLogFilepath, option)
	},
}
