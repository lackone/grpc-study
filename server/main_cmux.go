package main

import (
	"context"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lackone/grpc-study/pkg/errcode"
	"github.com/lackone/grpc-study/pkg/middleware"
	"github.com/lackone/grpc-study/pkg/service"
	"github.com/lackone/grpc-study/pkg/swagger"
	"github.com/lackone/grpc-study/pkg/tracer"
	pb "github.com/lackone/grpc-study/proto"
	"github.com/soheilhy/cmux"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
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
	tp, err := tracer.InitTracerProvider("127.0.0.1", "6831", "grpc-server")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()

	s, err := NewServer(
		WithEndpoint("127.0.0.1:8080"),
		WithHttp(func(ctx context.Context, s *Server) {
			mux := http.NewServeMux()

			if s.gwMux != nil {
				mux.Handle("/", s.gwMux)
			}

			//配置swagger-ui
			prefix := "/swagger-ui/"
			fileServer := http.FileServer(&assetfs.AssetFS{
				Asset:    swagger.Asset,
				AssetDir: swagger.AssetDir,
				Prefix:   "third_party/swagger-ui",
			})
			mux.Handle(prefix, http.StripPrefix(prefix, fileServer))

			//读取swagger.json文件
			mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "swagger.json") {
					http.NotFound(w, r)
					return
				}
				p := strings.TrimPrefix(r.URL.Path, "/swagger/")
				p = path.Join("proto", p)
				http.ServeFile(w, r, p)
			})

			mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("test"))
			})

			server := &http.Server{
				Handler: mux,
			}

			server.Serve(s.httpListen)
		}),
		WithGrpc(func(ctx context.Context, s *Server) {
			//拦截器
			opts := []grpc.ServerOption{
				//添加拦载器
				grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
					otelgrpc.UnaryServerInterceptor(),
					func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
						//获取客户端传过来的metadata
						md, ok := metadata.FromIncomingContext(ctx)
						if !ok {
							return resp, status.Error(codes.Unauthenticated, "token不正确")
						}

						fmt.Println(md)

						return handler(ctx, req)
					},
					middleware.AccessLog,
					middleware.Error,
					middleware.Recovery,
				)),
				grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
					otelgrpc.StreamServerInterceptor(),
				)),
			}

			server := grpc.NewServer(opts...)
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

// 一种类型的拦截器只允许设置一个，通过grpc_middleware可以设置多个
func TestInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Println("test调用之前")

	resp, err := handler(ctx, req)

	fmt.Println("test调用之后")

	return resp, err
}

// 一种类型的拦截器只允许设置一个，通过grpc_middleware可以设置多个
func HelloInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Println("hello调用之前")

	resp, err := handler(ctx, req)

	fmt.Println("hello调用之后")

	return resp, err
}
