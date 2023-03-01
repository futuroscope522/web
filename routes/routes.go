package routes

import (
	"webFrame/logger"

	"github.com/gin-gonic/gin"
)

func Setup(env string) *gin.Engine {
	if env == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.Default()
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	return r
}
