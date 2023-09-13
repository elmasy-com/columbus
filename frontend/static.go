package frontend

import (
	"crypto/md5"
	"encoding/hex"
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

	etag := md5.Sum(content)
	c.Header("etag", hex.EncodeToString(etag[:]))
	c.Header("vary", "Accept")

	if isImage(extension) {
		// Cache images for a week
		c.Header("cache-control", "public, max-age=604800")
		c.Header("expires", time.Now().In(time.UTC).Add(604800*time.Second).Format(time.RFC1123))
	} else {
		// Cache static files for a day
		c.Header("cache-control", "public, max-age=86400")
		c.Header("expires", time.Now().In(time.UTC).Add(86400*time.Second).Format(time.RFC1123))
	}

	c.Data(http.StatusOK, contentType, content)
}