// Version: 574fb16d86
// Version Date: 2019年 04月 12日 星期五 00:42:59 UTC

package svc

// This file provides server-side bindings for the gRPC transport.
// It utilizes the transport/grpc.Server.

import (
	"context"
	"net/http"
	
	"google.golang.org/grpc/metadata"
	
	grpctransport "github.com/go-kit/kit/transport/grpc"
	
	// This Service
	pb "github.com/fanioc/go-poetryminapp/services/book"
)

// MakeGRPCServer makes a set of endpoints available as a gRPC BookServer.
func MakeGRPCServer(endpoints Endpoints, serverOptions ...grpctransport.ServerOption) pb.BookServer {
	serverOptions = append(serverOptions, grpctransport.ServerBefore(metadataToContext))
	
	return &grpcServer{
		// book
		
		getbookinfo: grpctransport.NewServer(
			endpoints.GetBookInfoEndpoint,
			DecodeGRPCGetBookInfoRequest,
			EncodeGRPCGetBookInfoResponse,
			serverOptions...,
		),
		getbooklist: grpctransport.NewServer(
			endpoints.GetBookListEndpoint,
			DecodeGRPCGetBookListRequest,
			EncodeGRPCGetBookListResponse,
			serverOptions...,
		),
	}
}

// grpcServer implements the BookServer interface
type grpcServer struct {
	getbookinfo grpctransport.Handler
	getbooklist grpctransport.Handler
}

// Methods for grpcServer to implement BookServer interface

func (s *grpcServer) GetBookInfo(ctx context.Context, req *pb.BookInfoParams) (*pb.BookInfo, error) {
	_, rep, err := s.getbookinfo.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.BookInfo), nil
}

func (s *grpcServer) GetBookList(ctx context.Context, req *pb.BookListParams) (*pb.BookList, error) {
	_, rep, err := s.getbooklist.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.BookList), nil
}

// Server Decode

// DecodeGRPCGetBookInfoRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC getbookinfo request to a user-domain getbookinfo request. Primarily useful in a server.
func DecodeGRPCGetBookInfoRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.BookInfoParams)
	return req, nil
}

// DecodeGRPCGetBookListRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC getbooklist request to a user-domain getbooklist request. Primarily useful in a server.
func DecodeGRPCGetBookListRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.BookListParams)
	return req, nil
}

// Server Encode

// EncodeGRPCGetBookInfoResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain getbookinfo response to a gRPC getbookinfo reply. Primarily useful in a server.
func EncodeGRPCGetBookInfoResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.BookInfo)
	return resp, nil
}

// EncodeGRPCGetBookListResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain getbooklist response to a gRPC getbooklist reply. Primarily useful in a server.
func EncodeGRPCGetBookListResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.BookList)
	return resp, nil
}

// Helpers

func metadataToContext(ctx context.Context, md metadata.MD) context.Context {
	for k, v := range md {
		if v != nil {
			// The key is added both in metadata format (k) which is all lower
			// and the http.CanonicalHeaderKey of the key so that it can be
			// accessed in either format
			ctx = context.WithValue(ctx, k, v[0])
			ctx = context.WithValue(ctx, http.CanonicalHeaderKey(k), v[0])
		}
	}
	
	return ctx
}
