package lib

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func GenerateGraph(aggregates []string, nginxAccessLogFilepath string) error {
	// インスタンスを生成
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
	p.Y.Label.Text = "request count"

	logs, err := GetNginxAccessLog(nginxAccessLogFilepath)
	if err != nil {
		return err
	}

	rg := make([]*regexp.Regexp, len(aggregates))

	for i, aggregate := range aggregates {
		rg[i] = regexp.MustCompile(aggregate)
		pointsMap[aggregate] = make(map[float64]float64)
	}

	minTime := math.MaxFloat64

	for _, v := range logs {
		noMatch := true
		for _, r := range rg {
			if r.MatchString(v.GetEndPoint()) {
				pointsMap[r.String()][convertTimeToX(v.Time.Time)] += 1
				minTime = math.Min(minTime, convertTimeToX(v.Time.Time))
				noMatch = false
				break
			}
		}
		// どれにもマッチしなかったら
		if noMatch {
			if _, ok := pointsMap[v.GetEndPoint()]; !ok {
				pointsMap[v.GetEndPoint()] = make(map[float64]float64)
			}
			pointsMap[v.GetEndPoint()][convertTimeToX(v.Time.Time)] += 1
			minTime = math.Min(minTime, convertTimeToX(v.Time.Time))
		}
	}

	pointsList := make([]interface{}, 0)
	// plotするにはplotter.XYs型に変換する必要がある
	for k, v := range pointsMap {
		points := make(plotter.XYs, len(v))
		i := 0
		for x, y := range v {
			points[i].X = x
			points[i].Y = y - minTime
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
		return err
	}

	// legendは左上にする
	p.Legend.Left = true
	p.Legend.Top = true

	// 描画結果を保存
	// "5*vg.Inch" の数値を変更すれば，保存する画像のサイズを調整できます．
	if err := p.Save(5*vg.Inch, 5*vg.Inch, fmt.Sprintf("%s.png", time.Now().Format(time.RFC3339))); err != nil {
		return err
	}
	return nil
}

func convertTimeToX(t time.Time) float64 {
	return float64(t.Hour()*3600 + t.Minute()*60 + t.Second())
}
