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

var (
	imageWidth             graph.Inch
	imageHeight            graph.Inch
	reqMinCountPerSec      int
	outputPngFilePath      string
)

func init() {
	rootCmd.AddCommand(imageCmd)

	// define flags
	imageCmd.PersistentFlags().StringVarP(&outputPngFilePath, "output", "o", fmt.Sprintf("%s.png", time.Now().Format(time.RFC3339)), "output file path")
	imageCmd.PersistentFlags().IntVar(&imageWidth, "width", 10, "image width(Inch)")
	imageCmd.PersistentFlags().IntVar(&imageHeight, "height", 10, "image height(Inch)")
	imageCmd.PersistentFlags().IntVar(&reqMinCountPerSec, "mincount", 20, "required min request count per sec")
}

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Image commands",
	Long:  `Image commands`,
	RunE: func(cmd *cobra.Command, args []string) error {
		regStrs := strings.Split(aggregates, ",")
		option := graph.NewOption(imageWidth, imageHeight, reqMinCountPerSec, outputPngFilePath)

		regexps := make(lib.Regexps, len(regStrs))

		for i, aggregate := range regStrs {
			regexps[i] = regexp.MustCompile(aggregate)
		}

		return graph.GenerateGraph(regexps, nginxAccessLogFilepath, option)
	},
}
