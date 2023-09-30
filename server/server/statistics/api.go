package statistics

import (
	"net/http"
	"time"

	"github.com/elmasy-com/columbus/db"
	"github.com/gin-gonic/gin"
)

func GetApiStat(c *gin.Context) {

	s, err := db.StatisticsGetNewest()
	if err != nil {
		c.Error(err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Cache for 10 minutes.
	c.Header("cache-control", "public, max-age=600, stale-while-revalidate=600, stale-if-error=604800")
	c.Header("expires", time.Now().UTC().Add(600*time.Second).Format(time.RFC1123))

	c.JSON(http.StatusOK, s)
}
