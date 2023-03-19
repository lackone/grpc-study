package main

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/lackone/grpc-study/pkg/middleware"
	"github.com/lackone/grpc-study/pkg/tracer"
	pb "github.com/lackone/grpc-study/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
)

type Auth struct {
}

func (a *Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"app_key":    "xxxx",
		"app_secret": "xxxx",
		"aaa":        "aaa",
		"bbb":        "bbb",
	}, nil
}

func (a *Auth) RequireTransportSecurity() bool {
	return false
}

func main() {
	tp, err := tracer.InitTracerProvider("127.0.0.1", "6831", "grpc-client")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()

	opts := []grpc.DialOption{
		//客户端的拦截器
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				otelgrpc.UnaryClientInterceptor(),

				middleware.UnaryContextTimeout(),

				//grpc重试操作
				grpc_retry.UnaryClientInterceptor(
					grpc_retry.WithMax(2),
					grpc_retry.WithCodes(
						codes.Unknown,
						codes.Internal,
						codes.DeadlineExceeded,
					),
				),
			),
		),
		grpc.WithStreamInterceptor(
			grpc_middleware.ChainStreamClient(
				otelgrpc.StreamClientInterceptor(),
				middleware.StreamContextTimeout(),
			),
		),
		//RPC方法做自定义认证
		grpc.WithPerRPCCredentials(&Auth{}),
	}

	ctx := context.Background()
	//newCtx := metadata.AppendToOutgoingContext(ctx, "app_id", "xxx", "app_key", "xxx")

	//md := metadata.New(map[string]string{"aaa": "aaa", "bbb": "bbb"})
	//newCtx := metadata.NewOutgoingContext(ctx, md)

	newCtx := metadata.AppendToOutgoingContext(ctx, "aaa", "aaa", "bbb", "bbb")

	conn, err := getGrpcClient(newCtx, "127.0.0.1:8080", opts)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	client := pb.NewArticleServiceClient(conn)
	list, err := client.GetArticleList(context.Background(), &pb.GetArticleRequest{Page: 1, Size: 4})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(list)
}

func getGrpcClient(ctx context.Context, addr string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return grpc.DialContext(ctx, addr, opts...)
}
