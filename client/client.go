package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"math/rand"
	"time"

	pb "github.com/focusj/grpc-go-chatroom/chatroom"
)

var (
	uid = flag.Int64("user_id", 1, "user id")
)

func chat(sender int64, stream pb.ChatRoom_ChatClient) {
	for {
		message := pb.Message{
			Id:       rand.Int63(),
			GroupId:  1,
			Sender:   sender,
			Content:  fmt.Sprintf("greet from: %d", sender),
			Type:     0,
			SendTime: time.Now().UnixNano(),
		}
		err := stream.Send(&message)
		if err != nil {
			grpclog.Info(err)
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {
	flag.Parse()

	keepaliveParams := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             1 * time.Second,
		PermitWithoutStream: true,
	}
	conn, err := grpc.Dial(
		"127.0.0.1:8888",
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepaliveParams),
	)
	if err != nil {
		grpclog.Fatalf("dial failed: %s", err)
	}
	defer conn.Close()

	client := pb.NewChatRoomClient(conn)

	stream, err := client.Chat(context.Background())
	if err != nil {
		grpclog.Fatalf("chat failed: %s", err)
	}

	go chat(*uid, stream)

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				grpclog.Error(err)
				return
			}
			grpclog.Infof("receive a message from: %d, detail is: %s", msg.Sender, msg.Content)
		}
	}()

	time.Sleep(100 * time.Second)

}
