package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"

	"BlogService-gRPC/pb"
	"BlogService-gRPC/server"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// HTTP和gRPC同端口用cmux实现
// http/gRPC:9002

var cPort string

func init() {
	flag.StringVar(&cPort, "port", "9002", "启动端口号")
	flag.Parse()
}

func RunTCPServer(port string) (net.Listener, error) {
	return net.Listen("tcp", ":"+port)
}

func RunGrpcServer() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, &server.TagServer{})
	reflection.Register(s)
	return s
}

func RunHttpServer(port string) *http.Server {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})

	endpoint := "localhost:" + port
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	_ = pb.RegisterTagServiceHandlerFromEndpoint(context.Background(), gwmux, endpoint, opts)
	serveMux.Handle("/", gwmux)

	return &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}
}

func main() {
	l, err := RunTCPServer(cPort)
	if err != nil {
		log.Fatalf("Run TCP Server err: %v", err)
	}
	m := cmux.New(l)
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := m.Match(cmux.HTTP1Fast())
	grpcS := RunGrpcServer()
	httpS := RunHttpServer(cPort)
	go grpcS.Serve(grpcL)
	go httpS.Serve(httpL)
	err = m.Serve()
	if err != nil {
		log.Fatalf("Run Serve err: %v", err)
	}
}
