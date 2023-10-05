package history

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/columbus/server/common"
	"github.com/gin-gonic/gin"
)

type History struct {
	Domain  string
	Records []db.Record
}

func GetApiHistory(c *gin.Context) {

	var err error

	// Parse domain param
	d := c.Param("domain")

	// Parse days query param
	days, err := common.ParseQueryDays(c)
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

	_, err = db.TopListInsert(d)
	if err != nil {
		c.Error(fmt.Errorf("failed to insert topList: %w", err))
	}

	hs := make([]History, 0, len(doms))

	for i := range doms {

		// Send domains to db.UpdaterChan channel if not full to update the DNS records.
		if len(db.UpdaterChan) < cap(db.UpdaterChan) {
			db.UpdaterChan <- db.UpdateableDomain{Domain: doms[i].String(), Type: db.UpdateExistingDomain}
		}

		hs = append(hs, History{Domain: doms[i].String(), Records: doms[i].Records})
	}

	// Cache for 10 minutes, domains are not updated this often,
	// but caching saves a lot of processing power.
	c.Header("cache-control", "public, max-age=600, must-revalidate, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(600*time.Second).Format(time.RFC1123))

	c.JSON(http.StatusOK, hs)

}
