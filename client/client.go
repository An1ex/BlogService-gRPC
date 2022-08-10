package main

import (
	"context"
	"log"

	"BlogService-gRPC/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const PORT = "9002"

func main() {
	conn, err := grpc.Dial(":"+PORT, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
