package main

import (
	"fmt"
	"github.com/fanioc/go-poetryminapp/apigateway/discover"
	logger2 "github.com/fanioc/go-poetryminapp/apigateway/logger"
	"github.com/fanioc/go-poetryminapp/apigateway/router"
	"github.com/fanioc/go-poetryminapp/apigateway/tracer"
	"github.com/hashicorp/consul/api"
	"github.com/kataras/muxie"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//apigateway configure.
	var (
		httpAddr = "127.0.0.1:8080" //网关监听地址
		
		zipkinReportAddr = "http://localhost:9411/api/v2/spans"
		consulAddr       = "127.0.0.1:8500"
		//etcdv3Addr       = "127.0.0.1:2379" //etcd3 服务发现地址.
	)
	
	// Logging domain.
	var logger = logger2.CreateKitLog()
	
	// Service discovery domain. In this example we use consul.
	var consulClient, err = discover.ConnectConsul(consulAddr)
	if err != nil {
		panic(err)
	}
	
	err = consulClient.Deregister(&api.AgentServiceRegistration{
		ID:   "/book/:5070",
		Name: "",
	})
	logger.Log("deregister:", err)
	
	// tracer
	var zkClientTrace, report = tracer.RegisterZipkinTracer(zipkinReportAddr)
	defer report.Close()
	
	//resiger routers
	r := muxie.NewMux()
	router.RegisterRouter(r, &consulClient, &logger, &zkClientTrace)
	
	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()
	
	// HTTP transport.
	go func() {
		_ = logger.Log("transport", "HTTP", "addr", httpAddr)
		errc <- http.ListenAndServe(httpAddr, r)
	}()
	
	// Run!
	_ = logger.Log("exit", <-errc)
}
