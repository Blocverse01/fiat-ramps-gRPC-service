package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"

	pb "github.com/blocverse01/globachain-ramps-grpc-service/fiatramps"
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
