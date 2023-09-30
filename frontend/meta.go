package frontend

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/elmasy-com/slices"
)

const (
	DefaultTitle       = "Advanced subdomain discovery service."
	DefaultDescription = "Columbus Project is an API-first subdomain discovery service. A blazingly fast subdomain enumeration service with advanced queries."
)

type metaData struct {
	Proto       string
	Host        string
	Slug        string
	SlugParts   []string
	Title       string
	Description string
}

// BaseURL return the base URL (eg.: "https://example.com")
func (m *metaData) BaseURL() string {
	return fmt.Sprintf("%s://%s", m.Proto, m.Host)
}

// CreateURL create an URL from the base url and slug s.
// (eg.: "https://example.com/slug")
func (m *metaData) CreateURL(s string) string {
	return fmt.Sprintf("%s://%s%s", m.Proto, m.Host, s)
}

func getMetaData(r *http.Request, title string, description string) metaData {

	dat := metaData{}

	dat.Proto = r.Header.Get("X-Forwarded-Proto")
	if dat.Proto == "" {

		dat.Proto = r.URL.Scheme
		if dat.Proto == "" {
			dat.Proto = "http"
		}
	}

	dat.Host = r.Host
	if dat.Host == "" {
		dat.Host = "unknown"
	}

	dat.Slug = r.URL.Path

	dat.SlugParts = make([]string, 0)

	dat.SlugParts = slices.AppendUniques(dat.SlugParts, strings.Split(dat.Slug, "/")...)

	dat.Title = title
	dat.Description = description

	return dat
}
