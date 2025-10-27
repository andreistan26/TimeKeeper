compile-pb:
	protoc -I=pkg/protobuf --go_out=pkg/gen/v1 --go_opt=paths=source_relative   \
	--grpc-gateway_out=pkg/gen/v1 --go-grpc_opt=paths=source_relative \
	     --go-grpc_out=pkg/gen/v1 --grpc-gateway_opt=paths=source_relative pkg/protobuf/comms.proto 
