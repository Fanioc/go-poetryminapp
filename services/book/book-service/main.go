package main

import (
	"flag"
	"fmt"
	"github.com/fanioc/go-poetryminapp/services/book/book-service/svc/server"
	"github.com/go-kit/kit/log"
	"strconv"
	
	"github.com/go-kit/kit/sd/consul"
	
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/hashicorp/consul/api"
	"github.com/openzipkin/zipkin-go"
	reporter "github.com/openzipkin/zipkin-go/reporter/http"
	"os"
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
		svcAddr     = "127.0.0.1"
		instance    = svcAddr + grpcAddress
		port, _     = strconv.Atoi(GRPCAddr[1:])
		consulAddr  = "127.0.0.1:8500"
		
		//etcdAddr    = "127.0.0.1:2379"
		//key = prefix + instance
	)
	
	var logger log.Logger
	{
		_, _ = os.OpenFile("/var/log/bookservice.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestamp)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	
	// Service discovery domain. In this example we use Consul.
	var client consul.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulAddr
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consul.NewClient(consulClient)
		
		serviceConfig := api.AgentServiceRegistration{
			ID:                "book-" + svcAddr + "-" + GRPCAddr[1:],
			Name:              "book",
			Tags:              []string{"book", "book/list", "book/info"},
			Port:              port,
			Address:           svcAddr,
			EnableTagOverride: false,
			Check: &api.AgentServiceCheck{
				CheckID:                        "grpc port",
				Interval:                       "10s", // 健康检查间隔
				DeregisterCriticalServiceAfter: "1m",  // 注销时间，相当于过期时间
				TCP:                            instance,
			},
		}
		
		// 创建注册器
		registrar := consul.NewRegistrar(client, &serviceConfig, logger)
		
		// 注册器启动注册
		registrar.Register()
		
		//redis
		serviceConfig = api.AgentServiceRegistration{
			ID:                "redis-" + "127.0.0.1" + "-6379",
			Name:              "redis",
			Tags:              []string{"redis5.0", "redis"},
			Port:              6379,
			Address:           svcAddr,
			EnableTagOverride: false,
			Check: &api.AgentServiceCheck{
				CheckID:                        "redis",
				Interval:                       "10s", // 健康检查间隔
				DeregisterCriticalServiceAfter: "1h",  // 注销时间，相当于过期时间
				TCP:                            "127.0.0.1:6379",
			},
		}
		registrar = consul.NewRegistrar(client, &serviceConfig, logger)
		registrar.Register()
		
		//mysql
		serviceConfig = api.AgentServiceRegistration{
			ID:                "mysql-" + "127.0.0.1" + "-3306",
			Name:              "mysql",
			Tags:              []string{"mysql8.0", "mysql"},
			Port:              3306,
			Address:           svcAddr,
			EnableTagOverride: false,
			Check: &api.AgentServiceCheck{
				CheckID:                        "mysql",
				Interval:                       "10s", // 健康检查间隔
				DeregisterCriticalServiceAfter: "1h",  // 注销时间，相当于过期时间
				TCP:                            "127.0.0.1:3306",
			},
		}
		registrar = consul.NewRegistrar(client, &serviceConfig, logger)
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
