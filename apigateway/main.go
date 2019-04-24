package main

import (
	"context"
	"fmt"
	"github.com/fanioc/go-poetryminapp/apigateway/router"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/etcdv3"
	"github.com/kataras/muxie"
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
	
	// Transport domain.
	//tracer := stdopentracing.GlobalTracer() // no-op
	//zipkinTracer, _ := stdzipkin.NewTracer(nil, stdzipkin.WithNoopTracer(true))
	
	r := muxie.NewMux()
	
	router.RegisterRouter(r, &client, &logger)
	// Now we begin installing the router. Each route corresponds to a single
	// method: sum, concat, uppercase, and count.
	//{
	//	type BookService struct {
	//		GetBookList endpoint.Endpoint
	//		GetBookInfo endpoint.Endpoint
	//	}
	//	var (
	//		bookService    = BookService{}
	//		instancer, err = etcdv3.NewInstancer(client, "/book/", logger)
	//	)
	//	if err != nil {
	//		panic("cannot find discovered server:" + "/book/")
	//	}
	//
	//	{
	//		factory := bookFactory(makeBookListEndpoint)
	//		endpointer := sd.NewEndpointer(instancer, factory, logger)
	//		balancer := lb.NewRoundRobin(endpointer)
	//		retry := lb.Retry(retryMax, retryTimeout, balancer)
	//		bookService.GetBookList = retry
	//	}
	//
	//	// Here we leverage the fact that addsvc comes with a constructor for an
	//	// HTTP handler, and just install it under a particular path prefix in
	//	// our router.
	//	r.HandleFunc("/book/list/", func(writer http.ResponseWriter, request *http.Request) {
	//
	//		bookList, err := bookService.GetBookList(request.Context(), bookpb.BookListParams{Page: 1, Limit: 10})
	//		if err != nil {
	//			fmt.Println(err)
	//			writer.Write([]byte("xxx"))
	//		}
	//		bl := bookList.(*bookpb.BookList)
	//		writer.Write([]byte(bl.BookList[0].BookName))
	//		fmt.Println(*bl)
	//	})
	//}
	
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
