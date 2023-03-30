package grpc_server_router

import (
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/grpc_server_middleware"
	"github.com/e421083458/demo/reverse_proxy"
	"github.com/e421083458/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"log"
	"net"
)

var grpcServerList = []*warpGrpcServer{}

type warpGrpcServer struct {
	Addr string
	*grpc.Server
}


func GrpcServerRun() {
	serviceList := dao.ServiceManagerHandler.GetGrpcServiceList()
	for  _ , Item := range serviceList {
		tempItem := Item
		go func(serviceInfo *dao.ServiceDetail) {
			//
			addr := fmt.Sprintf(":%d", serviceInfo.GRPCRule.Port)
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceInfo)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatalf(" [INFO] GrpcListen %v err:%v\n", addr, err)
			}
			grpcHandler := reverse_proxy.NewGrpcLoadBalanceHandler(rb)

			s := grpc.NewServer(
				grpc.ChainStreamInterceptor(

					grpc_server_middleware.GrpcFlowCountMiddleware(serviceInfo),
					grpc_server_middleware.GrpcFlowLimitMiddleware(serviceInfo),

					grpc_server_middleware.GrpcJwtAuthTokenMiddleware(serviceInfo),
					grpc_server_middleware.GrpcJwtFlowCountMiddleware(serviceInfo),
					grpc_server_middleware.GrpcJwtFlowLimitMiddleware(serviceInfo),

					grpc_server_middleware.GrpcWhiteListMiddleware(serviceInfo),
					grpc_server_middleware.GrpcBlackListMiddleware(serviceInfo),

					grpc_server_middleware.GrpcHeaderTransferMiddleware(serviceInfo),
				),
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler))

			grpcServerList = append(grpcServerList, &warpGrpcServer{
				Addr:   addr,
				Server: s,
			})
			log.Printf(" [INFO] grpc_proxy_run %v\n", addr)
			if err := s.Serve(lis); err != nil {
				log.Fatalf(" [INFO] grpc_proxy_run %v err:%v\n", addr, err)
			}
		}(tempItem)
	}
}

func GrpcServerStop() {
	for _, grpcServer := range grpcServerList {
		grpcServer.GracefulStop()
		log.Printf(" [INFO] grpc_proxy_stop %v stopped\n", grpcServer.Addr)
	}
}
