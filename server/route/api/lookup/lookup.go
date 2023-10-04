package lookup

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/columbus/server/common"
	"github.com/gin-gonic/gin"
)

func GetApiLookup(c *gin.Context) {

	var err error

	// Parse domain param
	d := c.Param("domain")

	// Parse days query param
	days, err := common.ParseQueryDays(c)
	if err != nil {
		c.Error(err)
		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusBadRequest, fault.ErrInvalidDays.Err)
		} else {
			c.JSON(http.StatusBadRequest, fault.ErrInvalidDays)
		}
		return
	}

	subs, err := db.DomainsLookup(d, days)
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

		if c.GetHeader("Accept") == "text/plain" {
			c.String(respCode, err.Error())
		} else {
			c.JSON(respCode, gin.H{"error": err.Error()})
		}
		return
	}

	if len(subs) == 0 {

		c.Error(fault.ErrNotFound)

		_, err = db.NotFoundInsert(d)
		if err != nil {
			c.Error(fmt.Errorf("failed to insert notFound: %w", err))
		}

		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusNotFound, fault.ErrNotFound.Err)
		} else {
			c.JSON(http.StatusNotFound, fault.ErrNotFound)
		}
		return
	}

	_, err = db.TopListInsert(d)
	if err != nil {
		c.Error(fmt.Errorf("failed to insert topList: %w", err))
	}

	// Cache for 10 minutes.
	c.Header("cache-control", "public, max-age=600, must-revalidate, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(600*time.Second).Format(time.RFC1123))
	c.Header("vary", "Accept")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, strings.Join(subs, "\n"))
	} else {
		c.JSON(http.StatusOK, subs)
	}
}
