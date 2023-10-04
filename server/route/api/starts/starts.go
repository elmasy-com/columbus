package starts

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/gin-gonic/gin"
)

func GetApiStarts(c *gin.Context) {

	dom := c.Param("domain")

	if len(dom) < 5 {

		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusBadRequest, fault.ErrInvalidDomain.Error())
		} else {
			c.JSON(http.StatusBadRequest, fault.ErrInvalidDomain)
		}
		return
	}

	domains, err := db.DomainsStarts(dom)
	if err != nil {

		c.Error(err)
		code := 0

		if errors.Is(err, fault.ErrInvalidDomain) {
			code = http.StatusBadRequest
		} else {
			code = http.StatusInternalServerError
			err = fmt.Errorf("internal server error")
		}

		if c.GetHeader("Accept") == "text/plain" {
			c.String(code, err.Error())
		} else {
			c.JSON(code, err)
		}
		return
	}

	if len(domains) == 0 {

		c.Error(fault.ErrNotFound)

		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusNotFound, fault.ErrNotFound.Err)
		} else {
			c.JSON(http.StatusNotFound, fault.ErrNotFound)
		}
		return
	}

	// Cache for 10 minutes.
	c.Header("cache-control", "public, max-age=600, must-revalidate, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(600*time.Second).Format(time.RFC1123))
	c.Header("vary", "Accept")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, strings.Join(domains, "\n"))
	} else {
		c.JSON(http.StatusOK, domains)
	}
}
