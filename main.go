package main

import (
  "context"
  "google.golang.org/grpc"
  "net"
)

type BookServer struct{}

func (s *BookServer) GetBookList(ctx context.Context, in *BookListParams) (*BookList, error) {
  //请求列表时返回 书籍列表
  bl := new(BookList)
  bl.BookList = append(bl.BookList, &BookInfo{BookId: 1, BookName: "21天精通php"})
  bl.BookList = append(bl.BookList, &BookInfo{BookId: 2, BookName: "21天精通java"})
  return bl, nil
}

func (s *BookServer) GetBookInfo(ctx context.Context, in *BookInfoParams) (*BookInfo, error) {
  //请求详情时返回 书籍信息
  b := new(BookInfo)
  b.BookId = in.BookId
  b.BookName = "21天精通php"
  return b, nil
}

func main() {
  serviceAddress := ":50052"
  bookServer := new(BookServer)
  //创建tcp监听
  ls, _ := net.Listen("tcp", serviceAddress)
  //创建grpc服务
  gs := grpc.NewServer()
  //注册bookServer
  RegisterBookServer(gs, bookServer)
  //启动服务
  _ = gs.Serve(ls)
}
