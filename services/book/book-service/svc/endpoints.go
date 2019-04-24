// Version Date: 2019年 04月 12日 星期五 00:42:59 UTC

package svc

// This file contains methods to make individual endpoints from services,
// request and response types to serve those endpoints, as well as encoders and
// decoders for those types, for all of our supported transport serialization
// formats.

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"

	pb "github.com/fanioc/go-poetryminapp/services/book"
)

// Endpoints collects all of the endpoints that compose an add service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them into an Endpoints, and return it to the caller as a Service.
type Endpoints struct {
	GetBookInfoEndpoint endpoint.Endpoint
	GetBookListEndpoint endpoint.Endpoint
}

// Endpoints

func (e Endpoints) GetBookInfo(ctx context.Context, in *pb.BookInfoParams) (*pb.BookInfo, error) {
	response, err := e.GetBookInfoEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.BookInfo), nil
}

func (e Endpoints) GetBookList(ctx context.Context, in *pb.BookListParams) (*pb.BookList, error) {
	response, err := e.GetBookListEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.BookList), nil
}

// Make Endpoints

func MakeGetBookInfoEndpoint(s pb.BookServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.BookInfoParams)
		v, err := s.GetBookInfo(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeGetBookListEndpoint(s pb.BookServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.BookListParams)
		v, err := s.GetBookList(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

// WrapAllExcept wraps each Endpoint field of struct Endpoints with a
// go-kit/kit/endpoint.Middleware.
// Use this for applying a set of middlewares to every endpoint in the service.
// Optionally, endpoints can be passed in by name to be excluded from being wrapped.
// WrapAllExcept(middleware, "Status", "Ping")
func (e *Endpoints) WrapAllExcept(middleware endpoint.Middleware, excluded ...string) {
	included := map[string]struct{}{
		"GetBookInfo": struct{}{},
		"GetBookList": struct{}{},
	}

	for _, ex := range excluded {
		if _, ok := included[ex]; !ok {
			panic(fmt.Sprintf("Excluded endpoint '%s' does not exist; see middlewares/endpoints.go", ex))
		}
		delete(included, ex)
	}

	for inc, _ := range included {
		if inc == "GetBookInfo" {
			e.GetBookInfoEndpoint = middleware(e.GetBookInfoEndpoint)
		}
		if inc == "GetBookList" {
			e.GetBookListEndpoint = middleware(e.GetBookListEndpoint)
		}
	}
}

// LabeledMiddleware will get passed the endpoint name when passed to
// WrapAllLabeledExcept, this can be used to write a generic metrics
// middleware which can send the endpoint name to the metrics collector.
type LabeledMiddleware func(string, endpoint.Endpoint) endpoint.Endpoint

// WrapAllLabeledExcept wraps each Endpoint field of struct Endpoints with a
// LabeledMiddleware, which will receive the name of the endpoint. See
// LabeldMiddleware. See method WrapAllExept for details on excluded
// functionality.
func (e *Endpoints) WrapAllLabeledExcept(middleware func(string, endpoint.Endpoint) endpoint.Endpoint, excluded ...string) {
	included := map[string]struct{}{
		"GetBookInfo": struct{}{},
		"GetBookList": struct{}{},
	}

	for _, ex := range excluded {
		if _, ok := included[ex]; !ok {
			panic(fmt.Sprintf("Excluded endpoint '%s' does not exist; see middlewares/endpoints.go", ex))
		}
		delete(included, ex)
	}

	for inc, _ := range included {
		if inc == "GetBookInfo" {
			e.GetBookInfoEndpoint = middleware("GetBookInfo", e.GetBookInfoEndpoint)
		}
		if inc == "GetBookList" {
			e.GetBookListEndpoint = middleware("GetBookList", e.GetBookListEndpoint)
		}
	}
}
