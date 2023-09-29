package statistics

import (
	"bytes"
	"fmt"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/go-echarts/go-echarts/v2/types"
)

func CreateHistoryChart() ([]byte, error) {

	datas, err := db.StatisticsGets()
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// X Axis - Date
	dates := make([]time.Time, 0, len(datas))

	// Y Axis - numbers
	totals := make([]opts.LineData, 0, len(datas))
	updated := make([]opts.LineData, 0, len(datas))
	valids := make([]opts.LineData, 0, len(datas))
	ctlogs := make([]opts.LineData, 0, len(datas))

	for i := range datas {

		dates = append(dates, time.Unix(datas[i].Date, 0).UTC())

		totals = append(totals, opts.LineData{Value: datas[i].Total})

		updated = append(updated, opts.LineData{Value: datas[i].Updated})

		valids = append(valids, opts.LineData{Value: datas[i].Valid})

		var v int64

		for l := range datas[i].CTLogs {
			v += datas[i].CTLogs[l].Size
		}

		ctlogs = append(ctlogs, opts.LineData{Value: v})

	}

	line := charts.NewLine()

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:      "1000px",
			AssetsHost: "/",
			Theme:      types.ThemeInfographic,
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "axis"}),
	)

	line.SetXAxis(dates)

	line.AddSeries("Totals", totals)
	line.AddSeries("Updated", updated)
	line.AddSeries("Valid", valids)
	line.AddSeries("CT Logs", ctlogs)

	line.SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{
			Color: "#DADADB",
		}),
	)

	buf := new(bytes.Buffer)

	renderer := render.NewChartRender(line, line.Validate)

	err = renderer.Render(buf)

	return buf.Bytes(), err
}
