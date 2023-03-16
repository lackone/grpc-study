package main

import (
	"context"
	"github.com/lackone/grpc-study/pkg/db"
	"github.com/lackone/grpc-study/pkg/errcode"
	"github.com/lackone/grpc-study/pkg/model"
	pb "github.com/lackone/grpc-study/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
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

func main() {
	initTestData()

	server := grpc.NewServer()

	pb.RegisterArticleServiceServer(server, &ArticleService{})
	//注册了反射服务，使grpcurl可以使用
	reflection.Register(server)

	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	server.Serve(listen)
}

func initTestData() {
	db.DB.AutoMigrate(&model.Article{})

	db.DB.Create(&model.Article{
		Title: "aaa",
	})
	db.DB.Create(&model.Article{
		Title: "bbb",
	})
	db.DB.Create(&model.Article{
		Title: "ccc",
	})
}
