package main

import (
	"context"
	"fmt"
	bookpb "github.com/fanioc/go-poetryminapp/services/book"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/etcdv3"
	"github.com/go-kit/kit/sd/lb"
	"github.com/kataras/muxie"
	"google.golang.org/grpc"
	"io"
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
		httpAddr     = "127.0.0.1:8080"       //网关监听地址
		etcdv3Addr   = "127.0.0.1:2379"       //etcd3 服务发现地址.
		retryMax     = 3                      //负载均衡重试次数
		retryTimeout = 500 * time.Millisecond //负载均衡超时时间
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
	// Now we begin installing the router. Each route corresponds to a single
	// method: sum, concat, uppercase, and count.
	{
		type BookService struct {
			GetBookList endpoint.Endpoint
			GetBookInfo endpoint.Endpoint
		}
		var (
			bookService    = BookService{}
			instancer, err = etcdv3.NewInstancer(client, "/book/", logger)
		)
		if err != nil {
			panic("cannot find discovered server:" + "/book/")
		}
		
		{
			factory := bookFactory(makeBookListEndpoint)
			endpointer := sd.NewEndpointer(instancer, factory, logger)
			balancer := lb.NewRoundRobin(endpointer)
			retry := lb.Retry(retryMax, retryTimeout, balancer)
			bookService.GetBookList = retry
		}
		
		// Here we leverage the fact that addsvc comes with a constructor for an
		// HTTP handler, and just install it under a particular path prefix in
		// our router.
		r.HandleFunc("/book/list/", func(writer http.ResponseWriter, request *http.Request) {
			bookList, err := bookService.GetBookList(request.Context(), bookpb.BookListParams{Page: 1, Limit: 10})
			if err != nil {
				fmt.Println(err)
				writer.Write([]byte("xxx"))
			}
			bl := bookList.(*bookpb.BookList)
			writer.Write([]byte(bl.BookList[0].BookName))
			fmt.Println(*bl)
		})
	}
	
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

func bookFactory(makeBookEndpoint func(bookClient bookpb.BookClient) endpoint.Endpoint) sd.Factory {
	return func(instance string) (i endpoint.Endpoint, closer io.Closer, e error) {
		fmt.Println("instance:" + instance)
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		
		bookClient := bookpb.NewBookClient(conn)
		return makeBookEndpoint(bookClient), conn, nil
	}
}

func makeBookListEndpoint(bookClient bookpb.BookClient) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(bookpb.BookListParams)
		response, err = bookClient.GetBookList(ctx, &req)
		return response, err
	}
}
