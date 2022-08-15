package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"BlogService-gRPC/internal/middleware"
	"BlogService-gRPC/pb"
	"BlogService-gRPC/pkg/etcd"
	"BlogService-gRPC/server"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// HTTP和gRPC同端口通过h2c手动分流实现
// http/gRPC:9002

var sPort string

func init() {
	flag.StringVar(&sPort, "port", "9002", "启动端口号")
	flag.Parse()
}

const ServiceName = "tag-service"

func main() {
	tracer, closer, err := NewJaegerTracer("blog-service-grpc-server", "localhost:6831")
	if err != nil {
		log.Fatalf("Jaeger init err: %s", err.Error())
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	err = RunServer(sPort)
	if err != nil {
		log.Fatalf("Run Serve err: %s", err.Error())
	}
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func RunServer(port string) error {
	httpMux := runHttpServer()
	grpcS := runGrpcServer()
	gatewayMux := runGrpcGatewayServer()
	httpMux.Handle("/", gatewayMux)

	etcdReg, err := etcd.NewEtcdRegister()
	if err != nil {
		return err
	}
	defer etcdReg.Close()
	etcdReg.RegisterServer("/etcd/blog-service/grpc/"+ServiceName, "localhost:"+port, 5)

	return http.ListenAndServe(":"+port, grpcHandlerFunc(grpcS, httpMux))
}

func runHttpServer() *http.ServeMux {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})
	return serveMux
}

func runGrpcServer() *grpc.Server {
	opts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(
		middleware.AccessLog,
		middleware.ErrorLog,
		middleware.Recovery,
		middleware.ServerTracing,
	)}
	s := grpc.NewServer(opts...)
	pb.RegisterTagServiceServer(s, &server.TagServer{})
	reflection.Register(s)
	return s
}

func runGrpcGatewayServer() *runtime.ServeMux {
	endpoint := ":" + sPort
	gwmux := runtime.NewServeMux()
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	_ = pb.RegisterTagServiceHandlerFromEndpoint(context.Background(), gwmux, endpoint, dopts)
	return gwmux
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
