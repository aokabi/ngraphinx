package lib

import (
	"crypto/md5"
	"fmt"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"time"

	"gonum.org/v1/plot/plotutil"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type Inch = int

type Option struct {
	imageWidth  font.Length
	imageHeight font.Length
	minCount    int
}

func NewOption(w, h Inch, minCount int) *Option {
	return &Option{
		imageWidth:  font.Length(w),
		imageHeight: font.Length(h),
		minCount:    minCount,
	}
}

func ApplyPattern(key string, line *plotter.Line, point *plotter.Scatter) {
	md5 := md5.Sum([]byte(key))
	n := 0
	for i, c := range md5 {
		rand.Seed(int64(c)*int64(len(md5)) + int64(i))
		n = n + rand.Intn(65536)
	}
	line.Color = plotutil.Color(n)
	line.Dashes = plotutil.Dashes(n)
	point.Color = plotutil.Color(n)
	point.Shape = plotutil.Shape(n)
}

type PerSec struct {
	count int
	y     float64
}

type namedPoints struct {
	name     string
	points   plotter.XYs
	countSum float64
}

type regexps []*regexp.Regexp

func (rs regexps) findMatchStringFirst(s string) (retr *regexp.Regexp, find bool) {
	for _, r := range rs {
		if r.MatchString(s) {
			return r, true
		}
	}
	return nil, false
}

type pointsMap = map[string]map[float64]*PerSec

// x: float64, y: PerSec を plotter.XYsに詰め直す
func convPointsMap2NamedPointsSlice(pointsMap pointsMap, pointCountSumMap map[float64]int, option *Option, minTime float64, mapPerSecToY func(PerSec) float64) []namedPoints {
	namedPointsArray := make([]namedPoints, 0)
	for k, v := range pointsMap {
		points := make(plotter.XYs, len(v))
		i := 0
		countSum := 0.0
		for x, y := range v {
			if pointCountSumMap[x] < option.minCount {
				continue
			}
			points[i].X = x - minTime
			points[i].Y = mapPerSecToY(*y)
			countSum += points[i].Y
			i++
		}
		points = points[0:i]
		// sort points by x
		sort.Slice(points, func(i, j int) bool {
			return points[i].X < points[j].X
		})
		namedPointsArray = append(namedPointsArray, namedPoints{k, points, countSum})
	}
	return namedPointsArray
}

func generateGraphImpl(p *plot.Plot, aggregates []string, nginxAccessLogFilepath string, option *Option,
	mapLogToPerSec func(v log) float64, mapPerSecToY func(ps PerSec) float64) error {
	logs, err := GetNginxAccessLog(nginxAccessLogFilepath)
	if err != nil {
		return err
	}
	regexps := make(regexps, len(aggregates))

	for i, aggregate := range aggregates {
		regexps[i] = regexp.MustCompile(aggregate)
	}

	minTime := math.MaxFloat64

	// 単位時間ごとのリクエスト数を数えるのが大変なので一旦マップにする
	pointsMap := make(map[string]map[float64]*PerSec)
	for _, v := range logs {
		endpoint, err := v.GetEndPoint()
		if err != nil {
			continue
		}
		r, find := regexps.findMatchStringFirst(endpoint)
		// どれにもマッチしなかったら
		if find {
			key := makeKey(v.GetMethod(), r.String())
			if _, ok := pointsMap[key]; !ok {
				pointsMap[key] = make(map[float64]*PerSec)
			}
			logTime := convertTimeToX(v.Time.Time)
			if _, ok := pointsMap[key][logTime]; !ok {
				pointsMap[key][logTime] = &PerSec{
					count: 0,
					y:     0,
				}
			}
			pointsMap[key][logTime].count += 1
			pointsMap[key][logTime].y += mapLogToPerSec(v)
			minTime = math.Min(minTime, logTime)
		} else {
			key := makeKey(v.GetMethod(), endpoint)
			if _, ok := pointsMap[key]; !ok {
				pointsMap[key] = make(map[float64]*PerSec)
			}
			logTime := convertTimeToX(v.Time.Time)
			if _, ok := pointsMap[key][logTime]; !ok {
				pointsMap[key][logTime] = &PerSec{
					count: 0,
					y:     0,
				}
			}
			pointsMap[key][logTime].count += 1
			pointsMap[key][logTime].y += mapLogToPerSec(v)
			minTime = math.Min(minTime, logTime)
		}
	}
	pointCountSumMap := make(map[float64]int)
	for _, v := range pointsMap {
		for x, y := range v {
			pointCountSumMap[x] += y.count
		}
	}

	// plotするにはplotter.XYs型に変換する必要がある
	nameAndPoints := convPointsMap2NamedPointsSlice(pointsMap, pointCountSumMap, option, minTime, mapPerSecToY)

	// Legend が挿入順に生成されるため、時間総和数でソートする用途
	sort.Slice(nameAndPoints, func(i, j int) bool {
		return nameAndPoints[i].countSum > nameAndPoints[j].countSum
	})
	for _, v := range nameAndPoints {
		lpLine, lpPoints, err := plotter.NewLinePoints(v.points)
		if err != nil {
			return err
		}
		ApplyPattern(v.name, lpLine, lpPoints)
		p.Add(lpLine, lpPoints)
		p.Legend.Add(v.name, lpLine, lpPoints)
	}
	return nil
}

// generate graph of request time sum per second
func generateReqTimeSumGraph(aggregates []string, nginxAccessLogFilepath string, option *Option) (*plot.Plot, error) {
	// 表示項目の設定
	p := plot.New()
	p.Title.Text = "access.log"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "request time sum / sec"
	// legendは左上にする
	p.Legend.Left = true
	p.Legend.Top = true
	err := generateGraphImpl(p, aggregates, nginxAccessLogFilepath, option, func(v log) float64 {
		return v.ReqTime
	}, func(ps PerSec) float64 {
		// return ps.y / ps.count // if average
		return ps.y
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// generate graph of request num per second
func generateCountGraph(aggregates []string, nginxAccessLogFilepath string, option *Option) (*plot.Plot, error) {
	// 表示項目の設定
	p := plot.New()
	p.Title.Text = "access.log"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "request count / sec"
	// legendは左上にする
	p.Legend.Left = true
	p.Legend.Top = true
	err := generateGraphImpl(p, aggregates, nginxAccessLogFilepath, option, func(v log) float64 {
		return 1.0
	}, func(ps PerSec) float64 {
		return ps.y
	})
	if err != nil {
		return nil, err
	}
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
				p, err = generateReqTimeSumGraph(aggregates, nginxAccessLogFilepath, option)
				if err != nil {
					return err
				}
			} else {
				p, err = generateCountGraph(aggregates, nginxAccessLogFilepath, option)
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
