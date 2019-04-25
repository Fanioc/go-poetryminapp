package main

import (
	"context"
	"fmt"
	"github.com/fanioc/go-poetryminapp/apigateway/router"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/etcdv3"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/kataras/muxie"
	"github.com/openzipkin/zipkin-go"
	reporter "github.com/openzipkin/zipkin-go/reporter/http"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var err error

func main() {
	//apigateway configure.
	var (
		httpAddr   = "127.0.0.1:8080" //网关监听地址
		etcdv3Addr = "127.0.0.1:2379" //etcd3 服务发现地址.
	)
	
	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr) // try to openfile
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	
	// Service discovery domain. In this example we use Consul.
	var client etcdv3.Client
	{
		etcdConfig := etcdv3.ClientOptions{
			DialTimeout:   time.Second * 3,
			DialKeepAlive: time.Second * 3,
		}
		
		client, err = etcdv3.NewClient(context.Background(), []string{etcdv3Addr}, etcdConfig)
		if err != nil {
			panic(err)
		}
	}
	
	// Transport
	var zkClientTrace kitgrpc.ClientOption
	{
		//创建zipkin上报管理器
		reporte := reporter.NewReporter("http://localhost:9411/api/v2/spans")
		
		//运行结束，关闭上报管理器的for-select协程
		defer reporte.Close()
		
		//创建trace跟踪器
		zkTracer, err := zipkin.NewTracer(reporte)
		
		if err != nil {
			fmt.Println("err Tracer :" + err.Error())
		}
		//添加grpc请求的before after finalizer 事件对应要处理的trace操作方法
		zkClientTrace = kitzipkin.GRPCClientTrace(zkTracer)
	}
	
	//resiger routers
	r := muxie.NewMux()
	router.RegisterRouter(r, &client, &logger, &zkClientTrace)
	
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
