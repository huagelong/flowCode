package bootstrap

import (
	httptransport "anserflow/internal/transport/http"

	"github.com/gin-gonic/gin"
)

func NewHTTPServer(handlers *httptransport.Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	httptransport.RegisterRoutes(r, handlers)
	return r
}
