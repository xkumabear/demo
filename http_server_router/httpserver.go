package http_server_router

import (
	"context"
	"github.com/e421083458/demo/cert_file"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

var (
	HttpSrvHandler *http.Server
	HttpsSrvHandler *http.Server
)

func HttpServerRun() {
	//gin.SetMode(lib.ConfBase.DebugMode)
	gin.SetMode(lib.GetStringConf("proxy.base.debug_mode"))
	r := InitRouter(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog())

	HttpSrvHandler = &http.Server{
		Addr:           lib.GetStringConf("proxy.http.addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.http.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.http.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.http.max_header_bytes")),
	}

	log.Printf(" [INFO] HttpProxyRun:%s\n",lib.GetStringConf("proxy.http.addr"))
	if err := HttpSrvHandler.ListenAndServe(); err != nil && err != http.ErrServerClosed{
		log.Fatalf(" [ERROR] HttpProxyRun:%s err:%v\n", lib.GetStringConf("proxy.http.addr"), err)
	}
}

func HttpServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpSrvHandler.Shutdown(ctx); err != nil {
		log.Printf(" [ERROR] http_proxy_stop err:%v\n", err)
	}
	log.Printf(" [INFO] HttpProxyStop stopped\n")
}

func HttpsServerRun() {
	//gin.SetMode(lib.ConfBase.DebugMode)
	gin.SetMode(lib.GetStringConf("proxy.base.debug_mode"))
	r := InitRouter(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog())

	HttpsSrvHandler = &http.Server{
		Addr:           lib.GetStringConf("proxy.https.addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.https.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.https.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.https.max_header_bytes")),
	}

	log.Printf(" [INFO] HttpsProxyRun:%s\n",lib.GetStringConf("proxy.https.addr"))
	if err := HttpsSrvHandler.ListenAndServeTLS( cert_file.Path("server.crt"), cert_file.Path("server.key")); err != nil && err != http.ErrServerClosed{
		log.Fatalf(" [ERROR] HttpsProxyRun:%s err:%v\n", lib.GetStringConf("proxy.https.addr"), err)
	}
}

func HttpsServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpsSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] HttpsProxyStop err:%v\n", err)
	}
	log.Printf(" [INFO] HttpsProxyStop stopped\n")
}