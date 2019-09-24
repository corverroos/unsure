package server

import (
	"context"

	"github.com/corverroos/unsure"
	"github.com/corverroos/unsure/engine/db/events"
	pb "github.com/corverroos/unsure/engine/enginepb"
	"github.com/corverroos/unsure/engine/enginepb/protocp"
	"github.com/corverroos/unsure/engine/ops"
	"github.com/luno/reflex"
	"github.com/luno/reflex/reflexpb"
)

var _ pb.EngineServer = (*Server)(nil)

// Server implements the engine grpc server.
type Server struct {
	b       Backends
	rserver *reflex.Server
	stream  reflex.StreamFunc
}

// New returns a new server instance.
func New(b Backends) *Server {
	return &Server{
		b:       b,
		rserver: reflex.NewServer(),
		stream:  events.ToStream(b.EngineDB().DB),
	}
}

func (srv *Server) Stop() {
	srv.rserver.Stop()
}

func (srv *Server) Ping(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	return req, nil
}

func (srv *Server) Stream(req *reflexpb.StreamRequest, ss pb.Engine_StreamServer) error {
	return srv.rserver.Stream(srv.stream, req, ss)
}

func (srv *Server) StartMatch(ctx context.Context, req *pb.StartMatchReq) (*pb.Empty, error) {
	err := ops.StartMatch(unsure.ContextWithFate(ctx, unsure.DefaultFateP()), srv.b, req.Team, int(req.Players))
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (srv *Server) JoinRound(ctx context.Context, req *pb.JoinRoundReq) (*pb.JoinRoundRes, error) {
	included, err := ops.JoinRound(unsure.ContextWithFate(ctx, unsure.DefaultFateP()), srv.b, req.Team, req.Player)
	if err != nil {
		return nil, err
	}
	return &pb.JoinRoundRes{Included: included}, nil
}

func (srv *Server) CollectRound(ctx context.Context, req *pb.CollectRoundReq) (*pb.CollectRoundRes, error) {
	res, err := ops.CollectRound(unsure.ContextWithFate(ctx, unsure.DefaultFateP()),
		srv.b, req.Team, req.Player)
	if err != nil {
		return nil, err
	}
	return protocp.CollectRoundResToProto(res)
}

func (srv *Server) SubmitRound(ctx context.Context, req *pb.SubmitRoundReq) (*pb.Empty, error) {
	err := ops.SumbitRound(unsure.ContextWithFate(ctx, unsure.DefaultFateP()), srv.b, req.Team, req.Player, int(req.Total))
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}
