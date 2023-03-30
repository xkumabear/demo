package main

import (
	"flag"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/grpc_server_router"
	"github.com/e421083458/demo/http_server_router"
	"github.com/e421083458/demo/router"
	"github.com/e421083458/demo/tcp_server_router"
	"github.com/e421083458/golang_common/lib"
	"os"
	"os/signal"
	"syscall"
)


// endpoint { dashboard<admin>  proxy_server<agent>}
// config { 。/conf/prod/对应配置文件夹 }

var(
	endpoint = flag.String("endpoint","","input endpoint dashboard or service")
	config = flag.String("config","","input config file like  ./conf/dev/")
)

func main()  {

	flag.Parse()
	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *endpoint == "dashboard" {
		lib.InitModule(*config, []string{"base", "mysql", "redis",})
		//如果1为空，默认从命令行中读取
		defer lib.Destroy() //退出
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()
	} else {
		lib.InitModule(*config, []string{"base", "mysql", "redis",})
		//如果1为空，默认从命令行中读取
		defer lib.Destroy() //退出
		dao.ServiceManagerHandler.LoadOnce()//加载服务列表
		dao.AppManagerHandler.LoadOnce()//加载租户列表

		go func() {
			http_server_router.HttpServerRun()
		}()
		go func() {
			http_server_router.HttpsServerRun()
		}()
		go func() {
			tcp_server_router.TcpServerRun()
		}()
		go func() {
			grpc_server_router.GrpcServerRun()
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit


		grpc_server_router.GrpcServerStop()
		tcp_server_router.TcpServerStop()
		http_server_router.HttpServerStop()
		http_server_router.HttpsServerStop()
	}
}