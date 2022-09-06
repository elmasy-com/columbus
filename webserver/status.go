package webserver

import (
	"fmt"
	"net/http"

	"github.com/elmasy-com/columbus/fetcher"
	"github.com/elmasy-com/columbus/writer"
	"github.com/gin-gonic/gin"
)

type status struct {
	NumFiles int         `json:"numfiles"`
	Logs     []logStatus `json:"logs"`
}

type logStatus struct {
	Name   string `json:"name"`
	URI    string `json:"uri"`
	Index  int    `json:"index"`            // The currrent index
	Size   int    `json:"size"`             // The maximum index number
	Status string `json:"status,omitempty"` // Percent complete in format XXX.XXXXXX
	Err    string `json:"error,omitempty"`  // The last error while parsing the log
}

func getStatus(c *gin.Context) {

	v := make([]logStatus, 0, 26)

	for i := range fetcher.URIS {
		s := logStatus{
			Name:  fetcher.URIS[i].GetName(),
			URI:   fetcher.URIS[i].GetURI(),
			Index: fetcher.URIS[i].GetIndex(),
			Size:  fetcher.URIS[i].GetSize(),
		}

		if s.Index != 0 && s.Size != 0 {
			s.Status = fmt.Sprintf("%.6f", float64(s.Index)/float64(s.Size)*100)
		}

		if err := fetcher.URIS[i].GetError(); err != nil {
			s.Err = err.Error()
		}

		v = append(v, s)
	}

	c.JSON(http.StatusOK, status{NumFiles: writer.NumFiles, Logs: v})
}
