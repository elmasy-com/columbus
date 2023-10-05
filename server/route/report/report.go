package report

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/columbus/frontend"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"github.com/gin-gonic/gin"
)

func getReportDataStat(doms []db.Domain) frontend.ReportStat {

	rs := frontend.ReportStat{}

	rs.Total = len(doms)

	for i := range doms {

		l := len(doms[i].Records)
		rs.TotalRecords += l

		if l > 0 {
			rs.WithRecords++
		}
	}

	rs.WithRecordsPercent = fmt.Sprintf("%.2f", float64(rs.WithRecords)/float64(rs.Total)*100)

	return rs
}

func getReportDataDomains(doms []db.Domain) []frontend.DomainsData {

	dds := make([]frontend.DomainsData, 0, len(doms)/2)

	for i := range doms {

		// Send domains to db.UpdaterChan channel if not full to update the DNS records.
		if len(db.UpdaterChan) < cap(db.UpdaterChan) {
			db.UpdaterChan <- db.UpdateableDomain{Domain: doms[i].String(), Type: db.UpdateExistingDomain}
		}

		dd := frontend.DomainsData{
			Domain: doms[i].String(),
		}

		if doms[i].Updated > 0 {
			dd.Updated = time.Unix(doms[i].Updated, 0).UTC().Format(time.DateTime)
		}

		for ii := range doms[i].Records {

			dd.RecordsNum++

			dd.Records = append(dd.Records, frontend.RecordsData{
				Type:  dns.TypeToString(doms[i].Records[ii].Type),
				Value: doms[i].Records[ii].Value,
				Time:  time.Unix(doms[i].Records[ii].Time, 0).UTC().Format(time.DateTime),
			})
		}

		dds = append(dds, dd)
	}

	return dds
}

func GetReport(c *gin.Context) {

	// Parse domain param
	d := c.Param("domain")

	// Clean domain a
	if d != dns.Clean(d) {
		c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/report/%s", dns.Clean(d)))
		return
	}

	// Redirect client to the base domain if FQDN used (eg.: /search/www.example.com -> /search/example.com)
	if validator.Domain(d) && dns.HasSub(d) {
		c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/report/%s", dns.GetDomain(d)))
		return
	}

	doms, err := db.DomainsDomains(d, -1)
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

	sort.Slice(doms, func(i, j int) bool {

		iParts := strings.Split(doms[i].Sub, ".")
		slices.Reverse(iParts)

		jParts := strings.Split(doms[j].Sub, ".")
		slices.Reverse(jParts)

		for l := 0; l < 255; l++ {

			if l >= len(iParts) {
				return true
			}

			if l >= len(jParts) {
				return false
			}

			if iParts[l] < jParts[l] {
				return true
			}

			if iParts[l] > jParts[l] {
				return false
			}
		}

		return doms[i].Sub < doms[j].Sub
	})

	for l := range doms {

		sort.Slice(doms[l].Records, func(i, j int) bool {

			switch {
			case doms[l].Records[i].Type < doms[l].Records[j].Type:
				return true
			case doms[l].Records[i].Type > doms[l].Records[j].Type:
				return false
			case doms[l].Records[i].Time > doms[l].Records[j].Time:
				return true
			default:
				return false
			}

		})
	}

	_, err = db.TopListInsert(d)
	if err != nil {
		c.Error(fmt.Errorf("failed to insert topList: %w", err))
	}

	reportData := frontend.ReportData{}

	reportData.SubList = buildSubList(doms)
	reportData.Question = d
	reportData.Stat = getReportDataStat(doms)
	reportData.Domains = getReportDataDomains(doms)

	frontend.GetReport(c, reportData)
}
