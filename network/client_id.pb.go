// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: client_id.proto

package network

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ClientID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address string `protobuf:"bytes,1,opt,name=Address,proto3" json:"Address,omitempty"` // 地址
	Id      string `protobuf:"bytes,2,opt,name=Id,proto3" json:"Id,omitempty"`           // 唯一标记

	vaild int32 
	h *Handler
}

func (x *ClientID) Reset() {
	*x = ClientID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_client_id_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientID) ProtoMessage() {}

func (x *ClientID) ProtoReflect() protoreflect.Message {
	mi := &file_client_id_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientID.ProtoReflect.Descriptor instead.
func (*ClientID) Descriptor() ([]byte, []int) {
	return file_client_id_proto_rawDescGZIP(), []int{0}
}

func (x *ClientID) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ClientID) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_client_id_proto protoreflect.FileDescriptor

var file_client_id_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x22, 0x34, 0x0a, 0x08, 0x43, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x12, 0x0e, 0x0a, 0x02, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x64,
	0x42, 0x0b, 0x5a, 0x09, 0x2e, 0x3b, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_client_id_proto_rawDescOnce sync.Once
	file_client_id_proto_rawDescData = file_client_id_proto_rawDesc
)

func file_client_id_proto_rawDescGZIP() []byte {
	file_client_id_proto_rawDescOnce.Do(func() {
		file_client_id_proto_rawDescData = protoimpl.X.CompressGZIP(file_client_id_proto_rawDescData)
	})
	return file_client_id_proto_rawDescData
}

var file_client_id_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_client_id_proto_goTypes = []interface{}{
	(*ClientID)(nil), // 0: network.ClientID
}
var file_client_id_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_client_id_proto_init() }
func file_client_id_proto_init() {
	if File_client_id_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_client_id_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientID); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_client_id_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_client_id_proto_goTypes,
		DependencyIndexes: file_client_id_proto_depIdxs,
		MessageInfos:      file_client_id_proto_msgTypes,
	}.Build()
	File_client_id_proto = out.File
	file_client_id_proto_rawDesc = nil
	file_client_id_proto_goTypes = nil
	file_client_id_proto_depIdxs = nil
}
