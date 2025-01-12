protos: FORCE
	mkdir -p grpcsh/pb; \
	protoc -I protos/ \
	--go_out=grpcsh/pb \
	--go-grpc_out=grpcsh/pb \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	protos/*.proto

build_router:
	mkdir -p bin; \
	cd grpcsh; \
	go build -o ../bin/router router.go; \
	cd ..

build_agent:
	mkdir -p bin; \
	cd grpcsh; \
	go build -o ../bin/agent agent.go; \
	cd ..

build_grpcsh:
	mkdir -p bin; \
	cd grpcsh; \
	go build -o ../bin/grpcsh grpcsh.go; \
	cd ..

build: build_router build_agent build_grpcsh

start_router:
	./bin/router -r localhost:50051

start_agent:
	./bin/agent -r localhost:50051 -i abcde -s $(CURDIR)/agent.sock

run_grpcsh:
	./bin/grpcsh -i abcde -s $(CURDIR)/agent.sock -c "ls -l"

run_grpcsh_with_stdin:
	echo "this is from stdin" | ./bin/grpcsh -i abcde -s $(CURDIR)/agent.sock -c "cat"

FORCE: