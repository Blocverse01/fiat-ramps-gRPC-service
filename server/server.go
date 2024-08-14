package main

import (
	"flag"
	"fmt"
	pb "github.com/Blocverse01/globachain-ramps-gRPC-service/fiatramps"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	port = flag.Int("port", 5005, "The server port")
)

type fiatRampsServer struct {
	pb.UnimplementedFiatRampsServer
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
}
