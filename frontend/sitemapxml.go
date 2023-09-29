package frontend

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type SitemapURL struct {
	Loc        string `xml:"loc"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

type URLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URL     []SitemapURL `xml:"url"`
}

func urlsFromChaosBugbountyList(baseURL string) ([]SitemapURL, error) {

	v := struct {
		Programs []struct {
			Domains []string
		}
	}{}

	resp, err := http.Get("https://raw.githubusercontent.com/projectdiscovery/public-bugbounty-programs/main/chaos-bugbounty-list.json")
	if err != nil {
		return nil, fmt.Errorf("failed to GET chaos-bugbounty-list.json: %w", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	err = json.Unmarshal(out, &v)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}

	r := make([]SitemapURL, 0)

	for i := range v.Programs {

		for j := range v.Programs[i].Domains {

			r = append(r, SitemapURL{Loc: fmt.Sprintf("%s/report/%s", baseURL, v.Programs[i].Domains[j]), ChangeFreq: "hourly", Priority: "1.0"})

		}
	}

	return r, nil
}

func GetSitemapXML(c *gin.Context) {

	m := getMetaData(c.Request, "", "")

	v := URLSet{
		XMLNS: xmlns,
		URL: []SitemapURL{
			{Loc: m.CreateURL("/"), ChangeFreq: "weekly", Priority: "1.0"},
			{Loc: m.CreateURL("/search"), ChangeFreq: "weekly", Priority: "1.0"},
			{Loc: m.CreateURL("/about"), ChangeFreq: "weekly", Priority: "1.0"},
			{Loc: m.CreateURL("/statistics"), ChangeFreq: "hourly", Priority: "1.0"},
			{Loc: m.CreateURL("/api"), ChangeFreq: "daily", Priority: "1.0"},
			{Loc: m.CreateURL("/dns-server"), ChangeFreq: "weekly", Priority: "1.0"},
			{Loc: m.CreateURL("/privacy-policy"), ChangeFreq: "weekly", Priority: "1.0"},
			{Loc: m.CreateURL("/contact"), ChangeFreq: "weekly", Priority: "1.0"},
			{Loc: m.CreateURL("/report/elmasy.com"), ChangeFreq: "hourly", Priority: "1.0"},
		},
	}

	chaosURLs, err := urlsFromChaosBugbountyList(m.BaseURL())
	if err != nil {
		c.Error(fmt.Errorf("failed to parse chaos-bugbounty-list.json for sitemap.xml: %w", err))
		Get500(c)
		return
	}

	v.URL = append(v.URL, chaosURLs...)

	out, err := xml.Marshal(v)
	if err != nil {
		c.Error(fmt.Errorf("failed to marshal sitemap.xml: %w", err))
		Get500(c)
		return
	}

	c.Data(http.StatusOK, "application/xml", out)
}
