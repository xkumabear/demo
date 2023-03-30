package http_server_middleware

import (
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)


// 流量统计中间件
func HttpJwtFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appInterface,ok := c.Get("app")
		if !ok {
			c.Next()
			return
		}

		appInfo := appInterface.(*dao.App)
		appCount ,err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix+appInfo.AppID)
		if err != nil {
			middleware.ResponseError(c,4004,err)
			c.Abort()
			return
		}
		appCount.Increase()
		if appInfo.Qpd > 0 && appCount.TotalCount > appInfo.Qpd{
			middleware.ResponseError(c, 2003, errors.New(fmt.Sprintf("租户日请求量限流 limit:%v current:%v",appInfo.Qpd,appCount.TotalCount)))
			c.Abort()
			return
		}
		fmt.Println("appCount:",appCount)

		c.Next()
	}
}
