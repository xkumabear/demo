package http_server_router

import (
	"github.com/e421083458/demo/controller"
	"github.com/e421083458/demo/http_server_middleware"
	"github.com/e421083458/demo/middleware"
	"github.com/gin-gonic/gin"
)


func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	oauth := router.Group("/oauth")
	oauth.Use(middleware.TranslationMiddleware())
	{
		controller.RegisterOauth(oauth)
	}

	router.Use(http_server_middleware.HttpAccessModeMiddleware(),

		http_server_middleware.HttpFlowCountMiddleware(),
		http_server_middleware.HttpFlowLimiterMiddleware(),

		http_server_middleware.HttpJwtAuthTokenMiddleware(),
		http_server_middleware.HttpJwtFlowCountMiddleware(),
		http_server_middleware.HttpWhiteListMiddleware(),
		http_server_middleware.HttpBlackListMiddleware(),

		http_server_middleware.HttpHeaderTransferMiddleware(),
		http_server_middleware.HttpStripUriMiddleware(),
		http_server_middleware.HttpUrlWriteMiddleware(),

		http_server_middleware.HttpReverseProxyMiddleware())


	return router
}
