package frontend

import (
	"fmt"
	"net/http"
)

const (
	DescriptionShort = "Columbus Project - A fast, API-first subdomain discovery service with advanced queries."
	DescriptionLong  = "Columbus Project is an API-first subdomain discovery service. A blazingly fast subdomain enumeration service with advanced queries."
)

type metaData struct {
	Host             string
	Slug             string
	DescriptionShort string
	DescriptionLong  string
}

func getMetaData(r *http.Request, ds string, dl string) metaData {

	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {

		proto = r.URL.Scheme
		if proto == "" {
			proto = "http"
		}
	}

	host := r.Host
	if host == "" {
		host = "unknown"
	}

	return metaData{
		Host:             fmt.Sprintf("%s://%s", proto, host),
		Slug:             r.URL.RequestURI(),
		DescriptionShort: ds,
		DescriptionLong:  dl,
	}
}
