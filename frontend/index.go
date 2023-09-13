package frontend

import (
	"bytes"
	"fmt"
	"net/http"

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

	// Cache for an hour
	c.Header("X-Accel-Expires", "3600")

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
