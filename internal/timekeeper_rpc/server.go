package timekeeper_rpc

import (
	"log"
	"net"

	grpc "google.golang.org/grpc"

	"github.com/Quard/timekeeper/internal/storage"
)

type Opts struct {
	Bind string
}

type TimeKeeperRPC struct {
	opts       Opts
	listener   net.Listener
	grpcServer *grpc.Server
	storage    storage.Storage
}

func NewTimeKeeperRPC(opts Opts, stor storage.Storage) TimeKeeperRPC {
	srv := TimeKeeperRPC{
		opts:    opts,
		storage: stor,
	}

	listener, err := net.Listen("tcp", srv.opts.Bind)
	if err != nil {
		log.Fatal(err)
	}
	srv.listener = listener

	var grpcOpts []grpc.ServerOption
	srv.grpcServer = grpc.NewServer(grpcOpts...)
	RegisterTimeKeeperRPCServer(srv.grpcServer, &srv)

	return srv
}

func (srv TimeKeeperRPC) Run() {
	log.Printf("INTERNAL gRPC server listen on: %s", srv.opts.Bind)
	srv.grpcServer.Serve(srv.listener)
}
