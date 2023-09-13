package stat

import (
	"net/http"

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

	c.Header("X-Accel-Expires", "600")

	c.JSON(http.StatusOK, s)
}

func RedirectStat(c *gin.Context) {

	c.Header("location", "https://columbus.elmasy.com/#statistics")
	c.Status(http.StatusTemporaryRedirect)
}
