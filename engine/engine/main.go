package main

import (
	"flag"
	"net/http"

	"github.com/corverroos/unsure"
	"github.com/corverroos/unsure/engine/enginepb"
	engine_ops "github.com/corverroos/unsure/engine/ops"
	engine_server "github.com/corverroos/unsure/engine/server"
	"github.com/corverroos/unsure/engine/state"
	"github.com/luno/jettison/errors"
)

var (
	httpAddress = flag.String("http_address", "localhost:12047", "engine healthcheck address")
	grpcAddress = flag.String("grpc_address", "localhost:12048", "engine grpc server address")
)

func main() {
	unsure.Bootstrap()

	s, err := state.New()
	if err != nil {
		unsure.Fatal(errors.Wrap(err, "new state error"))
	}

	go serveGRPCForever(s)

	engine_ops.StartLoops(s)

	http.HandleFunc("/health", makeHealthCheckHandler())
	go unsure.ListenAndServeForever(*httpAddress, nil)

	unsure.WaitForShutdown()
}

func serveGRPCForever(s *state.State) {
	grpcServer, err := unsure.NewServer(*grpcAddress)
	if err != nil {
		unsure.Fatal(errors.Wrap(err, "new grpctls server"))
	}

	engineSrv := engine_server.New(s)
	enginepb.RegisterEngineServer(grpcServer.GRPCServer(), engineSrv)

	unsure.RegisterNoErr(func() {
		engineSrv.Stop()
		grpcServer.Stop()
	})

	unsure.Fatal(grpcServer.ServeForever())
}

func makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}
