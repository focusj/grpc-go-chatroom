package main

import (
	chantroom "github.com/focusj/grpc-go-chatroom/chatroom"
	"github.com/focusj/grpc-go-chatroom/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

func main() {
	listen, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	keepaliveOpts := keepalive.ServerParameters{
		MaxConnectionIdle:     10 * time.Second,
		Time:                  1 * time.Second,
		Timeout:               1 * time.Second,
	}

	server := grpc.NewServer(grpc.KeepaliveParams(keepaliveOpts))
	chantroom.RegisterChatRoomServer(server, service.New())
	err = server.Serve(listen)
	if err != nil {
		grpclog.Fatalf("server failed: %+v", err)
	}
}
