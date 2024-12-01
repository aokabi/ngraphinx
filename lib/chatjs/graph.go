package chartjs

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"math"
	"os"
	"sort"
	"time"

	"github.com/aokabi/ngraphinx/v2/lib"
	"github.com/aokabi/ngraphinx/v2/lib/nginx"
	"gonum.org/v1/plot/plotutil"
)

type Option struct {
	maxDatasetNum int
	outputFilePath string
}

func NewOption(maxDatasetNum int, outputFilePath string) *Option {
	return &Option{
		maxDatasetNum: maxDatasetNum,
		outputFilePath: outputFilePath,
	}
}

func GenerateGraph(regexps lib.Regexps, logFilePath string, option *Option) error {
	logs, err := nginx.GetNginxAccessLog(logFilePath)
	if err != nil {
		return err
	}

	// request count / sec
	mapLogToPerSec := func(_ nginx.Log) float64 {
		return 1.0
	}
	pointsMap := generateGraphImpl(logs, regexps, mapLogToPerSec)
	if err != nil {
		return err
	}

	// request time sum / sec
	mapLogToPerSec = func(l nginx.Log) float64 {
		return l.ReqTime
	}
	pointsMap2 := generateGraphImpl(logs, regexps, mapLogToPerSec)
	if err != nil {
		return err
	}

	r, err := generateHTML(pointsMap, pointsMap2, option)
	if err != nil {
		return err
	}

	// ここでHTMLを出力する
	file, err := os.Create(option.outputFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, r); err != nil {
		return err
	}

	return nil
}

type PerSec struct {
	count int
	y     float64
}

type endpointKey string

type pointsMap map[endpointKey]map[float64]*PerSec

// nginx logを line graphに出力するために集計する
func generateGraphImpl(logs []nginx.Log, regexps lib.Regexps, mapLogToPerSec func(v nginx.Log) float64) pointsMap {
	minTime := math.MaxFloat64

	// 単位時間ごとのリクエスト数を数えるのが大変なので一旦マップにする
	pMap := make(pointsMap)
	for _, v := range logs {
		endpoint, err := v.GetEndPoint()
		if err != nil {
			continue
		}
		r, find := regexps.FindMatchStringFirst(endpoint)
		var key endpointKey
		if find {
			key = makeKey(v.GetMethod(), r.String())
		} else {
			// どれにもマッチしなかったら
			key = makeKey(v.GetMethod(), endpoint)
		}
		if _, ok := pMap[key]; !ok {
			pMap[key] = make(map[float64]*PerSec)
		}
		logTime := convertTimeToX(v.Time.Time)
		if _, ok := pMap[key][logTime]; !ok {
			pMap[key][logTime] = &PerSec{
				count: 0,
				y:     0,
			}
		}
		pMap[key][logTime].count += 1
		pMap[key][logTime].y += mapLogToPerSec(v)
		minTime = math.Min(minTime, logTime)
	}

	pointCountSumMap := make(map[float64]int)
	for _, v := range pMap {
		for x, y := range v {
			pointCountSumMap[x] += y.count
		}
	}

	// normalize
	normalizedPointsMap := make(pointsMap)
	for k, v := range pMap {
		normalizedPointsMap[k] = make(map[float64]*PerSec, len(v))
		for x, y := range v {
			normalizedPointsMap[k][x-minTime] = y
		}
	}
	return normalizedPointsMap
}

func convertTimeToX(t time.Time) float64 {
	return float64(t.Hour()*3600 + t.Minute()*60 + t.Second())
}
func makeKey(httpMethod, endpoint string) endpointKey {
	return endpointKey(fmt.Sprintf("%s %s", httpMethod, endpoint))
}

type templateValues struct {
	DataSets  []*dataset
	Title     string
	DataSets2 []*dataset
	Title2    string
}

type dataset struct {
	Label           string
	Data            []*point
	BorderColor     string
	BackgroundColor string
}

type point struct {
	X float64
	Y float64
}

var Colors = plotutil.DefaultColors

