package frontend

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/gin-gonic/gin"
)

type ctLog struct {
	Name      string
	Index     int64
	Size      int64
	Remaining int64
	Progress  float64
}

type historyStat struct {
	Num        int64
	Date       int64
	Total      int64
	Updated    int64
	Valid      int64
	CTLogTotal int64
}

type statistics struct {
	Date             int64
	Total            int64
	Updated          int64
	UpdatedPercent   float64
	Valid            int64
	ValidPercent     float64
	CTTotalIndex     int64
	CTTotalSize      int64
	CTTotalRemaining int64
	CTTotalProgress  float64
	CTLogs           []ctLog
	History          []historyStat
	HistoryChart     template.HTML
}

type statisticsData struct {
	Meta       metaData
	Hero       heroData
	Statistics statistics
}

func parseStatistic() (statistics, error) {

	s, err := db.StatisticsGets()
	if err != nil {
		return statistics{}, fmt.Errorf("failed to get newset statistic: %w", err)
	}

	if len(s) == 0 {
		return statistics{}, nil
	}

	var stat statistics

	// The first element in the slice is the newest entry
	stat.Date = s[0].Date
	stat.Total = s[0].Total
	stat.Updated = s[0].Updated
	stat.UpdatedPercent = float64(s[0].Updated) / float64(s[0].Total) * 100
	stat.Valid = s[0].Valid
	stat.ValidPercent = float64(s[0].Valid) / float64(s[0].Total) * 100

	stat.CTLogs = make([]ctLog, len(s[0].CTLogs))

	for i := range s[0].CTLogs {
		stat.CTLogs[i].Name = s[0].CTLogs[i].Name
		stat.CTLogs[i].Index = s[0].CTLogs[i].Index
		stat.CTLogs[i].Size = s[0].CTLogs[i].Size
		stat.CTLogs[i].Remaining = s[0].CTLogs[i].Size - s[0].CTLogs[i].Index
		stat.CTLogs[i].Progress = float64(s[0].CTLogs[i].Index) / float64(s[0].CTLogs[i].Size) * 100

		stat.CTTotalIndex += s[0].CTLogs[i].Index
		stat.CTTotalSize += s[0].CTLogs[i].Size
		stat.CTTotalRemaining += s[0].CTLogs[i].Size - s[0].CTLogs[i].Index

	}

	stat.CTTotalProgress = float64(stat.CTTotalIndex) / float64(stat.CTTotalSize) * 100

	sort.Slice(stat.CTLogs, func(i, j int) bool { return stat.CTLogs[i].Progress > stat.CTLogs[j].Progress })

	// The remaining elements are the history
	hs := s[1:]

	var hNum int64 = 1

	stat.History = make([]historyStat, len(hs))

	for i := range hs {

		stat.History[i].Num = hNum
		stat.History[i].Date = hs[i].Date

		stat.History[i].Total = hs[i].Total
		stat.History[i].Updated = hs[i].Updated
		stat.History[i].Valid = hs[i].Valid

		for ii := range hs[i].CTLogs {
			stat.History[i].CTLogTotal += hs[i].CTLogs[ii].Index
		}

		hNum++
	}

	//hc, err := serverstatistics.CreateHistoryChart()

	//stat.HistoryChart = template.HTML(hc)

	return stat, err
}

func GetStatistics(c *gin.Context) {

	statistics, err := parseStatistic()
	if err != nil {
		c.Error(fmt.Errorf("failed to parse statistics: %w", err))
		Get500(c)
		return
	}

	dat := statisticsData{
		Meta:       getMetaData(c.Request, "Columbus Project - Statistics", DefaultDescription),
		Hero:       getHeroData("Statistics", "Database and Scanners statistics at "+time.Unix(statistics.Date, 0).UTC().Format(time.DateTime)),
		Statistics: statistics,
	}

	buf := new(bytes.Buffer)

	err = templates.ExecuteTemplate(buf, "statistics", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render statistics: %w", err))
		Get500(c)
		return
	}

	// Cache for an hour.
	c.Header("cache-control", "public, max-age=3600")
	c.Header("expires", time.Now().UTC().Add(3600*time.Second).Format(time.RFC1123))

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func RedirectStatToStatistics(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "statistics")
}
