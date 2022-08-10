package server

import (
	"context"
	"encoding/json"

	"BlogService-gRPC/pb"
	"BlogService-gRPC/pkg/api"
)

type TagServer struct {
	pb.UnimplementedTagServiceServer
}

func (t *TagServer) GetTagList(ctx context.Context, r *pb.GetTagListRequest) (*pb.GetTagListResponse, error) {
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
