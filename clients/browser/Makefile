compile-pb:
	protoc -I=../../pkg/protocol/v1/protobuf/  \
	../../pkg/protocol/v1/protobuf/comms.proto \
	--grpc-web_out=import_style=commonjs,mode=grpcwebtext:./extension \
	--js_out=import_style=commonjs:./extension
	
	protoc -I=../../pkg/protocol/v1/protobuf/  \
	../../pkg/protocol/v1/protobuf/google/api/annotations.proto \
	--js_out=import_style=commonjs:./extension
	
	protoc -I=../../pkg/protocol/v1/protobuf/  \
	../../pkg/protocol/v1/protobuf/google/api/http.proto \
	--js_out=import_style=commonjs:./extension

pack-extension:
	web-ext build --source-src ./extension --overwrite-dest