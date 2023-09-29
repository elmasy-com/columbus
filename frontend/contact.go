package frontend

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type contactData struct {
	Meta metaData
	Hero heroData
}

// Get400 set context.
// If failed to render the 404.html, returns code 500 with a string: "Internal Server Error".
//
// If d is set to a domain, the 404.html will show "No subdomains found for...".
func GetContact(c *gin.Context) {

	buf := new(bytes.Buffer)

	dat := contactData{
		Meta: getMetaData(c.Request, "Columbus Project - 400 Bad Request", DefaultDescription),
		Hero: getHeroData("Contact", "Follow us or send a feedback!"),
	}

	if err := templates.ExecuteTemplate(buf, "contact", dat); err != nil {
		c.Error(fmt.Errorf("failed to render contact: %w", err))
		Get500(c)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
