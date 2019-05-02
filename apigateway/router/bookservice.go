package router

import (
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	bookpb "github.com/fanioc/go-poetryminapp/services/book"
	"github.com/fanioc/go-poetryminapp/services/book/book-service/svc"
	grpcclient "github.com/fanioc/go-poetryminapp/services/book/book-service/svc/client/grpc"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/kataras/muxie"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"time"
)

var bookService = BookService{
	Name: "book",
	RouteConfig: map[string]*RouteConfig{
		"list": {
			retryMax:       3,
			retryTimeout:   1 * time.Second,
			circuitbreaker: hystrix.CommandConfig{},
			funcHttpHandl:  nil,
		},
		"info": {
			retryMax:       3,
			retryTimeout:   1 * time.Second,
			circuitbreaker: hystrix.CommandConfig{},
			funcHttpHandl:  nil,
		},
	},
}

type BookService struct {
	Name        string //etcd prefix
	RouteConfig map[string]*RouteConfig
	svc.Endpoints
}

func BookRegisterRouter(r *muxie.Mux, consulclient *consul.Client, logger *log.Logger, zkClientTrace *kitgrpc.ClientOption) {
	
	instancer := consul.NewInstancer(*consulclient, *logger, bookService.Name, []string{"book"}, false)
	
	commandName := bookService.Name
	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		Timeout:                1000 * 30,
		ErrorPercentThreshold:  1,
		SleepWindow:            10000,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 5,
	})
	
	// Add prefix in router
	sub := r.Of(bookService.Name)
	// book/list/
	{
		factory := bookEndpointFactory(makeBookList, *zkClientTrace)
		endpointer := sd.NewEndpointer(instancer, factory, *logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(bookService.RouteConfig["list"].retryMax, bookService.RouteConfig["list"].retryTimeout, balancer)
		
		//add ciruitbreaker
		breakerMw := circuitbreaker.Hystrix(commandName)
		retry = breakerMw(retry)
		
		bookService.GetBookListEndpoint = retry
		sub.HandleFunc("/list/", getBookList)
	}
	
	// /book/info/
	{
		factory := bookEndpointFactory(makeBookInfo, *zkClientTrace)
		endpointer := sd.NewEndpointer(instancer, factory, *logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(bookService.RouteConfig["info"].retryMax, bookService.RouteConfig["info"].retryTimeout, balancer)
		
		// hystrix config
		breakerMw := circuitbreaker.Hystrix(commandName)
		retry = breakerMw(retry)
		
		bookService.GetBookInfoEndpoint = retry
		sub.HandleFunc("/info/", getBookInfo)
	}
}

func bookEndpointFactory(makeEndpoint func(svc.Endpoints) endpoint.Endpoint, zkClientTrace kitgrpc.ClientOption) sd.Factory {
	return func(instance string) (i endpoint.Endpoint, closer io.Closer, e error) {
		fmt.Println("## instance:" + instance)
		conn, err := grpc.Dial(instance, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			return nil, nil, err
		}
		pbservice, err := grpcclient.New(conn, zkClientTrace)
		if err != nil {
			_ = conn.Close()
			return nil, nil, err
		}
		return makeEndpoint(pbservice), conn, nil
	}
}

func makeBookList(svc svc.Endpoints) endpoint.Endpoint {
	return svc.GetBookListEndpoint
}

func makeBookInfo(svc svc.Endpoints) endpoint.Endpoint {
	return svc.GetBookInfoEndpoint
}

func getBookList(writer http.ResponseWriter, request *http.Request) {
	bookList, err := bookService.GetBookList(request.Context(), &bookpb.BookListParams{Page: 1, Limit: 10})
	if err != nil {
		writer.WriteHeader(404)
		fmt.Println(err)
		writer.Write([]byte("请稍后再试."))
		return
	}
	
	writer.Write([]byte(bookList.BookList[0].BookName + "</br>" + bookList.BookList[1].BookName))
	fmt.Println(*bookList)
}

func getBookInfo(writer http.ResponseWriter, request *http.Request) {
	
	bookInfo, err := bookService.GetBookInfo(request.Context(), &bookpb.BookInfoParams{BookId: 1})
	
	if err != nil {
		writer.WriteHeader(404)
		fmt.Println(err)
		writer.Write([]byte("请稍后再试."))
		return
	}
	
	writer.Write([]byte(bookInfo.BookName))
	fmt.Println(*bookInfo)
}
