package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"

	"BlogService-gRPC/pb"
	"BlogService-gRPC/server"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var sPort string

func init() {
	flag.StringVar(&sPort, "port", "9002", "启动端口号")
	flag.Parse()
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
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, &server.TagServer{})
	reflection.Register(s)
	return s
}

func runGrpcGatewayServer() *runtime.ServeMux {
	endpoint := "0.0.0.0:" + sPort
	gwmux := runtime.NewServeMux()
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	_ = pb.RegisterTagServiceHandlerFromEndpoint(context.Background(), gwmux, endpoint, dopts)
	return gwmux
}

func main() {
	err := RunServer(sPort)
	if err != nil {
		log.Fatalf("Run Serve err: %v", err)
	}
}
