package frontend

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

type http500Data struct {
	Meta        metaData
	Hero        heroData
	GitHubIssue string
}

var DefaultGitHubIssue = "?title=Internal%%20Server%%20Error%%20on%%20%s&body=Internal%%20Server%%20Error%%20on%%20%%60%s%%60%%20%%2E"

func Get500Text(c *gin.Context) {
	c.String(http.StatusInternalServerError, "internal server error")
}

func Get500JSON(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

// Get500 set context.
// If failed to render the 500.html, returns a string.
func Get500HTML(c *gin.Context) {

	buf := new(bytes.Buffer)
	dat := http500Data{
		Meta: getMetaData(c.Request, "Columbus Project - 500 Internal Server Error", DefaultDescription),
		Hero: getHeroData("500 Internal Server Error", c.Request.Method+" "+c.Request.URL.Path),
		GitHubIssue: "https://github.com/elmasy-com/columbus/issues/new?" +
			"title=" + template.URLQueryEscaper(fmt.Sprintf("Internal Server Error on %s", c.Request.URL.Path)) +
			"&" +
			"body=" + template.URLQueryEscaper(fmt.Sprintf("Internal Server Error on %s.", c.Request.URL.Path)),
	}

	err := templates.ExecuteTemplate(buf, "500", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render 500: %w", err))
		Get500Text(c)
		return
	}

	c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", buf.Bytes())
}

func Get500(c *gin.Context) {

	switch c.GetHeader("Accept") {
	case "":
		Get500JSON(c)
	case "*/*":
		Get500JSON(c)
	case "text/plain":
		Get500Text(c)
	case "application/json":
		Get500JSON(c)
	default:
		Get500HTML(c)
	}
}
