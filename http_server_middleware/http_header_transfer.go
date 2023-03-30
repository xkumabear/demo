package http_server_middleware

import (
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)


// 匹配接入方式  基于请求信息
func HttpHeaderTransferMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		for _, item := range strings.Split(serviceDetail.HTTP.HeaderTransfor,","){
			items := strings.Split(item," ")
			if len(items) != 3{
				fmt.Println("no no no ")
				continue
			}
			if items[0] == "add" || items[0]=="edit" {
				c.Request.Header.Set(items[1],items[2])
				fmt.Println("items",items[0],items[1],items[2])
			}
			if items[0] == "del" {
				c.Request.Header.Del(items[1])
			}
		}
		c.Next()
	}
}
