package frontend

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type privayPolicyData struct {
	Meta metaData
	Hero heroData
}

// Get400 set context.
// If failed to render the 404.html, returns code 500 with a string: "Internal Server Error".
//
// If d is set to a domain, the 404.html will show "No subdomains found for...".
func GetPrivacyPolicy(c *gin.Context) {

	buf := new(bytes.Buffer)

	dat := privayPolicyData{
		Meta: getMetaData(c.Request, "Columbus Project - Privacy Policy", DefaultDescription),
		Hero: getHeroData("Privacy Policy", "Easy to understand privacy-first policy."),
	}

	if err := templates.ExecuteTemplate(buf, "privacy-policy", dat); err != nil {
		c.Error(fmt.Errorf("failed to render privacy-policy: %w", err))
		Get500(c)
		return
	}

	c.Header("cache-control", "public, max-age=3600, stale-while-revalidate=3600, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(3600*time.Second).Format(time.RFC1123))

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
