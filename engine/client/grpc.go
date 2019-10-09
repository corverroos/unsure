package client

import (
	"context"
	"flag"

	"github.com/corverroos/unsure"
	"github.com/corverroos/unsure/engine"
	pb "github.com/corverroos/unsure/engine/enginepb"
	"github.com/corverroos/unsure/engine/enginepb/protocp"
	"github.com/luno/reflex"
	"github.com/luno/reflex/reflexpb"
	"google.golang.org/grpc"
)

var addr = flag.String("engine_address", "", "host:port of engine gRPC service")

var _ engine.Client = (*client)(nil)

type client struct {
	address   string
	rpcConn   *grpc.ClientConn
	rpcClient pb.EngineClient
}

func IsEnabled() bool {
	return *addr != ""
}

type option func(*client)

func WithAddress(address string) option {
	return func(c *client) {
		c.address = address
	}
}

func New(opts ...option) (*client, error) {
	c := client{
		address: *addr,
	}
	for _, o := range opts {
		o(&c)
	}

	var err error
	c.rpcConn, err = unsure.NewClient(c.address)
	if err != nil {
		return nil, err
	}

	c.rpcClient = pb.NewEngineClient(c.rpcConn)

	return &c, nil
}

func (c *client) Ping(ctx context.Context) error {
	_, err := c.rpcClient.Ping(ctx, &pb.Empty{})
	return err
}

func (c *client) Stream(ctx context.Context, after string, opts ...reflex.StreamOption) (reflex.StreamClient, error) {
	sFn := reflex.WrapStreamPB(func(ctx context.Context,
		req *reflexpb.StreamRequest) (reflex.StreamClientPB, error) {
		return c.rpcClient.Stream(ctx, req)
	})
	return sFn(ctx, after, opts...)
}

func (c *client) StartMatch(ctx context.Context, team string, players int) error {
	_, err := c.rpcClient.StartMatch(ctx, &pb.StartMatchReq{
		Team:    team,
		Players: int64(players),
	})
	return err
}

func (c *client) JoinRound(ctx context.Context, team string, player string, roundID int64) (bool, error) {
	res, err := c.rpcClient.JoinRound(ctx, &pb.JoinRoundReq{
		Team:    team,
		Player:  player,
		RoundId: roundID,
	})
	if err != nil {
		return false, err
	}
	return res.Included, nil
}

func (c *client) CollectRound(ctx context.Context, team string, player string, roundID int64) (*engine.CollectRoundRes, error) {
	res, err := c.rpcClient.CollectRound(ctx, &pb.CollectRoundReq{
		Team:    team,
		Player:  player,
		RoundId: roundID,
	})
	if err != nil {
		return nil, err
	}
	return protocp.CollectRoundResFromProto(res), nil
}

func (c *client) SubmitRound(ctx context.Context, team string, player string, roundID int64, total int) error {
	_, err := c.rpcClient.SubmitRound(ctx, &pb.SubmitRoundReq{
		Team:    team,
		Player:  player,
		RoundId: roundID,
		Total:   int64(total),
	})
	return err
}
