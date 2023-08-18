package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Redirect(c *gin.Context) {

	c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("/api%s", c.Request.RequestURI))
}

func RedirectLookup(c *gin.Context) {

	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/search/%s", c.Param("domain")))
}
