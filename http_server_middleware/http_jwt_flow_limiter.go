package http_server_middleware

import (
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)


// 流量统计中间件
func HttpJwtFlowLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appInterface,ok := c.Get("app")
		if !ok {
			c.Next()
			return
		}
		appInfo := appInterface.(*dao.App)

		if appInfo.Qps > 0 {
			clientLimiter, err := public.FlowLimitHandler.GetFlowLimiter(public.FlowCountAppPrefix+appInfo.AppID+c.ClientIP(), float64(appInfo.Qps))
			if err != nil {
				middleware.ResponseError(c, 5001, err)
				c.Abort()
				return
			}
			//基于并发数拦截
			if !clientLimiter.Allow() {
				middleware.ResponseError(c, 5002, errors.New("QPS flow limit err"))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
