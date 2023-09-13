package frontend

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RecordsData struct {
	Type  string
	Value string
	Time  string
}

type DomainsData struct {
	Domain  string
	Records []RecordsData
}

type SearchStat struct {
	Total              string
	WithRecords        string
	WithRecordsPercent string
	TotalRecords       string
}

type SearchData struct {
	Meta     metaData
	Question string
	Stat     SearchStat
	Domains  []DomainsData
	Unknowns []string
	Error    error
}

// Get404 set context.
// If failed to render the 404.html, returns code 500 with a string: "Internal Server Error".
//
// If d is set to a domain, the 404.html will show "No subdomains found for...".
func GetSearchHtml(c *gin.Context, dat SearchData) {

	buf := new(bytes.Buffer)

	dat.Meta = getMetaData(c.Request, "Columbus Project - Subdomains of "+dat.Question, DescriptionLong)

	err := templates.ExecuteTemplate(buf, "search", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render search.html: %w", err))
		Get500(c)
		return
	}

	// Cache for an hour.
	c.Header("cache-control", "public, max-age=3600")
	c.Header("expires", time.Now().UTC().Add(3600*time.Second).Format(time.RFC1123))

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
