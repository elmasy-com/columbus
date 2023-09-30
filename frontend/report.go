package frontend

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SubData struct {
	Sub    string
	Domain string
	Childs []*SubData
}

func (s *SubData) Add(sub string, domain string) *SubData {

	for i := range s.Childs {
		if s.Childs[i].Sub == sub {
			return s.Childs[i]
		}
	}

	n := new(SubData)
	n.Sub = sub
	n.Domain = domain
	n.Childs = make([]*SubData, 0)

	s.Childs = append(s.Childs, n)

	return n
}

type RecordsData struct {
	Type  string
	Value string
	Time  string
}

type DomainsData struct {
	Domain     string
	Updated    string
	RecordsNum int
	Records    []RecordsData
}

type ReportStat struct {
	Total              int
	WithRecords        int
	WithRecordsPercent string
	TotalRecords       int
}

type ReportData struct {
	Meta     metaData
	Question string
	SubList  *SubData
	Stat     ReportStat
	Domains  []DomainsData
}

// Get404 set context.
// If failed to render the 404.html, returns code 500 with a string: "Internal Server Error".
//
// If d is set to a domain, the 404.html will show "No subdomains found for...".
func GetReport(c *gin.Context, dat ReportData) {

	buf := new(bytes.Buffer)

	dat.Meta = getMetaData(c.Request, "Columbus Project - Subdomains of "+dat.Question, DefaultDescription)

	err := templates.ExecuteTemplate(buf, "report", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render report: %w", err))
		Get500(c)
		return
	}

	// Cache for an hour.
	c.Header("cache-control", "public, max-age=3600, stale-while-revalidate=86400, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(3600*time.Second).Format(time.RFC1123))

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
