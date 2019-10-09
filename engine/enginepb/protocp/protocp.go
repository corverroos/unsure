package protocp

import (
	"github.com/corverroos/unsure/engine"
	pb "github.com/corverroos/unsure/engine/enginepb"
)

func CollectPlayerToProto(p *engine.CollectPlayer) *pb.CollectPlayer {
	return &pb.CollectPlayer{
		Name: p.Name,
		Part: int64(p.Part),
	}
}

func CollectPlayerFromProto(p *pb.CollectPlayer) *engine.CollectPlayer {
	return &engine.CollectPlayer{
		Name: p.Name,
		Part: int(p.Part),
	}
}

func CollectRoundResToProto(p *engine.CollectRoundRes) *pb.CollectRoundRes {
	var players []*pb.CollectPlayer
	for _, player := range p.Players {
		players = append(players, CollectPlayerToProto(&player))
	}

	return &pb.CollectRoundRes{
		Rank:    int64(p.Rank),
		Players: players,
	}
}

func CollectRoundResFromProto(p *pb.CollectRoundRes) *engine.CollectRoundRes {
	var players []engine.CollectPlayer
	for _, player := range p.Players {
		players = append(players, *CollectPlayerFromProto(player))
	}

	return &engine.CollectRoundRes{
		Rank:    int(p.Rank),
		Players: players,
	}
}
