package lib

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type Inch = int

type Option struct {
	imageWidth  font.Length
	imageHeight font.Length
}

func NewOption(w, h Inch) *Option {
	return &Option{
		imageWidth:  font.Length(w),
		imageHeight: font.Length(h),
	}
}

func generateReqTimeAverageGraph(aggregates []string, nginxAccessLogFilepath string) (*plot.Plot, error) {
	p := plot.New()

	type xyAxis struct {
		x float64
		y float64
	}

	type summary struct {
		count int
		time  float64
	}

	// 単位時間ごとのリクエスト数を数えるのが大変なので一旦マップにする
	pointsMap := make(map[string]map[float64]*summary)

	// 表示項目の設定
	p.Title.Text = "access.log"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "request time average"

	logs, err := GetNginxAccessLog(nginxAccessLogFilepath)
	if err != nil {
		return nil, err
	}

	rg := make([]*regexp.Regexp, len(aggregates))

	for i, aggregate := range aggregates {
		rg[i] = regexp.MustCompile(aggregate)
	}

	minTime := math.MaxFloat64

	for _, v := range logs {
		noMatch := true
		endpoint, err := v.GetEndPoint()
		if err != nil {
			return nil, err
		}

		for _, r := range rg {
			if r.MatchString(endpoint) {
				if _, ok := pointsMap[makeKey(v.GetMethod(), r.String())]; !ok {
					pointsMap[makeKey(v.GetMethod(), r.String())] = make(map[float64]*summary)
				}
				if _, ok := pointsMap[makeKey(v.GetMethod(), r.String())][convertTimeToX(v.Time.Time)]; !ok {
					pointsMap[makeKey(v.GetMethod(), r.String())][convertTimeToX(v.Time.Time)] = &summary{
						count: 0,
						time:  0,
					}
				}
				pointsMap[makeKey(v.GetMethod(), r.String())][convertTimeToX(v.Time.Time)].count += 1
				pointsMap[makeKey(v.GetMethod(), r.String())][convertTimeToX(v.Time.Time)].time += v.ReqTime
				minTime = math.Min(minTime, convertTimeToX(v.Time.Time))
				noMatch = false
				break
			}
		}
		// どれにもマッチしなかったら
		if noMatch {
			if _, ok := pointsMap[makeKey(v.GetMethod(), endpoint)]; !ok {
				pointsMap[makeKey(v.GetMethod(), endpoint)] = make(map[float64]*summary)
			}
			if _, ok := pointsMap[makeKey(v.GetMethod(), endpoint)][convertTimeToX(v.Time.Time)]; !ok {
				pointsMap[makeKey(v.GetMethod(), endpoint)][convertTimeToX(v.Time.Time)] = &summary{
					count: 0,
					time:  0,
				}
			}
			pointsMap[makeKey(v.GetMethod(), endpoint)][convertTimeToX(v.Time.Time)].count += 1
			pointsMap[makeKey(v.GetMethod(), endpoint)][convertTimeToX(v.Time.Time)].time += v.ReqTime
			minTime = math.Min(minTime, convertTimeToX(v.Time.Time))
		}
	}

	pointsList := make([]interface{}, 0)
	// plotするにはplotter.XYs型に変換する必要がある
	for k, v := range pointsMap {
		points := make(plotter.XYs, len(v))
		i := 0
		for x, y := range v {
			points[i].X = x - minTime
			points[i].Y = float64(y.time) / float64(y.count)
			i++
		}
		// sort points by x
		sort.Slice(points, func(i, j int) bool {
			return points[i].X < points[j].X
		})
		pointsList = append(pointsList, k, points)
	}
	// plotter.XYs型に変換してplot.Addを呼び出す
	if err := plotutil.AddLinePoints(p, pointsList...); err != nil {
		return nil, err
	}

	// legendは左上にする
	p.Legend.Left = true
	p.Legend.Top = true

	return p, nil

}

