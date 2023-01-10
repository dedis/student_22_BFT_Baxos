// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: proto/BFTBaxos/BFTBaxos.proto

package BFTBaxos

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

// BFTBaxosClient is the client API for BFTBaxos service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BFTBaxosClient interface {
	Promise(ctx context.Context, in *PrepareMsg, opts ...grpc.CallOption) (*PromiseMsg, error)
	PreAccept(ctx context.Context, in *PreProposeMsg, opts ...grpc.CallOption) (*AcceptMsg, error)
	Accept(ctx context.Context, in *ProposeMsg, opts ...grpc.CallOption) (*AcceptMsg, error)
	Commit(ctx context.Context, in *CommitMsg, opts ...grpc.CallOption) (*Empty, error)
}

type bFTBaxosClient struct {
	cc grpc.ClientConnInterface
}

func NewBFTBaxosClient(cc grpc.ClientConnInterface) BFTBaxosClient {
	return &bFTBaxosClient{cc}
}

func (c *bFTBaxosClient) Promise(ctx context.Context, in *PrepareMsg, opts ...grpc.CallOption) (*PromiseMsg, error) {
	out := new(PromiseMsg)
	err := c.cc.Invoke(ctx, "/BFTBaxos/Promise", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bFTBaxosClient) PreAccept(ctx context.Context, in *PreProposeMsg, opts ...grpc.CallOption) (*AcceptMsg, error) {
	out := new(AcceptMsg)
	err := c.cc.Invoke(ctx, "/BFTBaxos/PreAccept", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bFTBaxosClient) Accept(ctx context.Context, in *ProposeMsg, opts ...grpc.CallOption) (*AcceptMsg, error) {
	out := new(AcceptMsg)
	err := c.cc.Invoke(ctx, "/BFTBaxos/Accept", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bFTBaxosClient) Commit(ctx context.Context, in *CommitMsg, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/BFTBaxos/Commit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BFTBaxosServer is the server API for BFTBaxos service.
// All implementations must embed UnimplementedBFTBaxosServer
// for forward compatibility
type BFTBaxosServer interface {
	Promise(context.Context, *PrepareMsg) (*PromiseMsg, error)
	PreAccept(context.Context, *PreProposeMsg) (*AcceptMsg, error)
	Accept(context.Context, *ProposeMsg) (*AcceptMsg, error)
	Commit(context.Context, *CommitMsg) (*Empty, error)
	mustEmbedUnimplementedBFTBaxosServer()
}

// UnimplementedBFTBaxosServer must be embedded to have forward compatible implementations.
type UnimplementedBFTBaxosServer struct {
}

func (UnimplementedBFTBaxosServer) Promise(context.Context, *PrepareMsg) (*PromiseMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Promise not implemented")
}
func (UnimplementedBFTBaxosServer) PreAccept(context.Context, *PreProposeMsg) (*AcceptMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreAccept not implemented")
}
func (UnimplementedBFTBaxosServer) Accept(context.Context, *ProposeMsg) (*AcceptMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Accept not implemented")
}
func (UnimplementedBFTBaxosServer) Commit(context.Context, *CommitMsg) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commit not implemented")
}
func (UnimplementedBFTBaxosServer) mustEmbedUnimplementedBFTBaxosServer() {}

// UnsafeBFTBaxosServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BFTBaxosServer will
// result in compilation errors.
type UnsafeBFTBaxosServer interface {
	mustEmbedUnimplementedBFTBaxosServer()
}

func RegisterBFTBaxosServer(s grpc.ServiceRegistrar, srv BFTBaxosServer) {
	s.RegisterService(&BFTBaxos_ServiceDesc, srv)
}

func _BFTBaxos_Promise_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrepareMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BFTBaxosServer).Promise(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/BFTBaxos/Promise",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BFTBaxosServer).Promise(ctx, req.(*PrepareMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _BFTBaxos_PreAccept_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PreProposeMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BFTBaxosServer).PreAccept(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/BFTBaxos/PreAccept",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BFTBaxosServer).PreAccept(ctx, req.(*PreProposeMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _BFTBaxos_Accept_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProposeMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BFTBaxosServer).Accept(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/BFTBaxos/Accept",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BFTBaxosServer).Accept(ctx, req.(*ProposeMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _BFTBaxos_Commit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BFTBaxosServer).Commit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/BFTBaxos/Commit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BFTBaxosServer).Commit(ctx, req.(*CommitMsg))
	}
	return interceptor(ctx, in, info, handler)
}

// BFTBaxos_ServiceDesc is the grpc.ServiceDesc for BFTBaxos service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BFTBaxos_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "BFTBaxos",
	HandlerType: (*BFTBaxosServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Promise",
			Handler:    _BFTBaxos_Promise_Handler,
		},
		{
			MethodName: "PreAccept",
			Handler:    _BFTBaxos_PreAccept_Handler,
		},
		{
			MethodName: "Accept",
			Handler:    _BFTBaxos_Accept_Handler,
		},
		{
			MethodName: "Commit",
			Handler:    _BFTBaxos_Commit_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/BFTBaxos/BFTBaxos.proto",
}
