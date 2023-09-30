package frontend

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type dnsServerData struct {
	Meta metaData
	Hero heroData
}

// Get400 set context.
// If failed to render the 404.html, returns code 500 with a string: "Internal Server Error".
//
// If d is set to a domain, the 404.html will show "No subdomains found for...".
func GetDNSServer(c *gin.Context) {

	buf := new(bytes.Buffer)

	dat := dnsServerData{
		Meta: getMetaData(c.Request, "Columbus Project - DNS Server", DefaultDescription),
		Hero: getHeroData("DNS Server", "A high performance DNS server to collect and update domains for Columbus Project."),
	}

	if err := templates.ExecuteTemplate(buf, "dns-server", dat); err != nil {
		c.Error(fmt.Errorf("failed to render dns-server.html: %w", err))
		Get500(c)
		return
	}

	c.Header("cache-control", "public, max-age=3600, stale-while-revalidate=3600, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(3600*time.Second).Format(time.RFC1123))

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
