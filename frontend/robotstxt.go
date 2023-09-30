package frontend

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetRobotsTxt(c *gin.Context) {

	m := getMetaData(c.Request, DefaultTitle, DefaultDescription)

	buf := new(bytes.Buffer)

	err := templates.ExecuteTemplate(buf, "robots-txt", m)
	if err != nil {
		c.Error(fmt.Errorf("failed to render robots.txt: %w", err))
		Get500(c)
		return
	}

	// Cache for an hour.
	c.Header("cache-control", "public, max-age=3600, stale-while-revalidate=3600, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(3600*time.Second).Format(time.RFC1123))

	c.Data(http.StatusOK, "text/plain; charset=utf-8", bytes.TrimPrefix(buf.Bytes(), []byte{'\n'}))
}
