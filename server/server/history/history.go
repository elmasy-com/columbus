package history

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/gin-gonic/gin"
)

type History struct {
	Domain  string
	Records []db.Record
}

// Return the "days" query parameter.
// If not set, returns -1.
func getQueryDays(c *gin.Context) (int, error) {

	// Parse days query param
	daysStr, daysSet := c.GetQuery("days")
	if !daysSet {
		return -1, nil
	}

	if daysStr == "" {
		return -2, fmt.Errorf("empty")
	}

	return strconv.Atoi(daysStr)
}

func GetApiHistory(c *gin.Context) {

	var err error

	// Parse domain param
	d := c.Param("domain")

	// Parse days query param
	days, err := getQueryDays(c)
	if err != nil {
		c.Error(fault.ErrInvalidDays)
		c.JSON(http.StatusBadRequest, fault.ErrInvalidDays)
		return
	}

	doms, err := db.DomainsDomains(d, days)
	if err != nil {

		c.Error(err)

		respCode := 0

		switch {
		case errors.Is(err, fault.ErrInvalidDomain):
			respCode = http.StatusBadRequest
		case errors.Is(err, fault.ErrInvalidDays):
			respCode = http.StatusBadRequest
		case errors.Is(err, fault.ErrTLDOnly):
			respCode = http.StatusBadRequest
		default:
			respCode = http.StatusInternalServerError
			err = fmt.Errorf("internal server error")
		}

		c.JSON(respCode, gin.H{"error": err.Error()})

		return
	}

	if len(doms) == 0 {

		c.Error(fault.ErrNotFound)

		_, err = db.NotFoundInsert(d)
		if err != nil {
			c.Error(fmt.Errorf("failed to insert notFound: %w", err))
		}

		c.JSON(http.StatusNotFound, fault.ErrNotFound)
		return
	}

	// Send to db.RecordsUpdaterDomainChan if the channle if not full to update the DNS records.
	// Send only if any subdomain found.
	// In db.RecordsUpdaterDomainChan, every record for domain d is updated if not updated in the last hour.
	if len(db.RecordsUpdaterDomainChan) < cap(db.RecordsUpdaterDomainChan) {
		db.RecordsUpdaterDomainChan <- d
	}

	_, err = db.TopListInsert(d)
	if err != nil {
		c.Error(fmt.Errorf("failed to insert topList: %w", err))
	}

	hs := make([]History, 0, len(doms))

	for i := range doms {

		hs = append(hs, History{Domain: doms[i].String(), Records: doms[i].Records})
	}

	c.JSON(http.StatusOK, hs)

}
