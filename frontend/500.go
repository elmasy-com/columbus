package frontend

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type http500Data struct {
	Meta metaData
	Date string
}

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
	dat := http500Data{Meta: getMetaData(c.Request, "Columbus Project - Internal Server Error", DescriptionLong), Date: time.Now().Format("2006-01-02:15:04:05")}

	err := templates.ExecuteTemplate(buf, "500", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render 500.html: %w", err))
		c.String(http.StatusInternalServerError, "internal server error")
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
