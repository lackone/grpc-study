package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lackone/grpc-study/pkg/errcode"
	"github.com/lackone/grpc-study/pkg/service"
	pb "github.com/lackone/grpc-study/proto"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
)

type Server struct {
	endpoint   string
	tcpListen  net.Listener
	grpcListen net.Listener
	httpListen net.Listener

	cMux      cmux.CMux
	regHttp   registerFunc
	regGrpc   registerFunc
	regGrpcGw registerFunc

	gwMux *runtime.ServeMux
}

type Option func(*Server)

type registerFunc func(ctx context.Context, s *Server)

func WithEndpoint(endpoint string) Option {
	return func(s *Server) {
		s.endpoint = endpoint
	}
}

func WithHttp(fn registerFunc) Option {
	return func(s *Server) {
		s.regHttp = fn
	}
}

func WithGrpc(fn registerFunc) Option {
	return func(s *Server) {
		s.regGrpc = fn
	}
}

func WithGrpcGw(fn registerFunc) Option {
	return func(s *Server) {
		s.regGrpcGw = fn
	}
}

func NewServer(opts ...Option) (*Server, error) {
	s := &Server{}

	for _, fn := range opts {
		fn(s)
	}

	listen, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		return nil, err
	}

	s.tcpListen = listen
	s.cMux = cmux.New(listen)

	return s, nil
}

func (s *Server) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	s.grpcListen = s.cMux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	s.httpListen = s.cMux.Match(cmux.HTTP1Fast())

	go func() {
		s.regGrpc(ctx, s)
	}()

	go func() {
		s.regGrpcGw(ctx, s)
		s.regHttp(ctx, s)
	}()

	return s.cMux.Serve()
}

func main() {
	s, err := NewServer(
		WithEndpoint("127.0.0.1:8080"),
		WithHttp(func(ctx context.Context, s *Server) {
			mux := http.NewServeMux()

			if s.gwMux != nil {
				mux.Handle("/", s.gwMux)
			}

			mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("test"))
			})

			server := &http.Server{
				Handler: mux,
			}

			server.Serve(s.httpListen)
		}),
		WithGrpc(func(ctx context.Context, s *Server) {
			server := grpc.NewServer()
			pb.RegisterArticleServiceServer(server, &service.ArticleService{})
			reflection.Register(server)
			server.Serve(s.grpcListen)
		}),
		WithGrpcGw(func(ctx context.Context, s *Server) {
			s.gwMux = runtime.NewServeMux(
				runtime.WithErrorHandler(errcode.GrpcGatewayError),
			)
			pb.RegisterArticleServiceHandlerFromEndpoint(ctx, s.gwMux, s.endpoint, []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			})
		}),
	)

	if err != nil {
		panic(err)
	}

	s.Start()
}
