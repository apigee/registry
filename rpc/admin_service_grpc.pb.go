// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package rpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AdminClient is the client API for Admin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AdminClient interface {
	// GetStatus returns the status of the service.
	// GetStatus is for verifying open source deployments only
	// and is not included in hosted versions of the API.
	// (-- api-linter: core::0131::request-message-name=disabled
	//     aip.dev/not-precedent: Not in the official API. --)
	// (-- api-linter: core::0131::method-signature=disabled
	//     aip.dev/not-precedent: Not in the official API. --)
	// (-- api-linter: core::0131::http-uri-name=disabled
	//     aip.dev/not-precedent: Not in the official API. --)
	GetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*Status, error)
	// ListProjects returns matching projects.
	// (-- api-linter: standard-methods=disabled --)
	// (-- api-linter: core::0132::method-signature=disabled
	//     aip.dev/not-precedent: projects are top-level resources. --)
	ListProjects(ctx context.Context, in *ListProjectsRequest, opts ...grpc.CallOption) (*ListProjectsResponse, error)
	// GetProject returns a specified project.
	GetProject(ctx context.Context, in *GetProjectRequest, opts ...grpc.CallOption) (*Project, error)
	// CreateProject creates a specified project.
	// (-- api-linter: standard-methods=disabled --)
	// (-- api-linter: core::0133::http-uri-parent=disabled
	//     aip.dev/not-precedent: Project has an implicit parent. --)
	// (-- api-linter: core::0133::method-signature=disabled
	//     aip.dev/not-precedent: Project has an implicit parent. --)
	CreateProject(ctx context.Context, in *CreateProjectRequest, opts ...grpc.CallOption) (*Project, error)
	// UpdateProject can be used to modify a specified project.
	UpdateProject(ctx context.Context, in *UpdateProjectRequest, opts ...grpc.CallOption) (*Project, error)
	// DeleteProject removes a specified project and all of the resources that it
	// owns.
	DeleteProject(ctx context.Context, in *DeleteProjectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type adminClient struct {
	cc grpc.ClientConnInterface
}

func NewAdminClient(cc grpc.ClientConnInterface) AdminClient {
	return &adminClient{cc}
}

func (c *adminClient) GetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/google.cloud.apigeeregistry.v1.Admin/GetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) ListProjects(ctx context.Context, in *ListProjectsRequest, opts ...grpc.CallOption) (*ListProjectsResponse, error) {
	out := new(ListProjectsResponse)
	err := c.cc.Invoke(ctx, "/google.cloud.apigeeregistry.v1.Admin/ListProjects", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) GetProject(ctx context.Context, in *GetProjectRequest, opts ...grpc.CallOption) (*Project, error) {
	out := new(Project)
	err := c.cc.Invoke(ctx, "/google.cloud.apigeeregistry.v1.Admin/GetProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) CreateProject(ctx context.Context, in *CreateProjectRequest, opts ...grpc.CallOption) (*Project, error) {
	out := new(Project)
	err := c.cc.Invoke(ctx, "/google.cloud.apigeeregistry.v1.Admin/CreateProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) UpdateProject(ctx context.Context, in *UpdateProjectRequest, opts ...grpc.CallOption) (*Project, error) {
	out := new(Project)
	err := c.cc.Invoke(ctx, "/google.cloud.apigeeregistry.v1.Admin/UpdateProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) DeleteProject(ctx context.Context, in *DeleteProjectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/google.cloud.apigeeregistry.v1.Admin/DeleteProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AdminServer is the server API for Admin service.
// All implementations must embed UnimplementedAdminServer
// for forward compatibility
type AdminServer interface {
	// GetStatus returns the status of the service.
	// GetStatus is for verifying open source deployments only
	// and is not included in hosted versions of the API.
	// (-- api-linter: core::0131::request-message-name=disabled
	//     aip.dev/not-precedent: Not in the official API. --)
	// (-- api-linter: core::0131::method-signature=disabled
	//     aip.dev/not-precedent: Not in the official API. --)
	// (-- api-linter: core::0131::http-uri-name=disabled
	//     aip.dev/not-precedent: Not in the official API. --)
	GetStatus(context.Context, *emptypb.Empty) (*Status, error)
	// ListProjects returns matching projects.
	// (-- api-linter: standard-methods=disabled --)
	// (-- api-linter: core::0132::method-signature=disabled
	//     aip.dev/not-precedent: projects are top-level resources. --)
	ListProjects(context.Context, *ListProjectsRequest) (*ListProjectsResponse, error)
	// GetProject returns a specified project.
	GetProject(context.Context, *GetProjectRequest) (*Project, error)
	// CreateProject creates a specified project.
	// (-- api-linter: standard-methods=disabled --)
	// (-- api-linter: core::0133::http-uri-parent=disabled
	//     aip.dev/not-precedent: Project has an implicit parent. --)
	// (-- api-linter: core::0133::method-signature=disabled
	//     aip.dev/not-precedent: Project has an implicit parent. --)
	CreateProject(context.Context, *CreateProjectRequest) (*Project, error)
	// UpdateProject can be used to modify a specified project.
	UpdateProject(context.Context, *UpdateProjectRequest) (*Project, error)
	// DeleteProject removes a specified project and all of the resources that it
	// owns.
	DeleteProject(context.Context, *DeleteProjectRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedAdminServer()
}

// UnimplementedAdminServer must be embedded to have forward compatible implementations.
type UnimplementedAdminServer struct {
}

func (UnimplementedAdminServer) GetStatus(context.Context, *emptypb.Empty) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (UnimplementedAdminServer) ListProjects(context.Context, *ListProjectsRequest) (*ListProjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListProjects not implemented")
}
func (UnimplementedAdminServer) GetProject(context.Context, *GetProjectRequest) (*Project, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProject not implemented")
}
func (UnimplementedAdminServer) CreateProject(context.Context, *CreateProjectRequest) (*Project, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateProject not implemented")
}
func (UnimplementedAdminServer) UpdateProject(context.Context, *UpdateProjectRequest) (*Project, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProject not implemented")
}
func (UnimplementedAdminServer) DeleteProject(context.Context, *DeleteProjectRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteProject not implemented")
}
func (UnimplementedAdminServer) mustEmbedUnimplementedAdminServer() {}

// UnsafeAdminServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AdminServer will
// result in compilation errors.
type UnsafeAdminServer interface {
	mustEmbedUnimplementedAdminServer()
}

func RegisterAdminServer(s grpc.ServiceRegistrar, srv AdminServer) {
	s.RegisterService(&Admin_ServiceDesc, srv)
}

func _Admin_GetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).GetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.apigeeregistry.v1.Admin/GetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).GetStatus(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_ListProjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListProjectsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).ListProjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.apigeeregistry.v1.Admin/ListProjects",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).ListProjects(ctx, req.(*ListProjectsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_GetProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).GetProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.apigeeregistry.v1.Admin/GetProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).GetProject(ctx, req.(*GetProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_CreateProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).CreateProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.apigeeregistry.v1.Admin/CreateProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).CreateProject(ctx, req.(*CreateProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_UpdateProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).UpdateProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.apigeeregistry.v1.Admin/UpdateProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).UpdateProject(ctx, req.(*UpdateProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_DeleteProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).DeleteProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.apigeeregistry.v1.Admin/DeleteProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).DeleteProject(ctx, req.(*DeleteProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Admin_ServiceDesc is the grpc.ServiceDesc for Admin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Admin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "google.cloud.apigeeregistry.v1.Admin",
	HandlerType: (*AdminServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetStatus",
			Handler:    _Admin_GetStatus_Handler,
		},
		{
			MethodName: "ListProjects",
			Handler:    _Admin_ListProjects_Handler,
		},
		{
			MethodName: "GetProject",
			Handler:    _Admin_GetProject_Handler,
		},
		{
			MethodName: "CreateProject",
			Handler:    _Admin_CreateProject_Handler,
		},
		{
			MethodName: "UpdateProject",
			Handler:    _Admin_UpdateProject_Handler,
		},
		{
			MethodName: "DeleteProject",
			Handler:    _Admin_DeleteProject_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/cloud/apigeeregistry/v1/admin_service.proto",
}
