package router

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/etcdv3"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/kataras/muxie"
	"net/http"
	"time"
)

type RouteConfig struct {
	retryMax     int
	retryTimeout time.Duration
	
	// Timeout 【请求超时的时间】
	// ErrorPercentThreshold【允许出现的错误比例】
	// SleepWindow【熔断开启多久尝试发起一次请求】
	// MaxConcurrentRequests【允许的最大并发请求数】
	// RequestVolumeThreshold 【波动期内的最小请求数，默认波动期10S】
	circuitbreaker hystrix.CommandConfig
	funcHttpHandl  func(http.ResponseWriter, *http.Request)
}

type ServiceRoute interface {
	RegisterRouter(r *muxie.Mux, etcdclient *etcdv3.Client, logger *log.Logger)
}

func RegisterRouter(r *muxie.Mux, etcdclient *consul.Client, logger *log.Logger, zkClientTrace *kitgrpc.ClientOption) {
	BookRegisterRouter(r, etcdclient, logger, zkClientTrace)
}
