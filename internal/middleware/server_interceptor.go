package middleware

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"BlogService-gRPC/pkg/metatext"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func AccessLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestLog := "access request log: method: %s, begin_time: %d, request: %v"
	beginTime := time.Now().Local().Unix()
	log.Printf(requestLog, info.FullMethod, beginTime, req)
	resp, err := handler(ctx, req)

	responseLog := "access response log: method: %s, begin_time: %d, end_time: %d, response: %v"
	endTime := time.Now().Local().Unix()
	log.Printf(responseLog, info.FullMethod, beginTime, endTime, resp)
	return resp, err
}

func ErrorLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		errLog := "error log: method: %s, error: %s"
		log.Printf(errLog, info.FullMethod, err.Error())
	}
	return resp, err
}

func Recovery(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if e := recover(); e != nil {
			recoveryLog := "recovery log: method: %s, message: %v, stack: %s"
			log.Printf(recoveryLog, info.FullMethod, e, string(debug.Stack()[:]))
		}
	}()
	return handler(ctx, req)
}

func ServerTracing(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 获取gRPC附带信息
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	// 解析信息到SpanContext
	tracer := opentracing.GlobalTracer()
	opentracing.SetGlobalTracer(tracer)
	parentSpanContext, _ := tracer.Extract(opentracing.TextMap, metatext.MetadataTextMap{MD: md})

	// 设置当前跨度的信息和标签，生成跨度
	spanOpts := []opentracing.StartSpanOption{
		opentracing.Tag{
			Key:   string(ext.Component),
			Value: "gRPC",
		},
		ext.SpanKindRPCServer,
		ext.RPCServerOption(parentSpanContext),
	}
	span := tracer.StartSpan(info.FullMethod, spanOpts...)
	defer span.Finish()

	// 生成新的下文
	newCtx := opentracing.ContextWithSpan(ctx, span)
	return handler(newCtx, req)
}
