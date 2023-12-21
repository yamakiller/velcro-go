// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: closed.proto

package protocols

import (
	network "github.com/yamakiller/velcro-go/network"
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

// 套接字已关闭
type Closed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientID *network.ClientID `protobuf:"bytes,1,opt,name=ClientID,proto3" json:"ClientID,omitempty"`
}

func (x *Closed) Reset() {
	*x = Closed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_closed_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Closed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Closed) ProtoMessage() {}

func (x *Closed) ProtoReflect() protoreflect.Message {
	mi := &file_closed_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Closed.ProtoReflect.Descriptor instead.
func (*Closed) Descriptor() ([]byte, []int) {
	return file_closed_proto_rawDescGZIP(), []int{0}
}

func (x *Closed) GetClientID() *network.ClientID {
	if x != nil {
		return x.ClientID
	}
	return nil
}

var File_closed_proto protoreflect.FileDescriptor

var file_closed_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x1a, 0x0f, 0x63, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x5f, 0x69, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x37, 0x0a, 0x06, 0x43, 0x6c,
	0x6f, 0x73, 0x65, 0x64, 0x12, 0x2d, 0x0a, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x52, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x49, 0x44, 0x42, 0x0d, 0x5a, 0x0b, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_closed_proto_rawDescOnce sync.Once
	file_closed_proto_rawDescData = file_closed_proto_rawDesc
)

func file_closed_proto_rawDescGZIP() []byte {
	file_closed_proto_rawDescOnce.Do(func() {
		file_closed_proto_rawDescData = protoimpl.X.CompressGZIP(file_closed_proto_rawDescData)
	})
	return file_closed_proto_rawDescData
}

var file_closed_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_closed_proto_goTypes = []interface{}{
	(*Closed)(nil),           // 0: protocols.Closed
	(*network.ClientID)(nil), // 1: network.ClientID
}
var file_closed_proto_depIdxs = []int32{
	1, // 0: protocols.Closed.ClientID:type_name -> network.ClientID
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_closed_proto_init() }
func file_closed_proto_init() {
	if File_closed_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_closed_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Closed); i {
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
			RawDescriptor: file_closed_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_closed_proto_goTypes,
		DependencyIndexes: file_closed_proto_depIdxs,
		MessageInfos:      file_closed_proto_msgTypes,
	}.Build()
	File_closed_proto = out.File
	file_closed_proto_rawDesc = nil
	file_closed_proto_goTypes = nil
	file_closed_proto_depIdxs = nil
}
