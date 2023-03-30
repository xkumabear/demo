package http_server_middleware

import (
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/reverse_proxy"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)


// 匹配接入方式  基于请求信息
func HttpReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		//lb管理器
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		lb , err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			middleware.ResponseError(c,1003,err)
			c.Abort()
			return
		}
		transport , err := dao.TransporterHandler.GetTransporter(serviceDetail)
		if err != nil {
			middleware.ResponseError(c,1004,err)
			c.Abort()
			return
		}
		//创建reverse——proxy
		//使用reverserProxy。serverHttp（c。req，c.res）
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(c,lb,transport)
		proxy.ServeHTTP(c.Writer,c.Request)
		c.Abort()
		return
		//c.Next()
	}
}
