package frontend

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/elmasy-com/columbus/fault"
	"github.com/gin-gonic/gin"
)

type http404Data struct {
	Meta   metaData
	Hero   heroData
	Method string
	Domain string
}

func Get404Text(c *gin.Context) {
	c.String(http.StatusNotFound, fault.ErrNotFound.Error())
}

func Get404JSON(c *gin.Context) {
	c.JSON(http.StatusNotFound, fault.ErrNotFound)
}

func Get404HTML(c *gin.Context, d string) {

	buf := new(bytes.Buffer)

	title := "404 Not Found"
	subtitle := c.Request.Method + " " + c.Request.URL.Path
	if d != "" {
		subtitle = "No subdomain for " + d
	}

	dat := http404Data{
		Meta:   getMetaData(c.Request, "Columbus Project - "+title, DefaultDescription),
		Hero:   getHeroData("404 Not Found", subtitle),
		Method: c.Request.Method,
		Domain: d,
	}

	err := templates.ExecuteTemplate(buf, "404", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render 404: %w", err))
		Get500(c)
		return
	}

	c.Data(http.StatusNotFound, "text/html; charset=utf-8", buf.Bytes())
}

// Get404 set context.
// If failed to render the 404.html, returns code 500 with a string: "Internal Server Error".
//
// If d is set to a domain, the 404.html will show "No subdomains found for...".
func Get404(c *gin.Context) {

	switch c.GetHeader("Accept") {
	case "":
		Get404JSON(c)
	case "*/*":
		Get404JSON(c)
	case "text/plain":
		Get404Text(c)
	case "application/json":
		Get404JSON(c)
	default:
		Get404HTML(c, "")
	}
}
