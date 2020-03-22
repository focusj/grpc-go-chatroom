package service

import (
	"context"
	pb "github.com/focusj/grpc-go-chatroom/chatroom"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

type ChatServer struct {
	sync.RWMutex

	pb.UnimplementedChatRoomServer

	groups map[int64]*pb.Group

	channels map[int64]pb.ChatRoom_ChatServer
}

func (c *ChatServer) register(uid int64, srv pb.ChatRoom_ChatServer) {
	c.Lock()
	defer c.Unlock()
	grpclog.Infof("user of %d is login", uid)
	if _, exists := c.channels[uid]; !exists {
		c.channels[uid] = srv
	}
}

func (c *ChatServer) unRegister(uid int64) {
	c.Lock()
	defer c.Unlock()
	grpclog.Infof("user of: %d is logout", uid)
	delete(c.channels, uid)
}

func (c *ChatServer) deliver(groupId int64, msg *pb.Message) {
	group := c.groups[groupId]
	for _, uid := range group.Members {
		if ch, exists := c.channels[uid]; exists {
			timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
			host, _ := os.LookupEnv("HOSTNAME")
			headers := metadata.New(map[string]string{"Remote-Host": host, "Timestamp": timestamp})
			ch.SendHeader(headers)

			err := ch.Send(msg)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					if e.Code() == codes.Unavailable {
						c.unRegister(uid)
					}
				}
				grpclog.Errorf("deliver msg failed: %s", err)
			}
		}
	}
}

func New() *ChatServer {
	groups := make(map[int64]*pb.Group)
	groups[1] = &pb.Group{
		Id:      1,
		Name:    "grpc-go-chat-group",
		Members: []int64{1, 2, 3},
	}
	channels := make(map[int64]pb.ChatRoom_ChatServer)
	return &ChatServer{groups: groups, channels: channels}
}

func (c *ChatServer) Chat(srv pb.ChatRoom_ChatServer) error {
	var sender int64
	for {
		msg, err := srv.Recv()
		if err == io.EOF {
			// channel is closing by client
			if sender != 0 {
				c.unRegister(sender)
			}
			break
		}
		if err != nil {
			if grpcErr, ok := status.FromError(err); ok {
				if grpcErr.Code() == codes.Canceled {
					if sender != 0 {
						c.unRegister(sender)
					}
				}
			}
			grpclog.Errorf("chat error: %+v", err)
			return err
		}
		sender = msg.Sender

		c.register(msg.Sender, srv) // should be moved to login, just call once when a user login

		c.deliver(msg.GroupId, msg)
	}

	return status.Errorf(codes.Unimplemented, "method Chat not implemented")
}

func (c *ChatServer) Tell(ctx context.Context, req *pb.Message) (*pb.Empty, error) {
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	host, _ := os.LookupEnv("HOSTNAME")
	headers := metadata.New(map[string]string{"remote-host": host, "timestamp": timestamp})
	grpc.SendHeader(ctx, headers)

	return &pb.Empty{}, nil
}
