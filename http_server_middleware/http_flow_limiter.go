package http_server_middleware

import (
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)


// 流量统计中间件
func HttpFlowLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//1、
		//1、服务限流
		//3、租户
		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			qps := float64(serviceDetail.AccessControl.ServiceFlowLimit)
			serviceLimiter, err := public.FlowLimitHandler.GetFlowLimiter(public.FlowCountServicePrefix+serviceDetail.Info.ServiceName, qps)
			if err != nil {
				middleware.ResponseError(c, 5001, err)
				c.Abort()
				return
			}
			if !serviceLimiter.Allow() {
				middleware.ResponseError(c, 5002, errors.New("service flow limit err"))
				c.Abort()
				return
			}
		}
		//客户端限流
		if  serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientQps := float64(serviceDetail.AccessControl.ClientIPFlowLimit)
			clientLimiter, err := public.FlowLimitHandler.GetFlowLimiter(public.FlowCountServicePrefix+serviceDetail.Info.ServiceName+"_"+c.ClientIP(), clientQps)
			if err != nil {
				middleware.ResponseError(c, 5003, err)
				c.Abort()
				return
			}
			if !clientLimiter.Allow() {
				middleware.ResponseError(c, 5004, errors.New("clientIP flow limit err"))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
