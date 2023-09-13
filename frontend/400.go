package frontend

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type http400Data struct {
	Meta   metaData
	Method string
	Err    string
}

func Get400Code(c *gin.Context) {
	c.Status(http.StatusBadRequest)
}

func Get400Text(c *gin.Context, err error) {
	c.String(http.StatusBadRequest, err.Error())
}

func Get400JSON(c *gin.Context, err error) {
	c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
}

func Get400HTML(c *gin.Context, err error) {

	buf := new(bytes.Buffer)

	ds := "Columbus Project - 400 Bad Request"

	dat := http400Data{Meta: getMetaData(c.Request, ds, DescriptionLong), Method: c.Request.Method, Err: err.Error()}

	if err = templates.ExecuteTemplate(buf, "400", dat); err != nil {
		c.Error(fmt.Errorf("failed to render 400.html: %w", err))
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

	Get400HTML(c, fmt.Errorf("bad request"))
}
