package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aokabi/ngraphinx/v2/lib"
	chartjs "github.com/aokabi/ngraphinx/v2/lib/chatjs"
	"github.com/spf13/cobra"
)

var (
	maxDatasetNum int
	outputFilePath         string
)

func init() {
	rootCmd.AddCommand(chartjsCmd)

	// define flags
	chartjsCmd.PersistentFlags().IntVar(&maxDatasetNum, "maxdataset", 10, "max dataset num")
	chartjsCmd.PersistentFlags().StringVarP(&outputFilePath, "output", "o", fmt.Sprintf("%s.html", time.Now().Format(time.RFC3339)), "output file path")
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

		option := chartjs.NewOption(maxDatasetNum, outputFilePath)

		return chartjs.GenerateGraph(regexps, nginxAccessLogFilepath, option)
	},
}
