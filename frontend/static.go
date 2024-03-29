package frontend

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func isImage(ext string) bool {

	switch ext {
	case ".svg":
		return true
	case ".png":
		return true
	case ".gif":
		return true
	case ".ico":
		return true
	default:
		return false
	}

}

func GetStatic(c *gin.Context) {

	if c.Request.Method != "GET" {
		Get404(c)
		return
	}

	content, err := staticFS.ReadFile("static" + c.Request.URL.Path)
	if err != nil {

		c.Error(fmt.Errorf("failed to open %s: %w", c.Request.URL.Path, err))

		if errors.Is(err, os.ErrNotExist) {
			Get404(c)
		} else {
			Get500(c)
		}

		return
	}

	extension := c.Request.URL.Path[strings.LastIndexByte(c.Request.URL.Path, '.'):]

	contentType := mime.TypeByExtension(extension)

	if c.Request.URL.Path == "/site.webmanifest" {
		contentType = "application/json"
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if isImage(extension) {
		// Cache images for a week, stale can be used for +1 day while revalidate and +7 days in case of error
		c.Header("cache-control", "public, max-age=604800, stale-while-revalidate=86400, stale-if-error=604800")
		c.Header("expires", time.Now().UTC().Add(604800*time.Second).Format(time.RFC1123))
	} else {
		// Cache others for a day
		c.Header("cache-control", "public, max-age=86400, stale-while-revalidate=86400, stale-if-error=604800")
		c.Header("expires", time.Now().UTC().Add(86400*time.Second).Format(time.RFC1123))
	}

	c.Data(http.StatusOK, contentType, content)
}
