package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RedirectOldRoutes(c *gin.Context) {

	c.Header("location", fmt.Sprintf("/api%s", c.Request.RequestURI))
	c.Status(http.StatusMovedPermanently)
}
