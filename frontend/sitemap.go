package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type sitemapURL struct {
	Location string
	Freq     string
	Priority string
}

var predefinedURLs = []sitemapURL{
	{Location: "/", Freq: "daily", Priority: "1.0"},
	{Location: "/search/elmasy.com", Freq: "hourly", Priority: "0.8"},
}

func fetchChaosBugbountyPrograms() ([]sitemapURL, error) {

	resp, err := http.Get("https://raw.githubusercontent.com/projectdiscovery/public-bugbounty-programs/main/chaos-bugbounty-list.json")
	if err != nil {
		return nil, fmt.Errorf("failed to GET: %w", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var v = struct {
		Programs []struct {
			Domains []string `json:"domains"`
		} `json:"programs"`
	}{}

	err = json.Unmarshal(out, &v)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}

	r := make([]sitemapURL, 0, len(v.Programs))

	for i := range v.Programs {

		if len(v.Programs[i].Domains) < 1 {
			continue
		}

		r = append(r, sitemapURL{Location: "/search/" + v.Programs[i].Domains[0], Freq: "hourly", Priority: "0.8"})
	}

	return r, nil
}

func GetSitemapXML(c *gin.Context) {

	sitemapURLs := make([]sitemapURL, 0, 2)
	sitemapURLs = append(sitemapURLs, predefinedURLs...)

	chaosURLs, err := fetchChaosBugbountyPrograms()
	if err != nil {
		c.Error(fmt.Errorf("failed to fetch chaos-bugbounty-list.json: %w", err))
		Get500(c)
		return
	}

	sitemapURLs = append(sitemapURLs, chaosURLs...)

	m := getMetaData(c.Request, "", "")

	for i := range sitemapURLs {
		sitemapURLs[i].Location = m.Host + sitemapURLs[i].Location
	}

	buf := new(bytes.Buffer)

	err = templates.ExecuteTemplate(buf, "sitemap", sitemapURLs)
	if err != nil {
		c.Error(fmt.Errorf("failed to render sitemap.xml: %w", err))
		Get500(c)
		return
	}

	// Cache for an hour
	c.Header("X-Accel-Expire", "86400")

	c.Data(http.StatusOK, "application/xml", []byte(html.UnescapeString(buf.String())))
}
