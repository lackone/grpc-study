package middleware

import (
	"context"
	"fmt"
	"github.com/lackone/grpc-study/pkg/errcode"
	"google.golang.org/grpc"
)

func Error(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		errLog := "error log: method: %s, code: %v, message: %v, details: %v\n"
		s := errcode.FromError(err)
		fmt.Printf(errLog, info.FullMethod, s.Code(), s.Err().Error(), s.Details())
	}
	return resp, err
}
