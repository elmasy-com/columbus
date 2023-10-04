package lookup

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RedirectLookup(c *gin.Context) {

	accept := c.Request.Header.Get("Accept")
	if accept == "" || accept == "*/*" || accept == "application/json" || accept == "text/plain" {
		c.Header("location", fmt.Sprintf("/api/lookup/%s", c.Param("domain")))
		c.Status(http.StatusMovedPermanently)
	} else {
		c.Header("location", fmt.Sprintf("/report/%s", c.Param("domain")))
		c.Status(http.StatusMovedPermanently)
	}
}
