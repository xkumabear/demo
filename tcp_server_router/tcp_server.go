package tcp_server_router

import (
	"context"
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/reverse_proxy"
	"github.com/e421083458/demo/tcp_server"
	"github.com/e421083458/demo/tcp_server_middleware"
	"log"
	"net"
)

var tcpServerList = []*tcp_server.TcpServer{}

type tcpHandler struct {
}

func (t *tcpHandler) ServeTcp(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

func TcpServerRun() {
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, Item := range serviceList {
		tempItem := Item
		log.Printf(" [INFO] TcpProxyRun:%s\n", tempItem.TCPRule.Port)
		go func(serviceInfo *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", tempItem.TCPRule.Port)

			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceInfo)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}

			//构建路由及设置中间件
			router := tcp_server_middleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcp_server_middleware.TCPFlowCountMiddleware(),
				tcp_server_middleware.TCPFlowLimitMiddleware(),
				tcp_server_middleware.TCPWhiteListMiddleware(),
				tcp_server_middleware.TCPBlackListMiddleware(),
			)

			//构建回调Handler
			routerHandler := tcp_server_middleware.NewTcpSliceRouterHandler(
				func(c *tcp_server_middleware.TcpSliceRouterContext) tcp_server.TCPHandler {
					return reverse_proxy.NewTcpLoadBalanceReverseProxy(c, rb)
				}, router)
			baseCtx := context.WithValue(context.Background(), "service", serviceInfo)
			tcpserver := &tcp_server.TcpServer{
				Addr:    addr,
				Handler: routerHandler,
				BaseCtx: baseCtx,
			}
			tcpServerList = append(tcpServerList, tcpserver)
			log.Printf(" [INFO] tcp_proxy_run %v\n", addr)
			err = tcpserver.ListenAndServe()
			if err != nil && err != tcp_server.ErrServerClosed {
				log.Fatalf(" [INFO] TcpProxyRun %v err:%v\n", tempItem.TCPRule.Port, err)
			}
		}(tempItem)
	}
}

func TcpServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf(" [INFO] TcpProxyStop stopped %v\n", tcpServer.Addr)
	}
}
