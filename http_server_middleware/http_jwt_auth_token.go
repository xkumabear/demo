package http_server_middleware

import (
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)


// 匹配接入方式  基于请求信息
func HttpJwtAuthTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//decode jwt token
		//app_id  app_list 取appInfo，并存入gin.c
		//
		matched := false
		author := c.GetHeader("Authorization")
		token := strings.ReplaceAll(author,"Bearer ","")
		if token != ""{
			claims , err := public.JwtDecode(token)
			if err != nil {
				middleware.ResponseError(c,2003,err)
				c.Abort()
				return
			}
			appList := dao.AppManagerHandler.GetAppList()
			for _ , Info := range appList {
				if Info.AppID == claims.Issuer {
					c.Set("appDetail",Info)
					matched = true
					break
				}
			}
		}
		//权限校验
		if serviceDetail.AccessControl.OpenAuth == 1 && !matched {
			middleware.ResponseError(c,2003,errors.New("not matched appInfo"))
			c.Abort()
			return
		}
		c.Next()
	}
}
