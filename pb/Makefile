PROTOC:=protoc
GO_OUT:=./
GO_GRPC_OUT:=./
PROTO_SRC:=*.proto

all: $(PROTO_SRC)
	$(PROTOC) --go_out=$(GO_OUT) --go-grpc_out=$(GO_GRPC_OUT) $^

clean:
	find -name "*.go" -delete
