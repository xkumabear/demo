package http_server_middleware

import (
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)


// 匹配接入方式  基于请求信息
func HttpStripUriMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		//去除前缀，拿path后缀
		//前缀匹配
		if serviceDetail.HTTP.RuleType == public.HTTPRuleTypePrefixURL && serviceDetail.HTTP.NeedStripUri == 1{
			fmt.Println(c.Request.URL.Path)
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path,serviceDetail.HTTP.Rule,"",1)
			fmt.Println(c.Request.URL.Path)
		}
		c.Next()
	}
}
