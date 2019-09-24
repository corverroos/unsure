package main

import (
	"flag"
	"net/http"

	"github.com/corverroos/unsure"
	loser_ops "github.com/corverroos/unsure/loser/ops"
	"github.com/corverroos/unsure/loser/state"
	"github.com/luno/jettison/errors"
)

var (
	httpAddress = flag.String("http_address", ":23047", "loser healthcheck address")
	grpcAddress = flag.String("grpc_address", ":23048", "loser grpc server address")
)

func main() {
	unsure.Bootstrap()

	s, err := state.New()
	if err != nil {
		unsure.Fatal(errors.Wrap(err, "new state error"))
	}

	go serveGRPCForever(s)

	loser_ops.StartLoops(s)

	http.HandleFunc("/health", makeHealthCheckHandler())
	go unsure.ListenAndServeForever(*httpAddress, nil)

	unsure.WaitForShutdown()
}

func serveGRPCForever(s *state.State) {
	grpcServer, err := unsure.NewServer(*grpcAddress)
	if err != nil {
		unsure.Fatal(errors.Wrap(err, "new grpctls server"))
	}

	//loserSrv := loser_server.New(s)
	//loserpb.RegisterloserServer(grpcServer.GRPCServer(), loserSrv)

	unsure.RegisterNoErr(func() {
		//loserSrv.Stop()
		grpcServer.Stop()
	})

	unsure.Fatal(grpcServer.ServeForever())
}

func makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}
