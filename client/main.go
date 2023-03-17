package main

import (
	"context"
	"fmt"
	pb "github.com/lackone/grpc-study/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	conn, err := getGrpcClient(context.Background(), "127.0.0.1:8080", nil)
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
