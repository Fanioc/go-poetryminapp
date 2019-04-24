package main

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/etcdv3"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)

var err error

func main() {
	
	//server instance config
	var (
		instance       = "127.0.0.1:1070"
		serviceAddress = ":1070"
	)
	
	//server etce config
	var (
		prefix   = "/book/"
		key      = prefix + instance
		etcdAddr = "127.0.0.1:2379"
		client   etcdv3.Client
	)
	
	//create logger
	var logger log.Logger
	{
		_, _ = os.OpenFile("/var/log/kit.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "Time", log.DefaultTimestamp)
		logger = log.With(logger, "Caller", log.DefaultCaller)
	}
	
	{
		etcdConfig := etcdv3.ClientOptions{
			DialTimeout:   time.Second * 3,
			DialKeepAlive: time.Second * 3,
		}
		
		client, err = etcdv3.NewClient(context.Background(), []string{etcdAddr}, etcdConfig)
		if err != nil {
			panic(err)
		}
		
		// create regisert
		registrar := etcdv3.NewRegistrar(client, etcdv3.Service{
			Key:   key,
			Value: serviceAddress,
		}, logger)
		
		// register in etcd
		registrar.Register()
	}
	
	//create hystrix
	var commandName = "my-endpoint"
	{
		hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
			Timeout:                1000 * 30,
			ErrorPercentThreshold:  1,
			SleepWindow:            10000,
			MaxConcurrentRequests:  1000,
			RequestVolumeThreshold: 5,
		})
	}
	
	bookServer := new(BookServer)
	bookListHandler := grpc_transport.NewServer(
		makeGetBookListEndpoint(),
		decodeRequest,
		encodeResponse,
	)
	bookServer.bookListHandler = bookListHandler
	
	bookInfoHandler := grpc_transport.NewServer(
		makeGetBookInfoEndpoint(),
		decodeRequest,
		encodeResponse,
	)
	bookServer.bookInfoHandler = bookInfoHandler
	
	ls, _ := net.Listen("tcp", serviceAddress)
	gs := grpc.NewServer(grpc.UnaryInterceptor(grpc_transport.Interceptor))
	services.RegisterBookServiceServer(gs, bookServer)
	gs.Serve(ls)
}

type BookServer struct {
	bookListHandler grpc_transport.Handler
	bookInfoHandler grpc_transport.Handler
}

//通过grpc调用GetBookInfo时,GetBookInfo只做数据透传, 调用BookServer中对应Handler.ServeGRPC转交给go-kit处理
func (s *BookServer) GetBookInfo(ctx context.Context, in *services.BookInfoParams) (*services.BookInfo, error) {
	_, rsp, err := s.bookInfoHandler.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rsp.(*services.BookInfo), err
}

//通过grpc调用GetBookList时,GetBookList只做数据透传, 调用BookServer中对应Handler.ServeGRPC转交给go-kit处理
func (s *BookServer) GetBookList(ctx context.Context, in *services.BookListParams) (*services.BookList, error) {
	_, rsp, err := s.bookListHandler.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rsp.(*services.BookList), err
}

//创建bookList的EndPoint
func makeGetBookListEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		//请求列表时返回 书籍列表
		bl := new(services.BookList)
		bl.BookList = append(bl.BookList, &services.BookInfo{BookId: 1, BookName: "21天精通php"})
		bl.BookList = append(bl.BookList, &services.BookInfo{BookId: 2, BookName: "21天精通java"})
		return bl, nil
	}
}

//创建bookInfo的EndPoint
func makeGetBookInfoEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		//请求详情时返回 书籍信息
		req := request.(*services.BookInfoParams)
		b := new(services.BookInfo)
		b.BookId = req.BookId
		b.BookName = "21天精通php"
		return b, nil
	}
}

func decodeRequest(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func encodeResponse(_ context.Context, rsp interface{}) (interface{}, error) {
	return rsp, nil
}