func generateHTML(points pointsMap, points2 pointsMap, option *Option) (io.Reader, error) {
	// request count / sec
	datasets := make([]*dataset, 0)
	for k, v := range points {
		data := make([]*point, 0)
		for x, y := range v {
			data = append(data, &point{
				X: x,
				Y: y.y,
			})
		}
		sort.SliceStable(data, func(i, j int) bool {
			return data[i].X < data[j].X
		})

		datasets = append(datasets, &dataset{
			Label: string(k),
			Data:  data,
		})
	}
	sort.SliceStable(datasets, func(i, j int) bool {
		isum := 0.0
		jsum := 0.0
		for _, p := range datasets[i].Data {
			isum += p.Y
		}
		for _, p := range datasets[j].Data {
			jsum += p.Y
		}
		return isum > jsum
	})

	// limit the number of datasets
	if len(datasets) > option.maxDatasetNum {
		datasets = datasets[:option.maxDatasetNum]
	}

	for i, d := range datasets {
		r, g, b, _ := Colors[i%len(Colors)].RGBA()
		d.BorderColor = fmt.Sprintf("rgba(%d, %d, %d, 1)", r>>8, g>>8, b>>8)
		d.BackgroundColor = fmt.Sprintf("rgba(%d, %d, %d, 0.2)", r>>8, g>>8, b>>8)
	}

	values := templateValues{
		DataSets: datasets,
		Title:    "request count / sec",
	}

	{

		// request time sum / sec
		datasets := make([]*dataset, 0)
		for k, v := range points2 {
			data := make([]*point, 0)
			for x, y := range v {
				data = append(data, &point{
					X: x,
					Y: y.y,
				})
			}
			sort.SliceStable(data, func(i, j int) bool {
				return data[i].X < data[j].X
			})

			datasets = append(datasets, &dataset{
				Label: string(k),
				Data:  data,
			})
		}
		sort.SliceStable(datasets, func(i, j int) bool {
			isum := 0.0
			jsum := 0.0
			for _, p := range datasets[i].Data {
				isum += p.Y
			}
			for _, p := range datasets[j].Data {
				jsum += p.Y
			}
			return isum > jsum
		})

		// limit the number of datasets
		if len(datasets) > option.maxDatasetNum {
			datasets = datasets[:option.maxDatasetNum]
		}

		for i, d := range datasets {
			r, g, b, _ := Colors[i%len(Colors)].RGBA()
			d.BorderColor = fmt.Sprintf("rgba(%d, %d, %d, 1)", r>>8, g>>8, b>>8)
			d.BackgroundColor = fmt.Sprintf("rgba(%d, %d, %d, 0.2)", r>>8, g>>8, b>>8)
		}
		values.Title2 = "request time sum / sec"
		values.DataSets2 = datasets
	}

	t, err := template.New("chartjs").Parse(`
<!-- show line graph -->

<html>
<head>
    <title>Line Chart</title>
</head>
<body>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.9.3/Chart.min.js"></script>
    <button id="allHide2">hide all</button>
    <button id="allShow2">show all</button>
    <canvas id ='myLineChart2'></canvas>
    <button id="allHide">hide all</button>
    <button id="allShow">show all</button>
    <canvas id ='myLineChart'></canvas>
    <script>
        var ctx = document.getElementById('myLineChart').getContext('2d');
        var ctx2 = document.getElementById('myLineChart2').getContext('2d');
        var myLineChart = new Chart(ctx, {
            type: 'scatter',
            data: {
                datasets: [
					{{range .DataSets}}{
                    label: '{{.Label}}',
					type: 'line',
                    data: [{{range .Data}}{x: {{.X}}, y: {{.Y}}},{{end}}],
                    backgroundColor: '{{.BackgroundColor}}',
                    borderColor: '{{.BorderColor}}',
                    borderWidth: 1,
					fill: false,
                },{{end}}	
					]
            },
            options: {
                scales: {
                    yAxes: [{
                        ticks: {
                            beginAtZero: true
                        }
                    }]
                },
				legend: {
					position: 'left'
				},
				layout: {
					padding: {
						right: 50
					}
				},
                title: {
                    display: true,
					text: '{{.Title}}'
                }
            }
        });

        var myLineChart2 = new Chart(ctx2, {
            type: 'scatter',
            data: {
                datasets: [
					{{range .DataSets2}}{
                    label: '{{.Label}}',
					type: 'line',
                    data: [{{range .Data}}{x: {{.X}}, y: {{.Y}}},{{end}}],
                    backgroundColor: '{{.BackgroundColor}}',
                    borderColor: '{{.BorderColor}}',
                    borderWidth: 1,
					fill: false,
                },{{end}}	
					]
            },
            options: {
                scales: {
                    yAxes: [{
                        ticks: {
                            beginAtZero: true
                        }
                    }]
                },
				legend: {
					position: 'left'
				},
				layout: {
					padding: {
						right: 50
					}
				},
                title: {
                    display: true,
					text: '{{.Title2}}'
                }
            }
        });
    </script>
    <script>
            const hideAllHandlerBuilder = function (chartInstance) {
                return function (e) {
                    chartInstance.data.datasets.forEach((e, i) => {
                        const meta = chartInstance.getDatasetMeta(i);
                        meta.hidden = true;
                    });
                    chartInstance.update();

                }
            }
            const showAllHandlerBuilder = function (chartInstance) {
                return function (e) {
                    chartInstance.data.datasets.forEach((e, i) => {
                        const meta = chartInstance.getDatasetMeta(i);
                        meta.hidden = false;
                    });
                    chartInstance.update();

                }
            }
        document.getElementById('allHide').onclick = hideAllHandlerBuilder(myLineChart);
        document.getElementById('allShow').onclick = showAllHandlerBuilder(myLineChart);
        document.getElementById('allHide2').onclick = hideAllHandlerBuilder(myLineChart2);
        document.getElementById('allShow2').onclick = showAllHandlerBuilder(myLineChart2);
    </script>
</body>
</html>
	`)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, values)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
