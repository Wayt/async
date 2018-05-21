protoc:
	protoc -I pb/ pb/server.proto --go_out=plugins=grpc:pb/server
	protoc -I pb/ pb/worker.proto --go_out=plugins=grpc:pb/worker

.PHONY: protoc
