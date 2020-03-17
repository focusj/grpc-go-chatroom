package main

import (
	chantroom "github.com/focusj/grpc-go-chatroom/chatroom"
	"github.com/focusj/grpc-go-chatroom/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
)

func main() {
	listen, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	chantroom.RegisterChatRoomServer(server, service.New())
	server.Serve(listen)
}
