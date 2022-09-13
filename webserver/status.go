package webserver

import (
	"fmt"
	"net/http"

	"github.com/elmasy-com/columbus/fetcher"
	"github.com/gin-gonic/gin"
)

type status struct {
	Name    string `json:"name"`
	URI     string `json:"uri"`
	Running bool   `json:"running"`          // Indicate that the fetching is running
	Index   int    `json:"index"`            // The currrent index
	Size    int    `json:"size"`             // The maximum index number
	Status  string `json:"status,omitempty"` // Percent complete in format XXX.XXXXXX
	Err     string `json:"error,omitempty"`  // The last error while parsing the log
}

func getStatus(c *gin.Context) {

	v := make([]status, 0, 26)

	for i := range fetcher.Logs {
		s := status{
			Name:    fetcher.Logs[i].GetName(),
			URI:     fetcher.Logs[i].GetURI(),
			Running: fetcher.Logs[i].GetRunning(),
			Index:   fetcher.Logs[i].GetIndex(),
			Size:    fetcher.Logs[i].GetSize(),
		}

		if s.Index != 0 && s.Size != 0 {
			s.Status = fmt.Sprintf("%.6f", float64(s.Index)/float64(s.Size)*100)
		}

		if err := fetcher.Logs[i].GetError(); err != nil {
			s.Err = err.Error()
		}

		v = append(v, s)
	}

	c.JSON(http.StatusOK, v)
}
