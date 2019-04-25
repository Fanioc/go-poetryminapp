package tracer

import (
	"fmt"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/openzipkin/zipkin-go"
	reporter "github.com/openzipkin/zipkin-go/reporter/http"
	"io"
)

func RegisterZipkinTracer(zipkinReportAddr string) (kitgrpc.ClientOption, io.Closer) {
	{
		//创建zipkin上报管理器
		reporte := reporter.NewReporter(zipkinReportAddr)
		
		//创建trace跟踪器
		zkTracer, err := zipkin.NewTracer(reporte)
		
		if err != nil {
			fmt.Println("err Tracer :" + err.Error())
			return nil, nil
		}
		
		//添加grpc请求的before after finalizer 事件对应要处理的trace操作方法
		return kitzipkin.GRPCClientTrace(zkTracer), reporte //运行结束，关闭上报管理器的for-select协程
	}
}
