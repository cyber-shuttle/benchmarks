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

build_amd64:
	mkdir -p bin; \
	cd grpcsh; \
	GOOS=linux GOARCH=amd64 go build -o ../bin/router_amd64 router.go; \
	GOOS=linux GOARCH=amd64 go build -o ../bin/agent_amd64 agent.go; \
	cd ..

build: build_router build_agent build_grpcsh

start_router:
	./bin/router -r localhost:50051

start_agents:
	./bin/agent -r localhost:50051 -i A -s $(CURDIR)/agent_A.sock & \
	./bin/agent -r localhost:50051 -i B -s $(CURDIR)/agent_B.sock

run_grpcsh:
	./bin/grpcsh -i B -s $(CURDIR)/agent_A.sock -c "ls -l"

run_grpcsh_with_stdin:
	echo "this is from stdin" | ./bin/grpcsh -i B -s $(CURDIR)/agent_A.sock -c "cat"

FORCE:
