package service

import (
	cr "github.com/focusj/grpc-go-chatroom/chatroom"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"io"
	"sync"
)

type ChatServer struct {
	sync.RWMutex

	cr.UnimplementedChatRoomServer

	groups map[int64]*cr.Group

	channels map[int64]cr.ChatRoom_ChatServer
}

func (c *ChatServer) register(uid int64, srv cr.ChatRoom_ChatServer) {
	c.Lock()
	defer c.Unlock()
	grpclog.Info("user: ")
	if _, exists := c.channels[uid]; !exists {
		c.channels[uid] = srv
	}
}

func (c *ChatServer) unRegister(uid int64) {
	c.Lock()
	defer c.Unlock()
	delete(c.channels, uid)
}

func (c *ChatServer) deliver(groupId int64, msg *cr.Message) {
	group := c.groups[groupId]
	for _, uid := range group.Members {
		if ch, exists := c.channels[uid]; exists {
			err := ch.Send(msg)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					if e.Code() == codes.Unavailable {
						c.unRegister(msg.Sender)
					}
				}
				grpclog.Error(err)
			}
		}
	}
}

func New() *ChatServer {
	groups := make(map[int64]*cr.Group)
	groups[1] = &cr.Group{
		Id:      1,
		Name:    "group-01",
		Members: []int64{1, 2, 3},
	}
	channels := make(map[int64]cr.ChatRoom_ChatServer)
	return &ChatServer{groups: groups, channels: channels}
}

func (c *ChatServer) Chat(srv cr.ChatRoom_ChatServer) error {
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
			grpclog.Error(err)
			break
		}
		sender = msg.Sender

		c.register(msg.Sender, srv) // should be moved to login, just call once when a user login

		c.deliver(msg.GroupId, msg)
	}

	return status.Errorf(codes.Unimplemented, "method Chat not implemented")
}
