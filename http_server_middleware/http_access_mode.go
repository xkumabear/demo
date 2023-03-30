package http_server_middleware

import (
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
)


// 匹配接入方式  基于请求信息
func HttpAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service , err := dao.ServiceManagerHandler.HttpAccessMode(c)
		if err != nil {
			middleware.ResponseError(c,1001 , err)
			c.Abort()
			return
		}
		fmt.Println("matched service" , public.ObjToJson(service))
		c.Set("service",service)//设置上下文
		c.Next()
	}
}
