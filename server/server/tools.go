package server

import (
	"fmt"
	"net/http"

	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/elnet/dns"
	"github.com/gin-gonic/gin"
)

// GET /tools/tld/{fqdn}
// Returns the TLD part of a FQDN.
func ToolsTLDGet(c *gin.Context) {

	fqdn := c.Param("fqdn")

	if !dns.IsValid(fqdn) || fqdn == "." {
		c.Error(fault.ErrInvalidDomain)
		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusBadRequest, fault.ErrInvalidDomain.Err)
		} else {
			c.JSON(http.StatusBadRequest, fault.ErrInvalidDomain)
		}
		return
	}

	fqdn = dns.Clean(fqdn)

	d := dns.GetTLD(fqdn)
	if d == "" {
		c.Error(fault.ErrNotFound)
		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusNotFound, fault.ErrInvalidDomain.Err)
		} else {
			c.JSON(http.StatusNotFound, fault.ErrInvalidDomain)
		}
		return
	}

	c.Header("X-Accel-Expires", "600")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, d)
	} else {
		c.JSON(http.StatusOK, gin.H{"result": d})
	}
}

// GET /tools/domain/{fqdn}
// Returns the domain part of a FQDN.
func ToolsDomainGet(c *gin.Context) {

	fqdn := c.Param("fqdn")

	if !dns.IsValid(fqdn) || fqdn == "." {
		c.Error(fault.ErrInvalidDomain)
		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusBadRequest, fault.ErrInvalidDomain.Err)
		} else {
			c.JSON(http.StatusBadRequest, fault.ErrInvalidDomain)
		}
		return
	}

	fqdn = dns.Clean(fqdn)

	d := dns.GetDomain(fqdn)
	if d == "" {
		c.Error(fault.ErrNotFound)
		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusNotFound, fault.ErrInvalidDomain.Err)
		} else {
			c.JSON(http.StatusNotFound, fault.ErrInvalidDomain)
		}
		return
	}

	c.Header("X-Accel-Expires", "600")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, d)
	} else {
		c.JSON(http.StatusOK, gin.H{"result": d})
	}
}

// GET /tools/subdomain/{fqdn}
// Returns the subdomain part of a FQDN.
func ToolsSubdomainGet(c *gin.Context) {

	fqdn := c.Param("fqdn")

	if !dns.IsValid(fqdn) || fqdn == "." {
		c.Error(fault.ErrInvalidDomain)
		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusBadRequest, fault.ErrInvalidDomain.Err)
		} else {
			c.JSON(http.StatusBadRequest, fault.ErrInvalidDomain)
		}
		return
	}

	fqdn = dns.Clean(fqdn)

	d := dns.GetSub(fqdn)
	if d == "" {
		c.Error(fault.ErrNotFound)
		if c.GetHeader("Accept") == "text/plain" {
			c.String(http.StatusNotFound, fault.ErrInvalidDomain.Err)
		} else {
			c.JSON(http.StatusNotFound, fault.ErrInvalidDomain)
		}
		return
	}

	c.Header("X-Accel-Expires", "600")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, d)
	} else {
		c.JSON(http.StatusOK, gin.H{"result": d})
	}
}

// GET /tools/isvalid/{fqdn}
// Returns wether fqdn is valid.
func ToolsIsValidGet(c *gin.Context) {

	fqdn := c.Param("fqdn")
	fqdn = dns.Clean(fqdn)

	c.Header("X-Accel-Expires", "600")

	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, fmt.Sprintf("%v", dns.IsValid(fqdn)))
	} else {
		c.JSON(http.StatusOK, gin.H{"result": dns.IsValid(fqdn)})
	}
}
