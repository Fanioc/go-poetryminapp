package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fanioc/go-poetryminapp/services/book/book-service/svc/server"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/etcdv3"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/openzipkin/zipkin-go"
	reporter "github.com/openzipkin/zipkin-go/reporter/http"
	"os"
	"time"
)

func main() {
	
	DebugAddr := ""
	GRPCAddr := ""
	HTTPAddr := ""
	
	flag.StringVar(&DebugAddr, "debug.addr", ":5060", "Debug and metrics listen address")
	flag.StringVar(&HTTPAddr, "http.addr", ":5050", "HTTP listen address")
	flag.StringVar(&GRPCAddr, "grpc.addr", ":5040", "gRPC (HTTP) listen address")
	
	flag.Parse()
	
	var (
		grpcAddress = GRPCAddr
		instance    = "127.0.0.1" + grpcAddress
		prefix      = "/book/"
		etcdAddr    = "127.0.0.1:2379"
		key         = prefix + instance
	)
	var err error
	
	var logger log.Logger
	{
		_, _ = os.OpenFile("/var/log/bookservice.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestamp)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	
	var client etcdv3.Client
	{
		etcdConfig := etcdv3.ClientOptions{
			DialTimeout:   time.Second * 3,
			DialKeepAlive: time.Second * 3,
		}
		
		client, err = etcdv3.NewClient(context.Background(), []string{etcdAddr}, etcdConfig)
		if err != nil {
			panic(err)
		}
		
		// 创建注册器
		registrar := etcdv3.NewRegistrar(client, etcdv3.Service{
			Key:   key,
			Value: instance,
		}, logger)
		
		// 注册器启动注册
		registrar.Register()
	}
	
	var zkServerTrace kitgrpc.ServerOption
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
		zkServerTrace = kitzipkin.GRPCServerTrace(zkTracer)
	}
	
	server.Run(server.Config{
		HTTPAddr:  HTTPAddr,
		DebugAddr: DebugAddr,
		GRPCAddr:  grpcAddress,
	}, zkServerTrace)
}
