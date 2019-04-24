package router

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/etcdv3"
	"github.com/kataras/muxie"
	"net/http"
	"time"
)

type RouteConfig struct {
	retryMax      int
	retryTimeout  time.Duration
	funcHttpHandl func(http.ResponseWriter, *http.Request)
}

type ServiceRoute interface {
	RegisterRouter(r *muxie.Mux, etcdclient *etcdv3.Client, logger *log.Logger)
}

func RegisterRouter(r *muxie.Mux, etcdclient *etcdv3.Client, logger *log.Logger) {
	BookRegisterRouter(r, etcdclient, logger)
}
