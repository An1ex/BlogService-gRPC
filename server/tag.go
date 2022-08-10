package server

import (
	"context"
	"encoding/json"

	"BlogService-gRPC/pb"
	"BlogService-gRPC/pkg/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TagServer struct {
	pb.UnimplementedTagServiceServer
	auth *Auth
}

type Auth struct {
}

func (a *Auth) GetAppKey() string {
	return "an1ex_key"
}
func (a *Auth) GetAppSecret() string {
	return "an1ex_secret"
}

func (a *Auth) Authentication(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)
	var appKey, appSecret string
	if value, ok := md["app_key"]; ok {
		appKey = value[0]
	}
	if value, ok := md["app_secret"]; ok {
		appSecret = value[0]
	}
	if appKey != a.GetAppKey() || appSecret != a.GetAppSecret() {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}
	return nil
}

func (t *TagServer) GetTagList(ctx context.Context, r *pb.GetTagListRequest) (*pb.GetTagListResponse, error) {
	if err := t.auth.Authentication(ctx); err != nil {
		return nil, err
	}

	bApi := api.NewAPI("http://localhost:8080")
	body, err := bApi.GetTagList(ctx, r.GetName(), uint8(r.GetState()))
	if err != nil {
		return nil, err //errcode.TagRPCError(errcode.ERROR_GET_TAG_LIST_FAIL)
	}
	tagList := pb.GetTagListResponse{}
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		return nil, err //errcode.TagError(errcode.Fail)
	}
	return &tagList, nil
}
