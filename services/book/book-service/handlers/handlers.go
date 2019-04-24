package handlers

import (
	"context"
	"log"
	
	pb "github.com/fanioc/go-poetryminapp/services/book"
)

// NewService returns a naïve, stateless implementation of Service.
func NewService() pb.BookServer {
	return bookService{}
}

type bookService struct{}

// GetBookInfo implements Service.
func (s bookService) GetBookInfo(ctx context.Context, in *pb.BookInfoParams) (*pb.BookInfo, error) {
	var resp pb.BookInfo
	log.Println("call me : GetBookInfo")
	resp = pb.BookInfo{
		BookId:   1,
		BookName: "tese out",
	}
	return &resp, nil
}

// GetBookList implements Service.
func (s bookService) GetBookList(ctx context.Context, in *pb.BookListParams) (*pb.BookList, error) {
	var resp pb.BookList
	log.Println("call me : GetBookList")
	resp = pb.BookList{
		BookList: []*pb.BookInfo{
			{
				BookId:   0,
				BookName: "21天精通php",
			}, {
				BookId:   1,
				BookName: "tese out",
			},
		},
	}
	return &resp, nil
}
