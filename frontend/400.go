package frontend

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type http400Data struct {
	Meta metaData
	Hero heroData
	Err  string
}

func Get400Text(c *gin.Context, err error) {
	c.String(http.StatusBadRequest, err.Error())
}

func Get400JSON(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func Get400HTML(c *gin.Context, err error) {

	buf := new(bytes.Buffer)

	title := "400 Bad Request"
	subtitle := c.Request.Method + " " + c.Request.URL.Path

	dat := http400Data{
		Meta: getMetaData(c.Request, "Columbus Project - "+title, DefaultDescription),
		Hero: getHeroData(title, subtitle),
		Err:  err.Error(),
	}

	if err = templates.ExecuteTemplate(buf, "400", dat); err != nil {
		c.Error(fmt.Errorf("failed to render 400: %w", err))
		Get500(c)
		return
	}

	c.Data(http.StatusBadRequest, "text/html; charset=utf-8", buf.Bytes())
}

// Get400 set context.
// If failed to render the 404.html, returns code 500 with a string: "Internal Server Error".
//
// If d is set to a domain, the 404.html will show "No subdomains found for...".
func Get400(c *gin.Context) {

	switch c.GetHeader("Accept") {
	case "":
		Get400JSON(c, fmt.Errorf("bad request"))
	case "*/*":
		Get400JSON(c, fmt.Errorf("bad request"))
	case "text/plain":
		Get400Text(c, fmt.Errorf("bad request"))
	case "application/json":
		Get400JSON(c, fmt.Errorf("bad request"))
	default:
		Get400HTML(c, fmt.Errorf("bad request"))
	}
}
