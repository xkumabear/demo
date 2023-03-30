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


//白名单由于黑名单，优先校验白名单。
func HttpBlackListMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		//白名单
		whiteIpList :=[]string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			whiteIpList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		blackIpList := []string{}
		if serviceDetail.AccessControl.BlackList != "" {
			blackIpList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}

		if serviceDetail.AccessControl.OpenAuth == 1 && len(whiteIpList) == 0 && len(blackIpList) >0 {
			if public.InStringSlice(blackIpList,c.ClientIP()) {
				middleware.ResponseError(c,3001,errors.New(fmt.Sprintf("%s not in black ip list",c.ClientIP())))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
