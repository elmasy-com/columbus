package frontend

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type http504Data struct {
	Meta metaData
	Hero heroData
}

func Get504Text(c *gin.Context) {
	c.String(http.StatusGatewayTimeout, "gateway timeout")
}

func Get504JSON(c *gin.Context) {
	c.JSON(http.StatusGatewayTimeout, gin.H{"error": "gateway timeout"})
}

// Get504HTML
// If failed to render the 500.html, returns a string.
func Get504HTML(c *gin.Context) {

	buf := new(bytes.Buffer)

	dat := http504Data{
		Meta: getMetaData(c.Request, "Columbus Project - 504 Gateway Timeout", DefaultDescription),
		Hero: getHeroData("504 Gateway Timeout", c.Request.Method+" "+c.Request.URL.Path+" took too much time!"),
	}

	err := templates.ExecuteTemplate(buf, "504", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render 504: %w", err))
		Get500(c)
		return
	}

	c.Data(http.StatusGatewayTimeout, "text/html; charset=utf-8", buf.Bytes())
}

func Get504(c *gin.Context) {

	switch c.GetHeader("Accept") {
	case "":
		Get504JSON(c)
	case "*/*":
		Get504JSON(c)
	case "text/plain":
		Get504Text(c)
	case "application/json":
		Get504JSON(c)
	default:
		Get504HTML(c)
	}
}
