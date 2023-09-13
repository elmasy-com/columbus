package frontend

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"

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
		// Cache images for a week
		c.Header("X-Accel-Expires", "604800")
	} else {
		// Cache static files for a day
		c.Header("X-Accel-Expires", "86400")
	}

	c.Data(http.StatusOK, contentType, content)
}
