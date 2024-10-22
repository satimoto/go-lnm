// Code generated by protoc-gen-go. DO NOT EDIT.
// source: lsprpc/channel.proto

package lsprpc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type OpenChannelRequest struct {
	Pubkey               string   `protobuf:"bytes,1,opt,name=pubkey,proto3" json:"pubkey,omitempty"`
	Amount               int64    `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
	AmountMsat           int64    `protobuf:"varint,3,opt,name=amount_msat,json=amountMsat,proto3" json:"amount_msat,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OpenChannelRequest) Reset()         { *m = OpenChannelRequest{} }
func (m *OpenChannelRequest) String() string { return proto.CompactTextString(m) }
func (*OpenChannelRequest) ProtoMessage()    {}
func (*OpenChannelRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ff1f930398b2a84, []int{0}
}

func (m *OpenChannelRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OpenChannelRequest.Unmarshal(m, b)
}
func (m *OpenChannelRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OpenChannelRequest.Marshal(b, m, deterministic)
}
func (m *OpenChannelRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OpenChannelRequest.Merge(m, src)
}
func (m *OpenChannelRequest) XXX_Size() int {
	return xxx_messageInfo_OpenChannelRequest.Size(m)
}
func (m *OpenChannelRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_OpenChannelRequest.DiscardUnknown(m)
}

var xxx_messageInfo_OpenChannelRequest proto.InternalMessageInfo

func (m *OpenChannelRequest) GetPubkey() string {
	if m != nil {
		return m.Pubkey
	}
	return ""
}

func (m *OpenChannelRequest) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *OpenChannelRequest) GetAmountMsat() int64 {
	if m != nil {
		return m.AmountMsat
	}
	return 0
}

type OpenChannelResponse struct {
	PendingChanId             []byte   `protobuf:"bytes,1,opt,name=pending_chan_id,json=pendingChanId,proto3" json:"pending_chan_id,omitempty"`
	Scid                      uint64   `protobuf:"varint,2,opt,name=scid,proto3" json:"scid,omitempty"`
	FeeBaseMsat               int64    `protobuf:"varint,3,opt,name=fee_base_msat,json=feeBaseMsat,proto3" json:"fee_base_msat,omitempty"`
	FeeProportionalMillionths uint32   `protobuf:"varint,4,opt,name=fee_proportional_millionths,json=feeProportionalMillionths,proto3" json:"fee_proportional_millionths,omitempty"`
	CltvExpiryDelta           uint32   `protobuf:"varint,5,opt,name=cltv_expiry_delta,json=cltvExpiryDelta,proto3" json:"cltv_expiry_delta,omitempty"`
	XXX_NoUnkeyedLiteral      struct{} `json:"-"`
	XXX_unrecognized          []byte   `json:"-"`
	XXX_sizecache             int32    `json:"-"`
}

func (m *OpenChannelResponse) Reset()         { *m = OpenChannelResponse{} }
func (m *OpenChannelResponse) String() string { return proto.CompactTextString(m) }
func (*OpenChannelResponse) ProtoMessage()    {}
func (*OpenChannelResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ff1f930398b2a84, []int{1}
}

func (m *OpenChannelResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OpenChannelResponse.Unmarshal(m, b)
}
func (m *OpenChannelResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OpenChannelResponse.Marshal(b, m, deterministic)
}
func (m *OpenChannelResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OpenChannelResponse.Merge(m, src)
}
func (m *OpenChannelResponse) XXX_Size() int {
	return xxx_messageInfo_OpenChannelResponse.Size(m)
}
func (m *OpenChannelResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_OpenChannelResponse.DiscardUnknown(m)
}

var xxx_messageInfo_OpenChannelResponse proto.InternalMessageInfo

func (m *OpenChannelResponse) GetPendingChanId() []byte {
	if m != nil {
		return m.PendingChanId
	}
	return nil
}

func (m *OpenChannelResponse) GetScid() uint64 {
	if m != nil {
		return m.Scid
	}
	return 0
}

func (m *OpenChannelResponse) GetFeeBaseMsat() int64 {
	if m != nil {
		return m.FeeBaseMsat
	}
	return 0
}

func (m *OpenChannelResponse) GetFeeProportionalMillionths() uint32 {
	if m != nil {
		return m.FeeProportionalMillionths
	}
	return 0
}

func (m *OpenChannelResponse) GetCltvExpiryDelta() uint32 {
	if m != nil {
		return m.CltvExpiryDelta
	}
	return 0
}

type ListChannelsRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListChannelsRequest) Reset()         { *m = ListChannelsRequest{} }
func (m *ListChannelsRequest) String() string { return proto.CompactTextString(m) }
func (*ListChannelsRequest) ProtoMessage()    {}
func (*ListChannelsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ff1f930398b2a84, []int{2}
}

