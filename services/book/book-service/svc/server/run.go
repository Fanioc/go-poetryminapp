// Version: 574fb16d86
// Version Date: 2019年 04月 12日 星期五 00:42:59 UTC

package server

import (
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	
	// 3d Party
	"google.golang.org/grpc"
	
	// This Service
	pb "github.com/fanioc/go-poetryminapp/services/book"
	"github.com/fanioc/go-poetryminapp/services/book/book-service/handlers"
	"github.com/fanioc/go-poetryminapp/services/book/book-service/svc"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// Config contains the required fields for running a server
type Config struct {
	HTTPAddr  string
	DebugAddr string
	GRPCAddr  string
}

func NewEndpoints() svc.Endpoints {
	// Business domain.
	var service pb.BookServer
	{
		service = handlers.NewService()
		// Wrap Service with middlewares. See handlers/middlewares.go
		service = handlers.WrapService(service)
	}
	
	// Endpoint domain.
	var (
		getbookinfoEndpoint = svc.MakeGetBookInfoEndpoint(service)
		getbooklistEndpoint = svc.MakeGetBookListEndpoint(service)
	)
	
	endpoints := svc.Endpoints{
		GetBookInfoEndpoint: getbookinfoEndpoint,
		GetBookListEndpoint: getbooklistEndpoint,
	}
	
	// Wrap selected Endpoints with middlewares. See handlers/middlewares.go
	endpoints = handlers.WrapEndpoints(endpoints)
	
	return endpoints
}

// Run starts a new http server, gRPC server, and a debug server with the
// passed config and logger
func Run(cfg Config, serverOptions ...grpctransport.ServerOption) {
	endpoints := NewEndpoints()
	
	// Mechanical domain.
	errc := make(chan error)
	
	// Interrupt handler.
	go handlers.InterruptHandler(errc)
	
	// Debug listener.
	go func() {
		log.Println("transport", "debug", "addr", cfg.DebugAddr)
		
		m := http.NewServeMux()
		m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		
		errc <- http.ListenAndServe(cfg.DebugAddr, m)
	}()
	
	// HTTP transport.
	go func() {
		log.Println("transport", "HTTP", "addr", cfg.HTTPAddr)
		h := svc.MakeHTTPHandler(endpoints)
		errc <- http.ListenAndServe(cfg.HTTPAddr, h)
	}()
	
	// gRPC transport.
	go func() {
		log.Println("transport", "gRPC", "addr", cfg.GRPCAddr)
		ln, err := net.Listen("tcp", cfg.GRPCAddr)
		if err != nil {
			errc <- err
			return
		}
		
		srv := svc.MakeGRPCServer(endpoints, serverOptions...)
		s := grpc.NewServer()
		pb.RegisterBookServer(s, srv)
		
		errc <- s.Serve(ln)
	}()
	
	// Run!
	log.Println("exit", <-errc)
}
