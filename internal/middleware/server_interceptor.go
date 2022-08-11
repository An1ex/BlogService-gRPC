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
	// 获取gRPC Client Context中的附带信息
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	// 解析信息到SpanContext
	tracer := opentracing.GlobalTracer()
	parentSpanContext, _ := tracer.Extract(opentracing.TextMap, metatext.MetadataTextMap{MD: md})

	// 设置子Span的信息和标签，生成子Span
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

// 单HTTP请求的追踪链路还未完成
//func HttpTracingInject(handler http.Handler) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		// 因为是直接浏览器直接发HTTP请求，所以这里模拟HTTP网关应该做的初始root Span并注入HTTP Header
//		spanOpts := []opentracing.StartSpanOption{
//			opentracing.Tag{
//				Key:   string(ext.Component),
//				Value: "HTTP",
//			},
//		}
//		span := opentracing.StartSpan(r.Method+r.URL.Host+r.URL.Path, spanOpts...)
//		defer span.Finish()
//		tracer := opentracing.GlobalTracer()
//		tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
//		handler.ServeHTTP(w, r)
//	}
//}
//
//func HttpTracing(handler http.Handler) http.HandlerFunc {
//	// 如果不是用gRPC-gateway，则由中间的API服务提取HTTP Header中的SpanContext生成新的childSpan，再inject到gRPC Context的metadata中
//	tracer := opentracing.GlobalTracer()
//	// 提取HTTP Header中的SpanContext生成新的childSpan
//	parentSpanContext, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(reqHttp.Header))
//	// 设置子Span的信息和标签，生成子Span
//	spanOpts := []opentracing.StartSpanOption{
//		opentracing.Tag{
//			Key:   string(ext.Component),
//			Value: "API",
//		},
//		opentracing.ChildOf(parentSpanContext),
//	}
//
//	// 看源码
//	// opentracing.StartSpanFromContext() == opentracing.StartSpanFromContextWithTracer(ctx,GlobalTracer(),..)
//	// opentracing.StartSpan() == globalTracer.tracer.StartSpan() == tracer.StartSpan()
//	span := tracer.StartSpan("HTTP", spanOpts...)
//	defer span.Finish()
//
//	// 生成新的下文
//	md := metadata.New(nil)
//	tracer.Inject(span.Context(), opentracing.TextMap, md)
//	ctx := metadata.NewOutgoingContext(context.Background(), md)
//	newCtx := opentracing.ContextWithSpan(ctx, span)
//
//}
