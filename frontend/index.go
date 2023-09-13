package frontend

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type indexData struct {
	Meta       metaData
	Statistics statistics
}

func GetIndex(c *gin.Context) {

	statistics, err := parseStatistic()
	if err != nil {
		c.Error(fmt.Errorf("failed to parse statistics: %w", err))
		Get500(c)
		return
	}

	buf := new(bytes.Buffer)
	dat := indexData{Meta: getMetaData(c.Request, DescriptionShort, DescriptionLong), Statistics: statistics}

	err = templates.ExecuteTemplate(buf, "index", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render index.html: %w", err))
		Get500(c)
		return
	}

	// Set ETag
	etag := md5.Sum(buf.Bytes())
	c.Header("etag", hex.EncodeToString(etag[:]))
	c.Header("vary", "Accept")

	// Cache for an hour
	c.Header("cache-control", "public, max-age=3600")
	c.Header("expires", time.Now().In(time.UTC).Add(3600*time.Second).Format(time.RFC1123))

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
