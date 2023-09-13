package lookup

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
	"github.com/gin-gonic/gin"
)

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

func GetApiLookup(c *gin.Context) {

	var err error

	// Parse domain param
	d := c.Param("domain")

	// Parse days query param
	days, err := getQueryDays(c)
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
	c.Header("cache-control", "public, max-age=600")
	c.Header("expires", time.Now().UTC().Add(600*time.Second).Format(time.RFC1123))
	c.Header("vary", "Accept")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, strings.Join(subs, "\n"))
	} else {
		c.JSON(http.StatusOK, subs)
	}
}

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
	c.Header("cache-control", "public, max-age=600")
	c.Header("expires", time.Now().UTC().Add(600*time.Second).Format(time.RFC1123))
	c.Header("vary", "Accept")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, strings.Join(tlds, "\n"))
	} else {
		c.JSON(http.StatusOK, tlds)
	}
}

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
	c.Header("cache-control", "public, max-age=600")
	c.Header("expires", time.Now().UTC().Add(600*time.Second).Format(time.RFC1123))
	c.Header("vary", "Accept")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, strings.Join(domains, "\n"))
	} else {
		c.JSON(http.StatusOK, domains)
	}
}
