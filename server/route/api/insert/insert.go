package insert

import (
	"fmt"
	"net/http"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/server/config"
	"github.com/elmasy-com/elnet/dns"
	"github.com/elmasy-com/elnet/validator"
	"github.com/gin-gonic/gin"
)

func PutApiInsert(c *gin.Context) {

	if config.Blocklist.IsBlocked(c.ClientIP()) {
		c.Status(http.StatusForbidden)
		return
	}

	d := dns.Clean(c.Param("domain"))

	if !validator.Domain(d) {
		config.Blocklist.Block(c.ClientIP())
		c.Error(fmt.Errorf("invalid domain: %s", d))
		c.Status(http.StatusBadRequest)
		return
	}

	db.UpdaterChan <- db.UpdateableDomain{Domain: d, Type: db.InsertNewDomain}

	c.Status(http.StatusOK)
}
