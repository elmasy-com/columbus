package frontend

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/elmasy-com/elnet/dns"
	"github.com/gin-gonic/gin"
)

type searchData struct {
	Meta metaData
	Hero heroData
}

// Redirect old report path
func GetSearchRedirect(c *gin.Context) {

	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/report/%s", dns.Clean(c.Param("domain"))))
}

func GetSearch(c *gin.Context) {

	buf := new(bytes.Buffer)
	dat := searchData{
		Meta: getMetaData(c.Request, "Columbus Project - "+DefaultTitle, DefaultDescription),
		Hero: getHeroData("Columbus Project", DefaultTitle),
	}

	err := templates.ExecuteTemplate(buf, "search", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render index.html: %w", err))
		Get500(c)
		return
	}

	// Cache for an hour.
	c.Header("cache-control", "public, max-age=3600, stale-while-revalidate=3600, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(3600*time.Second).Format(time.RFC1123))

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
