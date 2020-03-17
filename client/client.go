package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"time"

	pb "github.com/focusj/grpc-go-chatroom/chatroom"
)

func chat(sender int64, stream pb.ChatRoom_ChatClient) {
	for i := 0; i < 10; i++ {
		message := pb.Message{
			Id:       1,
			GroupId:  1,
			Sender:   sender,
			Content:  "hello",
			Type:     0,
			SendTime: time.Now().UnixNano(),
		}
		err := stream.Send(&message)
		if err != nil {
			grpclog.Info(err)
		}
	}

}

func main() {
	conn, err := grpc.Dial("localhost:8888", grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalf("dial failed: %s", err)
	}
	defer conn.Close()

	client := pb.NewChatRoomClient(conn)

	stream, err := client.Chat(context.Background())
	if err != nil {
		grpclog.Fatalf("chat failed: %s", err)
	}

	chat(1, stream)

	go func() {
		for {
			msg, err := stream.Recv()
			fmt.Println(msg)
			if err != nil {
				grpclog.Error(err)
			}
			grpclog.Info("receive from server: %+v", msg)
		}
	}()

	time.Sleep(100 * time.Second)

}
