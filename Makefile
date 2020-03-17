gen_grpc_server:
	protoc -I ./chatroom --go_out=plugins=grpc:./chatroom chatroom/chatroom.proto