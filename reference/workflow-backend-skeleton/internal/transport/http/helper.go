package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func respondError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}

func mustUint64Param(c *gin.Context, name string) (uint64, bool) {
	raw := c.Param(name)
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return 0, false
	}
	return v, true
}

func errRequired(name string) error {
	return errors.New(name + " is required")
}
