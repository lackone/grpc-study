package service

import (
	"context"
	"github.com/lackone/grpc-study/pkg/db"
	"github.com/lackone/grpc-study/pkg/errcode"
	"github.com/lackone/grpc-study/pkg/model"
	pb "github.com/lackone/grpc-study/proto"
)

type ArticleService struct {
	pb.UnimplementedArticleServiceServer
}

func (a *ArticleService) GetArticleList(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	page := req.GetPage()
	size := req.GetSize()

	if page <= 0 || size <= 0 {
		return nil, errcode.TogRPCError(errcode.ErrorGetArticleListRequestFail)
	}

	offset := (page - 1) * size

	var articles []*pb.Article
	db.DB.Model(&model.Article{}).Select("id, title").Order("id desc").Limit(int(size)).Offset(int(offset)).Find(&articles)

	var totalRows int32
	db.DB.Model(&model.Article{}).Select("count(*) as cnt").Pluck("cnt", &totalRows)

	return &pb.GetArticleResponse{
		List: articles,
		Pager: &pb.Pager{
			Page:      page,
			Size:      size,
			TotalRows: totalRows,
		},
	}, nil
}
