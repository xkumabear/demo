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
func HttpWhiteListMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		//白名单
		ipList := []string{}
		if serviceDetail.AccessControl.WhiteList != ""{
			ipList = strings.Split(serviceDetail.AccessControl.WhiteList,",")
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(ipList) > 0 {
			if !public.InStringSlice(ipList,c.ClientIP()) {
				middleware.ResponseError(c,3001,errors.New(fmt.Sprintf("%s not in white ip list",c.ClientIP())))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