func (m *ListChannelsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListChannelsRequest.Unmarshal(m, b)
}
func (m *ListChannelsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListChannelsRequest.Marshal(b, m, deterministic)
}
func (m *ListChannelsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListChannelsRequest.Merge(m, src)
}
func (m *ListChannelsRequest) XXX_Size() int {
	return xxx_messageInfo_ListChannelsRequest.Size(m)
}
func (m *ListChannelsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListChannelsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListChannelsRequest proto.InternalMessageInfo

type ListChannelsResponse struct {
	ChannelIds           []string `protobuf:"bytes,1,rep,name=channel_ids,json=channelIds,proto3" json:"channel_ids,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListChannelsResponse) Reset()         { *m = ListChannelsResponse{} }
func (m *ListChannelsResponse) String() string { return proto.CompactTextString(m) }
func (*ListChannelsResponse) ProtoMessage()    {}
func (*ListChannelsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ff1f930398b2a84, []int{3}
}

func (m *ListChannelsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListChannelsResponse.Unmarshal(m, b)
}
func (m *ListChannelsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListChannelsResponse.Marshal(b, m, deterministic)
}
func (m *ListChannelsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListChannelsResponse.Merge(m, src)
}
func (m *ListChannelsResponse) XXX_Size() int {
	return xxx_messageInfo_ListChannelsResponse.Size(m)
}
func (m *ListChannelsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListChannelsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListChannelsResponse proto.InternalMessageInfo

func (m *ListChannelsResponse) GetChannelIds() []string {
	if m != nil {
		return m.ChannelIds
	}
	return nil
}

func init() {
	proto.RegisterType((*OpenChannelRequest)(nil), "channel.OpenChannelRequest")
	proto.RegisterType((*OpenChannelResponse)(nil), "channel.OpenChannelResponse")
	proto.RegisterType((*ListChannelsRequest)(nil), "channel.ListChannelsRequest")
	proto.RegisterType((*ListChannelsResponse)(nil), "channel.ListChannelsResponse")
}

func init() { proto.RegisterFile("lsprpc/channel.proto", fileDescriptor_2ff1f930398b2a84) }

var fileDescriptor_2ff1f930398b2a84 = []byte{
	// 390 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x52, 0xcd, 0x8e, 0xd3, 0x30,
	0x10, 0x56, 0x68, 0x59, 0xb4, 0x93, 0x2d, 0x2b, 0xbc, 0x0b, 0x0a, 0xbb, 0xa0, 0x0d, 0x41, 0x42,
	0x11, 0x12, 0x8d, 0x04, 0x07, 0x6e, 0x1c, 0x16, 0x90, 0xa8, 0xa0, 0x02, 0x85, 0x1b, 0x17, 0xcb,
	0x49, 0xa6, 0xad, 0x85, 0x63, 0x9b, 0x8c, 0x53, 0xd1, 0x27, 0xe2, 0xb9, 0x78, 0x13, 0x94, 0xc4,
	0x40, 0x2a, 0xca, 0xcd, 0xdf, 0xcf, 0xf8, 0xf3, 0x78, 0x06, 0xce, 0x15, 0xd9, 0xc6, 0x96, 0x59,
	0xb9, 0x11, 0x5a, 0xa3, 0x9a, 0xdb, 0xc6, 0x38, 0xc3, 0x6e, 0x79, 0x98, 0x20, 0xb0, 0x8f, 0x16,
	0xf5, 0xeb, 0x01, 0xe6, 0xf8, 0xad, 0x45, 0x72, 0xec, 0x1e, 0x1c, 0xd9, 0xb6, 0xf8, 0x8a, 0xbb,
	0x28, 0x88, 0x83, 0xf4, 0x38, 0xf7, 0xa8, 0xe3, 0x45, 0x6d, 0x5a, 0xed, 0xa2, 0x1b, 0x71, 0x90,
	0x4e, 0x72, 0x8f, 0xd8, 0x15, 0x84, 0xc3, 0x89, 0xd7, 0x24, 0x5c, 0x34, 0xe9, 0x45, 0x18, 0xa8,
	0x25, 0x09, 0x97, 0xfc, 0x0c, 0xe0, 0x6c, 0x2f, 0x87, 0xac, 0xd1, 0x84, 0xec, 0x09, 0x9c, 0x5a,
	0xd4, 0x95, 0xd4, 0x6b, 0xde, 0xbd, 0x88, 0xcb, 0xaa, 0x4f, 0x3c, 0xc9, 0x67, 0x9e, 0xee, 0x0a,
	0x16, 0x15, 0x63, 0x30, 0xa5, 0x52, 0x56, 0x7d, 0xec, 0x34, 0xef, 0xcf, 0x2c, 0x81, 0xd9, 0x0a,
	0x91, 0x17, 0x82, 0x70, 0x1c, 0x1b, 0xae, 0x10, 0xaf, 0x05, 0x61, 0x97, 0xcb, 0x5e, 0xc1, 0x65,
	0xe7, 0xb1, 0x8d, 0xb1, 0xa6, 0x71, 0xd2, 0x68, 0xa1, 0x78, 0x2d, 0x95, 0x92, 0x46, 0xbb, 0x0d,
	0x45, 0xd3, 0x38, 0x48, 0x67, 0xf9, 0xfd, 0x15, 0xe2, 0xa7, 0x91, 0x63, 0xf9, 0xc7, 0xc0, 0x9e,
	0xc2, 0x9d, 0x52, 0xb9, 0x2d, 0xc7, 0xef, 0x56, 0x36, 0x3b, 0x5e, 0xa1, 0x72, 0x22, 0xba, 0xd9,
	0x57, 0x9d, 0x76, 0xc2, 0xdb, 0x9e, 0x7f, 0xd3, 0xd1, 0xc9, 0x5d, 0x38, 0xfb, 0x20, 0xc9, 0xf9,
	0x16, 0xc9, 0xff, 0x65, 0xf2, 0x12, 0xce, 0xf7, 0x69, 0xdf, 0xfa, 0x15, 0x84, 0x7e, 0x08, 0x5c,
	0x56, 0x14, 0x05, 0xf1, 0x24, 0x3d, 0xce, 0xc1, 0x53, 0x8b, 0x8a, 0x9e, 0xff, 0x08, 0xe0, 0xb6,
	0xaf, 0xfa, 0x8c, 0xcd, 0x56, 0x96, 0xc8, 0xde, 0x41, 0x38, 0xfa, 0x45, 0x76, 0x39, 0xff, 0x3d,
	0xd5, 0x7f, 0x67, 0x78, 0xf1, 0xe0, 0xb0, 0xe8, 0xd3, 0xdf, 0xc3, 0xc9, 0xf8, 0x55, 0xec, 0xaf,
	0xfb, 0x40, 0x0f, 0x17, 0x0f, 0xff, 0xa3, 0x0e, 0x97, 0x5d, 0x3f, 0xfe, 0xf2, 0x68, 0x2d, 0xdd,
	0xa6, 0x2d, 0xe6, 0xa5, 0xa9, 0x33, 0x12, 0x4e, 0xd6, 0xc6, 0x99, 0x6c, 0x6d, 0x9e, 0x29, 0xb2,
	0xd9, 0xb0, 0x80, 0xc5, 0x51, 0xbf, 0x79, 0x2f, 0x7e, 0x05, 0x00, 0x00, 0xff, 0xff, 0x6b, 0x03,
	0xfd, 0xfd, 0x91, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ChannelServiceClient is the client API for ChannelService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ChannelServiceClient interface {
	OpenChannel(ctx context.Context, in *OpenChannelRequest, opts ...grpc.CallOption) (*OpenChannelResponse, error)
	ListChannels(ctx context.Context, in *ListChannelsRequest, opts ...grpc.CallOption) (*ListChannelsResponse, error)
}

type channelServiceClient struct {
	cc *grpc.ClientConn
}

func NewChannelServiceClient(cc *grpc.ClientConn) ChannelServiceClient {
	return &channelServiceClient{cc}
}

func (c *channelServiceClient) OpenChannel(ctx context.Context, in *OpenChannelRequest, opts ...grpc.CallOption) (*OpenChannelResponse, error) {
	out := new(OpenChannelResponse)
	err := c.cc.Invoke(ctx, "/channel.ChannelService/OpenChannel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelServiceClient) ListChannels(ctx context.Context, in *ListChannelsRequest, opts ...grpc.CallOption) (*ListChannelsResponse, error) {
	out := new(ListChannelsResponse)
	err := c.cc.Invoke(ctx, "/channel.ChannelService/ListChannels", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelServiceServer is the server API for ChannelService service.
type ChannelServiceServer interface {
	OpenChannel(context.Context, *OpenChannelRequest) (*OpenChannelResponse, error)
	ListChannels(context.Context, *ListChannelsRequest) (*ListChannelsResponse, error)
}

// UnimplementedChannelServiceServer can be embedded to have forward compatible implementations.
type UnimplementedChannelServiceServer struct {
}

func (*UnimplementedChannelServiceServer) OpenChannel(ctx context.Context, req *OpenChannelRequest) (*OpenChannelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OpenChannel not implemented")
}
func (*UnimplementedChannelServiceServer) ListChannels(ctx context.Context, req *ListChannelsRequest) (*ListChannelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListChannels not implemented")
}

func RegisterChannelServiceServer(s *grpc.Server, srv ChannelServiceServer) {
	s.RegisterService(&_ChannelService_serviceDesc, srv)
}

func _ChannelService_OpenChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OpenChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServiceServer).OpenChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/channel.ChannelService/OpenChannel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServiceServer).OpenChannel(ctx, req.(*OpenChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChannelService_ListChannels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListChannelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServiceServer).ListChannels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/channel.ChannelService/ListChannels",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServiceServer).ListChannels(ctx, req.(*ListChannelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ChannelService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "channel.ChannelService",
	HandlerType: (*ChannelServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "OpenChannel",
			Handler:    _ChannelService_OpenChannel_Handler,
		},
		{
			MethodName: "ListChannels",
			Handler:    _ChannelService_ListChannels_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "lsprpc/channel.proto",
}
