package http

import (
	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/config"
	v1 "github.com/maypok86/payment-api/internal/handler/http/v1"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger(), gin.Recovery())

	if config.Get().IsProd() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	api := router.Group("/api")
	{
		v1Handler := v1.NewHandler(logger)
		v1Handler.InitAPI(api)
	}

	return router
}
