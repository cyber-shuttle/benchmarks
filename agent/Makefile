protos: FORCE
	mkdir -p grpcsh/pb; \
	protoc -I protos/ \
	--go_out=grpcsh/pb \
	--go-grpc_out=grpcsh/pb \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	protos/*.proto

build:
	mkdir -p bin; \
	cd grpcsh; \
	go build -o ../bin/agent agent.go; \
	go build -o ../bin/server router.go; \
	cd ..

FORCE: