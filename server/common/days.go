package common

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseQueryDays return the "days" query parameter (eg.: "/api/lookup/example.com?days=1").
// If not set, returns -1.
func ParseQueryDays(c *gin.Context) (int, error) {

	// Parse days query param
	daysStr, daysSet := c.GetQuery("days")
	if !daysSet {
		return -1, nil
	}

	if daysStr == "" {
		return -2, fmt.Errorf("empty")
	}

	return strconv.Atoi(daysStr)
}
