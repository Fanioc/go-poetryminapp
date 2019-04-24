// Package grpc provides a gRPC client for the Book service.
package grpc

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	// This Service
	pb "github.com/fanioc/go-poetryminapp/services/book"
	"github.com/fanioc/go-poetryminapp/services/book/book-service/svc"
)

// New returns an service backed by a gRPC client connection. It is the
// responsibility of the caller to dial, and later close, the connection.
func New(conn *grpc.ClientConn, options ...ClientOption) (svc.Endpoints, error) {
	var cc clientConfig

	for _, f := range options {
		err := f(&cc)
		if err != nil {
			return svc.Endpoints{}, errors.Wrap(err, "cannot apply option")
		}
	}

	clientOptions := []grpctransport.ClientOption{
		grpctransport.ClientBefore(
			contextValuesToGRPCMetadata(cc.headers)),
	}
	var getbookinfoEndpoint endpoint.Endpoint
	{
		getbookinfoEndpoint = grpctransport.NewClient(
			conn,
			"book.Book",
			"GetBookInfo",
			EncodeGRPCGetBookInfoRequest,
			DecodeGRPCGetBookInfoResponse,
			pb.BookInfo{},
			clientOptions...,
		).Endpoint()
	}

	var getbooklistEndpoint endpoint.Endpoint
	{
		getbooklistEndpoint = grpctransport.NewClient(
			conn,
			"book.Book",
			"GetBookList",
			EncodeGRPCGetBookListRequest,
			DecodeGRPCGetBookListResponse,
			pb.BookList{},
			clientOptions...,
		).Endpoint()
	}

	return svc.Endpoints{
		GetBookInfoEndpoint: getbookinfoEndpoint,
		GetBookListEndpoint: getbooklistEndpoint,
	}, nil
}

// GRPC Client Decode

// DecodeGRPCGetBookInfoResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC getbookinfo reply to a user-domain getbookinfo response. Primarily useful in a client.
func DecodeGRPCGetBookInfoResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.BookInfo)
	return reply, nil
}

// DecodeGRPCGetBookListResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC getbooklist reply to a user-domain getbooklist response. Primarily useful in a client.
func DecodeGRPCGetBookListResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.BookList)
	return reply, nil
}

// GRPC Client Encode

// EncodeGRPCGetBookInfoRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain getbookinfo request to a gRPC getbookinfo request. Primarily useful in a client.
func EncodeGRPCGetBookInfoRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.BookInfoParams)
	return req, nil
}

// EncodeGRPCGetBookListRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain getbooklist request to a gRPC getbooklist request. Primarily useful in a client.
func EncodeGRPCGetBookListRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.BookListParams)
	return req, nil
}

type clientConfig struct {
	headers []string
}

// ClientOption is a function that modifies the client config
type ClientOption func(*clientConfig) error

func CtxValuesToSend(keys ...string) ClientOption {
	return func(o *clientConfig) error {
		o.headers = keys
		return nil
	}
}

func contextValuesToGRPCMetadata(keys []string) grpctransport.ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		var pairs []string
		for _, k := range keys {
			if v, ok := ctx.Value(k).(string); ok {
				pairs = append(pairs, k, v)
			}
		}

		if pairs != nil {
			*md = metadata.Join(*md, metadata.Pairs(pairs...))
		}

		return ctx
	}
}
