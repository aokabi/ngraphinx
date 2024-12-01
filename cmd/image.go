package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aokabi/ngraphinx/v2/lib"
	graph "github.com/aokabi/ngraphinx/v2/lib/graph"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(imageCmd)

	// define flags
	imageCmd.PersistentFlags().StringVarP(&outputFilePath, "output", "o", fmt.Sprintf("%s.png", time.Now().Format(time.RFC3339)), "output file path")
}

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Image commands",
	Long:  `Image commands`,
	RunE: func(cmd *cobra.Command, args []string) error {
		regStrs := strings.Split(aggregates, ",")
		option := graph.NewOption(imageWidth, imageHeight, reqMinCountPerSec, outputFilePath)

		regexps := make(lib.Regexps, len(regStrs))

		for i, aggregate := range regStrs {
			regexps[i] = regexp.MustCompile(aggregate)
		}

		return graph.GenerateGraph(regexps, nginxAccessLogFilepath, option)
	},
}
