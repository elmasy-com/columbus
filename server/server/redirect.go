package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Redirect(c *gin.Context) {

	c.Header("location", fmt.Sprintf("/api%s", c.Request.RequestURI))
	c.Status(http.StatusMovedPermanently)
}

func RedirectLookup(c *gin.Context) {

	accept := c.Request.Header.Get("Accept")
	if accept == "" || accept == "*/*" || accept == "application/json" || accept == "text/plain" {
		c.Header("location", fmt.Sprintf("/api/lookup/%s", c.Param("domain")))
		c.Status(http.StatusMovedPermanently)
	} else {
		c.Header("location", fmt.Sprintf("/search/%s", c.Param("domain")))
		c.Status(http.StatusMovedPermanently)
	}
}

func RedirectSwagger(c *gin.Context) {

	c.Header("location", "https://columbus.elmasy.com/swagger/index.html")
	c.Status(http.StatusTemporaryRedirect)
}
