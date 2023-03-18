package middleware

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"runtime/debug"
)

func Recovery(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if e := recover(); e != nil {
			recoveryLog := "recovery log: method: %s, message: %v, stack: %s\n"
			fmt.Printf(recoveryLog, info.FullMethod, e, string(debug.Stack()[:]))
		}
	}()

	return handler(ctx, req)
}
