package tld

import (
	"net/http"
	"strings"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
	"github.com/gin-gonic/gin"
)

func GetApiTLD(c *gin.Context) {

	dom := c.Param("domain")

	if !dns.IsValidSLD(dom) {

		c.Error(fault.ErrInvalidDomain)

		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusBadRequest, fault.ErrInvalidDomain.Error())
		} else {
			c.JSON(http.StatusBadRequest, fault.ErrInvalidDomain)
		}
		return
	}

	dom = dns.Clean(dom)

	tlds, err := db.DomainsTLD(dom)
	if err != nil {

		c.Error(err)

		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusInternalServerError, "internal server error")
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	if len(tlds) == 0 {
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
		c.String(http.StatusOK, strings.Join(tlds, "\n"))
	} else {
		c.JSON(http.StatusOK, tlds)
	}
}
