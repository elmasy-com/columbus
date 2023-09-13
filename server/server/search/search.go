package search

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/columbus/frontend"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"github.com/gin-gonic/gin"
)

func GetSearchRedirect(c *gin.Context) {

	c.Redirect(http.StatusFound, "/#search")
}

func GetSearch(c *gin.Context) {

	var err error
	var doms []string

	// Parse domain param
	d := dns.Clean(c.Param("domain"))

	// Redirect client to the base domain if FQDN used (eg.: /search/www.example.com -> /search/example.com)
	if validator.Domain(d) && dns.HasSub(d) {
		c.Header("location", fmt.Sprintf("/search/%s", dns.GetDomain(d)))
		c.Status(http.StatusFound)
		return
	}

	doms, err = db.DomainsLookupFull(d, -1)
	if err != nil {

		c.Error(fmt.Errorf("fail to lookup full: %w", err))

		switch {
		case errors.Is(err, fault.ErrInvalidDomain):
			frontend.Get400HTML(c, fault.ErrInvalidDomain)
		case errors.Is(err, fault.ErrInvalidDays):
			frontend.Get400HTML(c, fault.ErrInvalidDays)
		case errors.Is(err, fault.ErrTLDOnly):
			frontend.Get400HTML(c, fault.ErrTLDOnly)
		default:
			frontend.Get500HTML(c)
		}

		return
	}

	if len(doms) == 0 {

		_, err = db.NotFoundInsert(d)
		if err != nil {
			c.Error(fmt.Errorf("failed to insert notFound: %w", err))
		}

		frontend.Get404HTML(c, d)
		return

	}

	_, err = db.TopListInsert(d)
	if err != nil {
		c.Error(fmt.Errorf("failed to insert topList: %w", err))
	}

	searchData := frontend.SearchData{Question: d}

	for i := range doms {

		rs, err := db.DomainsRecords(doms[i], 0)
		if err != nil {

			c.Error(fmt.Errorf("fail to get record for %s: %w", doms[i], err))

			switch {
			case errors.Is(err, fault.ErrInvalidDomain):
				frontend.Get400HTML(c, fault.ErrInvalidDomain)
			case errors.Is(err, fault.ErrInvalidDays):
				frontend.Get400HTML(c, fault.ErrInvalidDays)
			case errors.Is(err, fault.ErrTLDOnly):
				frontend.Get400HTML(c, fault.ErrTLDOnly)
			default:
				frontend.Get500HTML(c)
			}

			return
		}

		if len(rs) == 0 {
			searchData.Unknowns = append(searchData.Unknowns, doms[i])
			continue
		}

		v := frontend.DomainsData{Domain: doms[i]}

		for ii := range rs {
			v.Records = append(v.Records, frontend.RecordsData{Type: dns.TypeToString(rs[ii].Type), Value: rs[ii].Value, Time: time.Unix(rs[ii].Time, 0).String()})
		}

		sort.Slice(v.Records, func(i, j int) bool {

			if v.Records[i].Type != v.Records[j].Type {
				return v.Records[i].Type < v.Records[j].Type
			}

			return v.Records[i].Time > v.Records[j].Time
		})

		searchData.Domains = append(searchData.Domains, v)
	}

	searchData.Stat.Total = strconv.Itoa(len(doms))
	searchData.Stat.WithRecords = strconv.Itoa(len(searchData.Domains))
	searchData.Stat.WithRecordsPercent = fmt.Sprintf("%.2f", float64(len(searchData.Domains))/float64(len(doms))*100.0)

	totalrecords := 0

	for i := range searchData.Domains {
		totalrecords += len(searchData.Domains[i].Records)
	}

	searchData.Stat.TotalRecords = strconv.Itoa(totalrecords)

	frontend.GetSearchHtml(c, searchData)
}
