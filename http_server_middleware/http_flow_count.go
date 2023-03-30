package http_server_middleware

import (
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)


// 流量统计中间件
func HttpFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface,ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c,1002,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//1、全站流量统计
		//2、服务
		//3、租户
		totalCount ,err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			middleware.ResponseError(c,4004,err)
			c.Abort()
			return
		}
		totalCount.Increase()
		totalDayCount,_:=totalCount.GetDayData(time.Now())
		fmt.Println("totalCount:",totalDayCount,":qps:",totalCount.QPS)

		serviceCount ,err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix+serviceDetail.Info.ServiceName)
		if err != nil {
			middleware.ResponseError(c,4004,err)
			c.Abort()
			return
		}
		serviceCount.Increase()
		serviceDayCount,_:=serviceCount.GetDayData(time.Now())
		fmt.Println("serviceCount:",serviceDayCount,":qps:",serviceCount.QPS)

		//appCount ,err := public.FlowCounterHandler.GetCounter(public.FlowCountAppPrefix)
		//if err != nil {
		//	middleware.ResponseError(c,4004,err)
		//	c.Abort()
		//	return
		//}
		//appCount.Increase()
		//fmt.Println("appCount:",appCount)

		c.Next()
	}
}
