package unsure

import (
	"context"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/interceptors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"google.golang.org/grpc"
)

func NewClient(url string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithChainUnaryInterceptor(
		interceptors.UnaryClientInterceptor, unaryClientInterceptor))
	opts = append(opts, grpc.WithChainStreamInterceptor(
		interceptors.StreamClientInterceptor, streamClientInterceptor))
	opts = append(opts, grpc.WithInsecure())

	return grpc.Dial(url, opts...)
}

// Server wraps a gRPC server.
type Server struct {
	listener   net.Listener
	grpcServer *grpc.Server
}

// NewServer returns a new `Server`.
func NewServer(address string) (*Server, error) {
	if address == "" {
		return nil, errors.New("no address provided")
	}

	var srv Server

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	srv.listener = listener

	var opts []grpc.ServerOption

	opts = append(opts, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		interceptors.StreamServerInterceptor, streamServerInterceptor)))
	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		interceptors.UnaryServerInterceptor, unaryServerInterceptor)))

	srv.grpcServer = grpc.NewServer(opts...)
	return &srv, nil
}

// Listener returns the server's net.Listener.
func (srv *Server) Listener() net.Listener {
	return srv.listener
}

// GRPCServer returns the server's grpc.Server.
func (srv *Server) GRPCServer() *grpc.Server {
	return srv.grpcServer
}

// Stop stops the gRPC server.
func (srv *Server) Stop() {
	srv.grpcServer.GracefulStop()
}

// ServeForever listens for gRPC requests.
func (srv *Server) ServeForever() error {
	log.Info(nil, "grpctls: ServeForever listening", j.KV("addr", srv.listener.Addr()))
	return srv.grpcServer.Serve(srv.listener)
}

// unaryClientInterceptor returns an interceptor that tempts fate.
func unaryClientInterceptor(ctx context.Context, method string,
	req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	if err := temptCtx(ctx); err != nil {
		return err
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// streamClientInterceptor returns an interceptor that tempts fate.
func streamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {

	cs, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}

	return &clientStream{ClientStream: cs, ctx: ctx}, nil
}

// unaryServerInterceptor returns an interceptor that tempts fate.
func unaryServerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	interface{}, error) {
	ctx = ContextWithFate(ctx, DefaultFateP()) // Server injects default fate.
	if err := temptCtx(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

// streamServerInterceptor returns an interceptor that tempts fate.
func streamServerInterceptor(srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	return handler(srv, &serverStream{ServerStream: ss})
}

// serverStream is a wrapper of a grpc.ServerStream implementation that tempts fate on each msg.
type serverStream struct {
	grpc.ServerStream
}

func (ss *serverStream) SendMsg(m interface{}) error {
	ctx := ContextWithFate(ss.Context(), DefaultFateP()) // Server injects default fate.
	if err := temptCtx(ctx); err != nil {
		return err
	}
	return ss.ServerStream.SendMsg(m)
}

func (ss *serverStream) RecvMsg(m interface{}) error {
	ctx := ContextWithFate(ss.Context(), DefaultFateP()) // Server injects default fate.
	if err := temptCtx(ctx); err != nil {
		return err
	}
	return ss.ServerStream.RecvMsg(m)
}

// clientStream is a wrapper of a grpc.ClientStream implementation that tempts fate on each msg.
type clientStream struct {
	grpc.ClientStream
	ctx context.Context
}

func (ss *clientStream) SendMsg(m interface{}) error {
	if err := temptCtx(ss.ctx); err != nil {
		return err
	}
	return ss.ClientStream.SendMsg(m)
}

func (ss *clientStream) RecvMsg(m interface{}) error {
	if err := temptCtx(ss.ctx); err != nil {
		return err
	}
	return ss.ClientStream.RecvMsg(m)
}
