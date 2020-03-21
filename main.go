package main

import (
	"flag"
	"fmt"
	"github.com/focusj/grpc-go-chatroom/chatroom"
	"github.com/focusj/grpc-go-chatroom/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

var (
	host = flag.String("host", "0.0.0.0", "remote server address")
	port = flag.Int64("prot", 8888, "remote server port")
)

type loggingServerStream struct {
	grpc.ServerStream
	counter int64
}

func newLoggingServerStream(ss grpc.ServerStream) *loggingServerStream {
	return &loggingServerStream{ServerStream: ss}
}

func (ss *loggingServerStream) RecvMsg(m interface{}) error {
	grpclog.Infof("receive message: %T at %s", m, time.Now().Format(time.RFC3339))
	ss.counter += 1
	return ss.ServerStream.RecvMsg(m)
}

func (ss *loggingServerStream) SendMsg(m interface{}) error {
	grpclog.Infof("send message: %T at %s", m, time.Now().Format(time.RFC3339))
	return ss.ServerStream.SendMsg(m)
}

func tracingInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, newLoggingServerStream(ss))
	if err != nil {
		grpclog.Error(err)
	}
	return err
}

func main() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	keepaliveOpts := keepalive.ServerParameters{
		MaxConnectionIdle: 10 * time.Second,
		Time:              1 * time.Second,
		Timeout:           1 * time.Second,
	}

	server := grpc.NewServer(
		grpc.KeepaliveParams(keepaliveOpts),
		grpc.StreamInterceptor(tracingInterceptor),
	)
	chatroom.RegisterChatRoomServer(server, service.New())

	err = server.Serve(listen)
	if err != nil {
		grpclog.Fatalf("server failed: %+v", err)
	}

	grpclog.Infoln("grpc-go-chatroom is up")
}
