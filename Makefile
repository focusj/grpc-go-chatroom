gen_grpc_server:
	protoc -I ./chatroom --go_out=plugins=grpc:./chatroom chatroom/chatroom.proto

build:
	go build .

docker_latest:
	docker build -t focusjx/chatroom:latest .

