package errcode

import (
	"context"
	"encoding/json"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lackone/grpc-study/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type HttpError struct {
	Code    int32  `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func GrpcGatewayError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	pb := s.Proto()

	httpError := HttpError{
		Code:    int32(s.Code()),
		Message: s.Message(),
	}

	details := s.Details()
	for _, detail := range details {
		if v, ok := detail.(*proto.Error); ok {
			httpError.Code = v.Code
			httpError.Message = v.Message
		}
	}

	contentType := marshaler.ContentType(pb)
	w.Header().Set("Content-Type", contentType)

	resp, _ := json.Marshal(httpError)
	w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
	w.Write(resp)
}
