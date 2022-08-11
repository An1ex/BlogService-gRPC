package middleware

import (
	"context"
	"time"

	"BlogService-gRPC/pkg/metatext"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func defaultContextTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		defaultTimeout := 60 * time.Second
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	}
	return ctx, cancel
}

func ContextTimeout(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := defaultContextTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func ContextRetry() grpc.UnaryClientInterceptor {
	return grpc_retry.UnaryClientInterceptor(
		grpc_retry.WithMax(2),
		grpc_retry.WithCodes(
			codes.Unknown,
			codes.Internal,
			codes.DeadlineExceeded,
		),
	)
}

func ClientTracing(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var spanOpts []opentracing.StartSpanOption

	// 获取上文信息，如果有，加入spanOpts
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		spanOpts = append(spanOpts, opentracing.ChildOf(parentSpan.Context()))
	}

	// 设置当前Span的信息和标签，生成Span
	spanOpts = append(spanOpts, opentracing.Tag{Key: string(ext.Component), Value: "gRPC"}, ext.SpanKindRPCClient)
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan(method, spanOpts...)
	defer span.Finish()

	// 获取gRPC Client的附带信息
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	// 注入信息，生成新的下文
	_ = tracer.Inject(span.Context(), opentracing.TextMap, metatext.MetadataTextMap{MD: md})
	newCtx := opentracing.ContextWithSpan(metadata.NewOutgoingContext(ctx, md), span)

	return invoker(newCtx, method, req, reply, cc, opts...)
}
