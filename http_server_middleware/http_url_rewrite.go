package http_server_middleware

import (
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)


// 匹配接入方式  基于请求信息
func HttpUrlWriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		//重写uri
		for _ ,item := range strings.Split(serviceDetail.HTTP.UrlRewrite,","){
			items := strings.Split(item," ")
			if len(items) != 2 {
				continue
			}
			regexp,err := regexp.Compile(items[0])
			if err != nil {
				fmt.Println("url_rewrite_err:",err)
				continue
			}
			fmt.Println("bef_uri_rewrite",c.Request.URL.Path)
			replacePath := regexp.ReplaceAll([]byte(c.Request.URL.Path),[]byte(items[1]))
			c.Request.URL.Path = string(replacePath)
			fmt.Println("aft_uri_rewrite",c.Request.URL.Path)
		}

		c.Next()
	}
}
