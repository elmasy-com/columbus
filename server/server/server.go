package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elmasy-com/columbus/frontend"
	"github.com/elmasy-com/columbus/server/config"
	"github.com/elmasy-com/columbus/server/server/history"
	"github.com/elmasy-com/columbus/server/server/lookup"
	"github.com/elmasy-com/columbus/server/server/report"
	"github.com/elmasy-com/columbus/server/server/statistics"

	"github.com/gin-gonic/gin"
)

func GinLog(param gin.LogFormatterParams) string {

	if param.StatusCode >= 200 && param.StatusCode < 300 && param.Latency < time.Second && config.LogErrorOnly {
		return ""
	}

	return fmt.Sprintf("%s - [%s] \"%s %s\" %d %d \"%s\" %s\n%s",
		param.ClientIP,
		param.TimeStamp.Format(time.RFC1123),
		param.Method,
		param.Path,
		param.StatusCode,
		param.BodySize,
		param.Request.UserAgent(),
		param.Latency,
		param.ErrorMessage,
	)
}

// ServerRun start the http server and block.
// The server can stopped with a SIGINT.
func Run() error {

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	var (
		err    error
		router = gin.New()
		quit   = make(chan os.Signal, 1)
	)

	router.Use(gin.LoggerWithFormatter(GinLog))
	router.Use(gin.Recovery())

	router.NoRoute(frontend.GetStatic)

	router.SetTrustedProxies(config.TrustedProxies)

	router.GET("/", frontend.GetSearch)
	router.GET("/search", frontend.GetSearch)

	router.GET("/api", frontend.GetAPI)

	router.GET("/about", frontend.GetAbout)
	router.GET("/dns-server", frontend.GetDNSServer)
	router.GET("/privacy-policy", frontend.GetPrivacyPolicy)
	router.GET("/contact", frontend.GetContact)

	router.GET("/api/lookup/:domain", lookup.GetApiLookup)
	router.GET("/api/starts/:domain", lookup.GetApiStarts)
	router.GET("/api/tld/:domain", lookup.GetApiTLD)
	router.GET("/api/history/:domain", history.GetApiHistory)

	router.GET("/api/stat", statistics.GetApiStat)
	router.GET("/statistics", frontend.GetStatistics)
	router.GET("/stat", frontend.RedirectStatToStatistics)

	router.GET("/search/:domain", frontend.GetSearchRedirect)
	router.GET("/report/:domain", report.GetReport)
	router.GET("/report", report.RedirectDomainParam)

	router.GET("/api/tools/tld/:fqdn", ToolsTLDGet)
	router.GET("/api/tools/domain/:fqdn", ToolsDomainGet)
	router.GET("/api/tools/subdomain/:fqdn", ToolsSubdomainGet)
	router.GET("/api/tools/isvalid/:fqdn", ToolsIsValidGet)

	// Redirect to /search/:domain
	router.GET("/lookup/:domain", RedirectLookup)

	// Permanent Redirect
	router.GET("/tld/:domain", Redirect)
	router.GET("/tools/tld/:fqdn", Redirect)
	router.GET("/tools/domain/:fqdn", Redirect)
	router.GET("/tools/subdomain/:fqdn", Redirect)
	router.GET("/tools/isvalid/:fqdn", Redirect)

	router.GET("/400", frontend.Get400)
	router.GET("/404", frontend.Get404)
	router.GET("/500", frontend.Get500)
	router.GET("/502", frontend.Get502)
	router.GET("/504", frontend.Get504)

	router.GET("/sitemap.xml", frontend.GetSitemapXML)
	router.GET("/robots.txt", frontend.GetRobotsTxt)

	srv := &http.Server{
		Addr:    config.Address,
		Handler: router,
	}

	go func() {
		if config.SSLCert != "" && config.SSLKey != "" {
			err = srv.ListenAndServeTLS(config.SSLCert, config.SSLKey)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "HTTP Server failed: %s\n", err)
			os.Exit(1)
		}
	}()

	signal.Notify(quit, os.Interrupt, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
