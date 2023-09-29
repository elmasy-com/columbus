package frontend

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type http502Data struct {
	Meta metaData
	Hero heroData
}

func Get502Text(c *gin.Context) {
	c.String(http.StatusBadGateway, "bad gateway")
}

func Get502JSON(c *gin.Context) {
	c.JSON(http.StatusBadGateway, gin.H{"error": "bad gateway"})
}

// Get502HTML
// If failed to render the 500.html, returns a string.
func Get502HTML(c *gin.Context) {

	buf := new(bytes.Buffer)
	dat := http502Data{
		Meta: getMetaData(c.Request, "Columbus Project - 502 Bad Gateway", DefaultDescription),
		Hero: getHeroData("502 Bad Gateway", "The servers are down!"),
	}

	err := templates.ExecuteTemplate(buf, "502", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render 502: %w", err))
		Get500(c)
		return
	}

	c.Data(http.StatusBadGateway, "text/html; charset=utf-8", buf.Bytes())
}

func Get502(c *gin.Context) {

	switch c.GetHeader("Accept") {
	case "":
		Get502JSON(c)
	case "*/*":
		Get502JSON(c)
	case "text/plain":
		Get502Text(c)
	case "application/json":
		Get502JSON(c)
	default:
		Get502HTML(c)
	}
}
