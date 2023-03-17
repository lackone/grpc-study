package main

import (
	"context"
	"flag"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lackone/grpc-study/pkg/db"
	"github.com/lackone/grpc-study/pkg/errcode"
	"github.com/lackone/grpc-study/pkg/model"
	"github.com/lackone/grpc-study/pkg/service"
	pb "github.com/lackone/grpc-study/proto"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"net/http"
	"strings"
)

var port string

func init() {
	flag.StringVar(&port, "port", "8080", "启动端口号")
	flag.Parse()
}

func main() {
	initTestData()

	RunServer(port)
}

func RunServer(port string) error {
	httpMux := NewHttpServer()
	grpcServer := NewGrpcServer()
	gwMux := NewGrpcGatewayServer(port)

	httpMux.Handle("/", gwMux)
	return http.ListenAndServe(":"+port, grpcHandlerFunc(grpcServer, httpMux))
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

// grpc-gateway服务
func NewGrpcGatewayServer(port string) *runtime.ServeMux {
	endpoint := "127.0.0.1:" + port

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(errcode.GrpcGatewayError),
	)

	pb.RegisterArticleServiceHandlerFromEndpoint(context.Background(), mux, endpoint, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	})

	return mux
}

// http服务
func NewHttpServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	return mux
}

// grpc服务
func NewGrpcServer() *grpc.Server {
	opts := []grpc.ServerOption{}

	server := grpc.NewServer(opts...)

	//注册服务
	pb.RegisterArticleServiceServer(server, &service.ArticleService{})
	reflection.Register(server)

	return server
}

func initTestData() {
	db.DB.AutoMigrate(&model.Article{})
}
