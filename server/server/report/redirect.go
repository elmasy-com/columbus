package report

import (
	"fmt"
	"net/http"

	"github.com/elmasy-com/columbus/fault"
	"github.com/elmasy-com/columbus/frontend"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"github.com/gin-gonic/gin"
)

// Redirect queries /report?domain=example.com
func RedirectDomainParam(c *gin.Context) {

	dom := c.Query("domain")

	if dom == "" {
		frontend.Get400HTML(c, fmt.Errorf("domain is empty"))
	}

	if !validator.Domain(dom) {
		frontend.Get400HTML(c, fault.ErrInvalidDomain)
		return
	}

	dom = dns.Clean(dom)

	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/report/%s", dns.GetDomain(dom)))
}
