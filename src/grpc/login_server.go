package grpc

import (
	"github.com/opentibiabr/login-server/src/grpc/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

type LoginServer struct {
	proto.LoginServer
}

func (ls *LoginServer) Login(stream proto.LoginServer) error {
	return nil
}

func Run(addr string) {
	c, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Fatal(grpc.NewServer().Serve(c))
}
