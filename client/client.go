package main

import (
	"context"
	"log"

	"BlogService-gRPC/internal/middleware"
	"BlogService-gRPC/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const PORT = "9002"

type Auth struct {
	AppKey    string
	AppSecret string
}

func (a *Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"app_key": a.AppKey, "app_secret": a.AppSecret}, nil
}

func (a *Auth) RequireTransportSecurity() bool {
	return false
}

func main() {
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
	conn, err := grpc.Dial(":"+PORT, opts...)
	if err != nil {
		log.Fatalf("grpc.Dial err: %v", err)
	}
	defer conn.Close()

	client := pb.NewTagServiceClient(conn)
	resp, err := client.GetTagList(context.Background(), &pb.GetTagListRequest{
		Name:  "",
		State: 1,
	})
	log.Printf("resp:%v", resp)
}
