package main

import (
	"context"
	"github.com/fanioc/go-poetryminapp/services/book/book-service/svc/server"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/etcdv3"
	"os"
	"time"
)

func main() {
	var (
		grpcAddress = ":5050"
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
		HTTPAddr:  ":5051",
		DebugAddr: ":5052",
		GRPCAddr:  grpcAddress,
	})
}
