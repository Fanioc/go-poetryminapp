package main

import (
	"context"
	"flag"
	"github.com/fanioc/go-poetryminapp/services/book/book-service/svc/server"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/etcdv3"
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
	}
	
	// 创建注册器
	registrar := etcdv3.NewRegistrar(client, etcdv3.Service{
		Key:   key,
		Value: instance,
	}, logger)
	
	// 注册器启动注册
	registrar.Register()
	
	server.Run(server.Config{
		HTTPAddr:  HTTPAddr,
		DebugAddr: DebugAddr,
		GRPCAddr:  grpcAddress,
	})
}
