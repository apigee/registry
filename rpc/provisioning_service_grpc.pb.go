// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.9
// source: google/cloud/apigeeregistry/v1/provisioning_service.proto

package rpc

import (
	context "context"

	longrunning "cloud.google.com/go/longrunning/autogen/longrunningpb"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Provisioning_CreateInstance_FullMethodName = "/google.cloud.apigeeregistry.v1.Provisioning/CreateInstance"
	Provisioning_DeleteInstance_FullMethodName = "/google.cloud.apigeeregistry.v1.Provisioning/DeleteInstance"
	Provisioning_GetInstance_FullMethodName    = "/google.cloud.apigeeregistry.v1.Provisioning/GetInstance"
)

// ProvisioningClient is the client API for Provisioning service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProvisioningClient interface {
	// Provisions instance resources for the Registry.
	CreateInstance(ctx context.Context, in *CreateInstanceRequest, opts ...grpc.CallOption) (*longrunning.Operation, error)
	// Deletes the Registry instance.
	DeleteInstance(ctx context.Context, in *DeleteInstanceRequest, opts ...grpc.CallOption) (*longrunning.Operation, error)
	// Gets details of a single Instance.
	GetInstance(ctx context.Context, in *GetInstanceRequest, opts ...grpc.CallOption) (*Instance, error)
}

type provisioningClient struct {
	cc grpc.ClientConnInterface
}

func NewProvisioningClient(cc grpc.ClientConnInterface) ProvisioningClient {
	return &provisioningClient{cc}
}

func (c *provisioningClient) CreateInstance(ctx context.Context, in *CreateInstanceRequest, opts ...grpc.CallOption) (*longrunning.Operation, error) {
	out := new(longrunning.Operation)
	err := c.cc.Invoke(ctx, Provisioning_CreateInstance_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *provisioningClient) DeleteInstance(ctx context.Context, in *DeleteInstanceRequest, opts ...grpc.CallOption) (*longrunning.Operation, error) {
	out := new(longrunning.Operation)
	err := c.cc.Invoke(ctx, Provisioning_DeleteInstance_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *provisioningClient) GetInstance(ctx context.Context, in *GetInstanceRequest, opts ...grpc.CallOption) (*Instance, error) {
	out := new(Instance)
	err := c.cc.Invoke(ctx, Provisioning_GetInstance_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProvisioningServer is the server API for Provisioning service.
// All implementations must embed UnimplementedProvisioningServer
// for forward compatibility
type ProvisioningServer interface {
	// Provisions instance resources for the Registry.
	CreateInstance(context.Context, *CreateInstanceRequest) (*longrunning.Operation, error)
	// Deletes the Registry instance.
	DeleteInstance(context.Context, *DeleteInstanceRequest) (*longrunning.Operation, error)
	// Gets details of a single Instance.
	GetInstance(context.Context, *GetInstanceRequest) (*Instance, error)
	mustEmbedUnimplementedProvisioningServer()
}

// UnimplementedProvisioningServer must be embedded to have forward compatible implementations.
type UnimplementedProvisioningServer struct {
}

func (UnimplementedProvisioningServer) CreateInstance(context.Context, *CreateInstanceRequest) (*longrunning.Operation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateInstance not implemented")
}
func (UnimplementedProvisioningServer) DeleteInstance(context.Context, *DeleteInstanceRequest) (*longrunning.Operation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteInstance not implemented")
}
func (UnimplementedProvisioningServer) GetInstance(context.Context, *GetInstanceRequest) (*Instance, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstance not implemented")
}
func (UnimplementedProvisioningServer) mustEmbedUnimplementedProvisioningServer() {}

// UnsafeProvisioningServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProvisioningServer will
// result in compilation errors.
type UnsafeProvisioningServer interface {
	mustEmbedUnimplementedProvisioningServer()
}

func RegisterProvisioningServer(s grpc.ServiceRegistrar, srv ProvisioningServer) {
	s.RegisterService(&Provisioning_ServiceDesc, srv)
}

func _Provisioning_CreateInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProvisioningServer).CreateInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Provisioning_CreateInstance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProvisioningServer).CreateInstance(ctx, req.(*CreateInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provisioning_DeleteInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProvisioningServer).DeleteInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Provisioning_DeleteInstance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProvisioningServer).DeleteInstance(ctx, req.(*DeleteInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provisioning_GetInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProvisioningServer).GetInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Provisioning_GetInstance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProvisioningServer).GetInstance(ctx, req.(*GetInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Provisioning_ServiceDesc is the grpc.ServiceDesc for Provisioning service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Provisioning_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "google.cloud.apigeeregistry.v1.Provisioning",
	HandlerType: (*ProvisioningServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateInstance",
			Handler:    _Provisioning_CreateInstance_Handler,
		},
		{
			MethodName: "DeleteInstance",
			Handler:    _Provisioning_DeleteInstance_Handler,
		},
		{
			MethodName: "GetInstance",
			Handler:    _Provisioning_GetInstance_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/cloud/apigeeregistry/v1/provisioning_service.proto",
}
