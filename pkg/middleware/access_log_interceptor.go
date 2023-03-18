package middleware

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"time"
)

func AccessLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestLog := "access request log: method: %s, begin_time: %d, request: %v\n"
	beginTime := time.Now().Local().Unix()
	fmt.Printf(requestLog, info.FullMethod, beginTime, req)

	resp, err := handler(ctx, req)

	responseLog := "access response log: method: %s, begin_time: %d, end_time: %d, response: %v\n"
	endTime := time.Now().Local().Unix()
	fmt.Printf(responseLog, info.FullMethod, beginTime, endTime, resp)
	return resp, err
}
