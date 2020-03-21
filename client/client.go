package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"math/rand"
	"time"

	pb "github.com/focusj/grpc-go-chatroom/chatroom"
)

var (
	uid = flag.Int64("user_id", 1, "user id")

	host = flag.String("host", "0.0.0.0", "remote server address")
	port = flag.Int64("prot", 8888, "remote server port")
)

func chat(sender int64, stream pb.ChatRoom_ChatClient) {
	for {
		message := pb.Message{
			Id:       rand.Int63(),
			GroupId:  1,
			Sender:   sender,
			Content:  fmt.Sprintf("chating from: %d", sender),
			Type:     0,
			SendTime: time.Now().UnixNano(),
		}
		err := stream.Send(&message)
		if err != nil {
			grpclog.Error(err)
		}
		time.Sleep(1 * time.Second)
	}
}

func tell(sender int64, client pb.ChatRoomClient) {
	for {
		message := pb.Message{
			Id:       rand.Int63(),
			GroupId:  1,
			Sender:   sender,
			Content:  fmt.Sprintf("telling from: %d", sender),
			Type:     0,
			SendTime: time.Now().UnixNano(),
		}

		var header metadata.MD // variable to store header and trailer
		_, err := client.Tell(context.Background(), &message, grpc.Header(&header))
		grpclog.Infof("unary msg from remote: %s as %s", header["Remote_Host"], header["timestamp"])

		if err != nil {
			grpclog.Errorf("telling failed, %s", err)
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
		fmt.Sprintf("%s:%d", *host, *port),
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

	go tell(*uid, client)

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				grpclog.Error(err)
				return
			}
			grpclog.Infof("streaming msg from: %d, detail is: [%s]", msg.Sender, msg.Content)
		}
	}()

	time.Sleep(100 * time.Second)

}
