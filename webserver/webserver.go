package webserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
)

var (
	running bool
	m       sync.Mutex
)

func getRunning() bool {

	m.Lock()
	defer m.Unlock()

	return running
}

func setRunning(v bool) {

	m.Lock()
	defer m.Unlock()

	running = v
}

func Start() {

	if getRunning() {
		return
	}

	setRunning(true)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.GET("/status", getStatus)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Webserver failed: %s\n", err)
		}
	}()

	go func() {
		<-quit
		srv.Shutdown(context.TODO())
		setRunning(false)
		fmt.Printf("webserver closed!\n")
	}()
}
