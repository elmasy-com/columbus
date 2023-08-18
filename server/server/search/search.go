package search

import (
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
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

type SearchData struct {
	Question string
	Domains  []DomainsData
	Unknowns []string
	Error    error
}

//go:embed search.html
var searchHtml string

//go:embed searchResult.html
var searchResultHtml string

//go:embed searchBadRequest.html
var searchBadrequestHtml string

//go:embed searchInternalServerError.html
var searchInternalServerErrorHtml string

//go:embed searchNotFound.html
var searchNotFoundHtml string

func GetSearch(c *gin.Context) {

	c.Data(http.StatusOK, "text/html", []byte(searchHtml))
}

func GetSearchResult(c *gin.Context) {

	var err error
	var doms []string

	// Parse domain param
	d := c.Param("domain")

	doms, err = db.DomainsLookupFull(d, -1)
	if err != nil {

		c.Error(fmt.Errorf("fail to lookup full: %w", err))

		switch {
		case errors.Is(err, fault.ErrInvalidDomain):
			c.Data(http.StatusBadRequest, "text/html", []byte(searchBadrequestHtml))
		case errors.Is(err, fault.ErrInvalidDays):
			c.Data(http.StatusBadRequest, "text/html", []byte(searchBadrequestHtml))
		default:
			c.Data(http.StatusInternalServerError, "text/html", []byte(searchInternalServerErrorHtml))
		}

		return
	}

	if len(doms) == 0 {

		_, err = db.NotFoundInsert(d)
		if err != nil {
			c.Error(fmt.Errorf("failed to insert notFound: %w", err))
		}

		c.Data(http.StatusNotFound, "text/html", []byte(searchNotFoundHtml))
		return

	}

	_, err = db.TopListInsert(d)
	if err != nil {
		c.Error(fmt.Errorf("failed to insert topList: %w", err))
	}

	// Send to db.RecordsUpdaterDomainChan if the channle if not full to update the DNS records.
	// Send only if any subdomain found.
	// In db.RecordsUpdaterDomainChan, every record for domain d is updated if not updated in the last hour.

	if len(db.RecordsUpdaterDomainChan) < cap(db.RecordsUpdaterDomainChan) {
		db.RecordsUpdaterDomainChan <- d
	}

	searchData := SearchData{Question: d}

	for i := range doms {

		rs, err := db.DomainsRecords(doms[i], 0)
		if err != nil {

			c.Error(fmt.Errorf("fail to get record for %s: %w", doms[i], err))

			switch {
			case errors.Is(err, fault.ErrInvalidDomain):
				c.Data(http.StatusBadRequest, "text/html", []byte(searchBadrequestHtml))
			case errors.Is(err, fault.ErrInvalidDays):
				c.Data(http.StatusBadRequest, "text/html", []byte(searchBadrequestHtml))
			default:
				c.Data(http.StatusInternalServerError, "text/html", []byte(searchInternalServerErrorHtml))
			}

			return
		}

		if len(rs) == 0 {
			searchData.Unknowns = append(searchData.Unknowns, doms[i])
			continue
		}

		v := DomainsData{Domain: doms[i]}

		for ii := range rs {
			v.Records = append(v.Records, RecordsData{Type: dns.TypeToString(rs[ii].Type), Value: rs[ii].Value, Time: time.Unix(rs[ii].Time, 0).String()})
		}

		sort.Slice(v.Records, func(i, j int) bool { return v.Records[i].Time > v.Records[j].Time })

		searchData.Domains = append(searchData.Domains, v)
	}

	t := template.New("searchResult")

	t, err = t.Parse(searchResultHtml)
	if err != nil {
		c.Error(fmt.Errorf("failed to parse template: %w", err))
		c.Data(http.StatusInternalServerError, "text/html", []byte(searchInternalServerErrorHtml))
		return
	}

	err = t.Execute(c.Writer, searchData)
	if err != nil {
		c.Error(fmt.Errorf("failed to execute template: %w", err))
		c.Data(http.StatusInternalServerError, "text/html", []byte(searchInternalServerErrorHtml))
		return
	}
}