func generateCountGraph(aggregates []string, nginxAccessLogFilepath string) (*plot.Plot, error) {
	p := plot.New()

	type xyAxis struct {
		x float64
		y float64
	}

	// 単位時間ごとのリクエスト数を数えるのが大変なので一旦マップにする
	pointsMap := make(map[string]map[float64]float64)

	// 表示項目の設定
	p.Title.Text = "access.log"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "request count / sec"

	logs, err := GetNginxAccessLog(nginxAccessLogFilepath)
	if err != nil {
		return nil, err
	}

	rg := make([]*regexp.Regexp, len(aggregates))

	for i, aggregate := range aggregates {
		rg[i] = regexp.MustCompile(aggregate)
	}

	minTime := math.MaxFloat64

	for _, v := range logs {
		noMatch := true
		endpoint, err := v.GetEndPoint()
		if err != nil {
			return nil, err
		}

		for _, r := range rg {
			if r.MatchString(endpoint) {
				if _, ok := pointsMap[makeKey(v.GetMethod(), r.String())]; !ok {
					pointsMap[makeKey(v.GetMethod(), r.String())] = make(map[float64]float64)
				}
				pointsMap[makeKey(v.GetMethod(), r.String())][convertTimeToX(v.Time.Time)] += 1
				minTime = math.Min(minTime, convertTimeToX(v.Time.Time))
				noMatch = false
				break
			}
		}
		// どれにもマッチしなかったら
		if noMatch {
			if _, ok := pointsMap[makeKey(v.GetMethod(), endpoint)]; !ok {
				pointsMap[makeKey(v.GetMethod(), endpoint)] = make(map[float64]float64)
			}
			pointsMap[makeKey(v.GetMethod(), endpoint)][convertTimeToX(v.Time.Time)] += 1
			minTime = math.Min(minTime, convertTimeToX(v.Time.Time))
		}
	}

	pointsList := make([]interface{}, 0)
	// plotするにはplotter.XYs型に変換する必要がある
	for k, v := range pointsMap {
		points := make(plotter.XYs, len(v))
		i := 0
		for x, y := range v {
			points[i].X = x - minTime
			points[i].Y = y
			i++
		}
		// sort points by x
		sort.Slice(points, func(i, j int) bool {
			return points[i].X < points[j].X
		})
		pointsList = append(pointsList, k, points)
	}
	// plotter.XYs型に変換してplot.Addを呼び出す
	if err := plotutil.AddLinePoints(p, pointsList...); err != nil {
		return nil, err
	}

	// legendは左上にする
	p.Legend.Left = true
	p.Legend.Top = true

	return p, nil
}

func GenerateGraph(aggregates []string, nginxAccessLogFilepath string, option *Option) error {
	// インスタンスを生成
	// 縦に２つ並べる
	const rows, cols = 2, 1
	plots := make([][]*plot.Plot, rows)
	for j := 0; j < rows; j++ {
		plots[j] = make([]*plot.Plot, cols)
		for i := 0; i < cols; i++ {
			var p *plot.Plot
			var err error
			if j == 0 {
				p, err = generateCountGraph(aggregates, nginxAccessLogFilepath)
				if err != nil {
					return err
				}
			} else {
				p, err = generateReqTimeAverageGraph(aggregates, nginxAccessLogFilepath)
				if err != nil {
					return err
				}
			}
			plots[j][i] = p
		}
	}
	// 描画
	img := vgimg.New(option.imageWidth*vg.Inch, option.imageHeight*vg.Inch)
	dc := draw.New(img)
	t := draw.Tiles{
		Rows:      rows,
		Cols:      cols,
		PadX:      vg.Millimeter,
		PadY:      vg.Millimeter,
		PadTop:    vg.Points(2),
		PadBottom: vg.Points(2),
		PadLeft:   vg.Points(2),
		PadRight:  vg.Points(2),
	}

	canvases := plot.Align(plots, t, dc)
	for j := 0; j < rows; j++ {
		for i := 0; i < cols; i++ {
			if plots[j][i] != nil {
				plots[j][i].Draw(canvases[j][i])
			}
		}
	}

	w, err := os.Create(fmt.Sprintf("%s.png", time.Now().Format(time.RFC3339)))
	if err != nil {
		return err
	}
	defer w.Close()
	png := vgimg.PngCanvas{Canvas: img}
	if _, err := png.WriteTo(w); err != nil {
		return err
	}

	return nil
}

func convertTimeToX(t time.Time) float64 {
	return float64(t.Hour()*3600 + t.Minute()*60 + t.Second())
}

func makeKey(httpMethod, endpoint string) string {
	return fmt.Sprintf("%s %s", httpMethod, endpoint)
}
