package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	"BlogService-gRPC/internal/middleware"
	"BlogService-gRPC/pb"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var port string

type Auth struct {
	AppKey    string
	AppSecret string
}

func init() {
	flag.StringVar(&port, "port", "9002", "启动端口号")
	flag.Parse()
}

func (a *Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"app_key": a.AppKey, "app_secret": a.AppSecret}, nil
}

func (a *Auth) RequireTransportSecurity() bool {
	return false
}

func NewJaegerTracer(serviceName, agentHostPort string) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agentHostPort,
		},
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, nil
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer, nil
}

func RunClient(port string) (*pb.GetTagListResponse, error) {
	opts := []grpc.DialOption{grpc.WithChainUnaryInterceptor(
		middleware.ContextTimeout,
		middleware.ContextRetry(),
		middleware.ClientTracing,
	)}

	auth := Auth{
		AppKey:    "an1ex_key",
		AppSecret: "an1ex_secret",
	}
	opts = append(opts, grpc.WithPerRPCCredentials(&auth))
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(":"+port, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := pb.NewTagServiceClient(conn)
	resp, err := client.GetTagList(context.Background(), &pb.GetTagListRequest{
		Name:  "",
		State: 1,
	})

	return resp, nil
}

func main() {
	tracer, closer, err := NewJaegerTracer("blog-service-grpc-client", "localhost:6831")
	if err != nil {
		log.Fatalf("Jaeger init err: %s", err.Error())
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	resp, err := RunClient(port)
	if err != nil {
		log.Fatalf("Run client err; %s", err.Error())
	}
	log.Printf("resp:%v", resp)
}
