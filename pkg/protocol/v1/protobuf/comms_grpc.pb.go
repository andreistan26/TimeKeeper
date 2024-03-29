// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: comms.proto

package protobuf

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TimeKeeperServiceClient is the client API for TimeKeeperService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TimeKeeperServiceClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	SendData(ctx context.Context, in *SendDataRequest, opts ...grpc.CallOption) (*SendDataResponse, error)
}

type timeKeeperServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTimeKeeperServiceClient(cc grpc.ClientConnInterface) TimeKeeperServiceClient {
	return &timeKeeperServiceClient{cc}
}

func (c *timeKeeperServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, "/timekeeper.TimeKeeperService/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *timeKeeperServiceClient) SendData(ctx context.Context, in *SendDataRequest, opts ...grpc.CallOption) (*SendDataResponse, error) {
	out := new(SendDataResponse)
	err := c.cc.Invoke(ctx, "/timekeeper.TimeKeeperService/SendData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TimeKeeperServiceServer is the server API for TimeKeeperService service.
// All implementations must embed UnimplementedTimeKeeperServiceServer
// for forward compatibility
type TimeKeeperServiceServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	SendData(context.Context, *SendDataRequest) (*SendDataResponse, error)
	mustEmbedUnimplementedTimeKeeperServiceServer()
}

// UnimplementedTimeKeeperServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTimeKeeperServiceServer struct {
}

func (UnimplementedTimeKeeperServiceServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedTimeKeeperServiceServer) SendData(context.Context, *SendDataRequest) (*SendDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendData not implemented")
}
func (UnimplementedTimeKeeperServiceServer) mustEmbedUnimplementedTimeKeeperServiceServer() {}

// UnsafeTimeKeeperServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TimeKeeperServiceServer will
// result in compilation errors.
type UnsafeTimeKeeperServiceServer interface {
	mustEmbedUnimplementedTimeKeeperServiceServer()
}

func RegisterTimeKeeperServiceServer(s grpc.ServiceRegistrar, srv TimeKeeperServiceServer) {
	s.RegisterService(&TimeKeeperService_ServiceDesc, srv)
}

func _TimeKeeperService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimeKeeperServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/timekeeper.TimeKeeperService/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimeKeeperServiceServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TimeKeeperService_SendData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimeKeeperServiceServer).SendData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/timekeeper.TimeKeeperService/SendData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimeKeeperServiceServer).SendData(ctx, req.(*SendDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TimeKeeperService_ServiceDesc is the grpc.ServiceDesc for TimeKeeperService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TimeKeeperService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "timekeeper.TimeKeeperService",
	HandlerType: (*TimeKeeperServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _TimeKeeperService_Register_Handler,
		},
		{
			MethodName: "SendData",
			Handler:    _TimeKeeperService_SendData_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "comms.proto",
}
