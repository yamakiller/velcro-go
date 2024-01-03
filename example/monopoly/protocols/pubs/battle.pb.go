// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.2
// source: battle.proto

package pubs

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

type CreateBattleSpace struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MapURI   string `protobuf:"bytes,1,opt,name=mapURI,proto3" json:"mapURI,omitempty"`
	MaxCount uint32 `protobuf:"fixed32,2,opt,name=maxCount,proto3" json:"maxCount,omitempty"`
}

func (x *CreateBattleSpace) Reset() {
	*x = CreateBattleSpace{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateBattleSpace) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateBattleSpace) ProtoMessage() {}

func (x *CreateBattleSpace) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateBattleSpace.ProtoReflect.Descriptor instead.
func (*CreateBattleSpace) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{0}
}

func (x *CreateBattleSpace) GetMapURI() string {
	if x != nil {
		return x.MapURI
	}
	return ""
}

func (x *CreateBattleSpace) GetMaxCount() uint32 {
	if x != nil {
		return x.MaxCount
	}
	return 0
}

type CreateBattleSpaceResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId string `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
	MapURI  string `protobuf:"bytes,2,opt,name=mapURI,proto3" json:"mapURI,omitempty"`
}

func (x *CreateBattleSpaceResp) Reset() {
	*x = CreateBattleSpaceResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateBattleSpaceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateBattleSpaceResp) ProtoMessage() {}

func (x *CreateBattleSpaceResp) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateBattleSpaceResp.ProtoReflect.Descriptor instead.
func (*CreateBattleSpaceResp) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{1}
}

func (x *CreateBattleSpaceResp) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

func (x *CreateBattleSpaceResp) GetMapURI() string {
	if x != nil {
		return x.MapURI
	}
	return ""
}

type BattleSpacePlayerSimple struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Display string `protobuf:"bytes,1,opt,name=display,proto3" json:"display,omitempty"`
	Pos     int32  `protobuf:"varint,2,opt,name=pos,proto3" json:"pos,omitempty"`
}

func (x *BattleSpacePlayerSimple) Reset() {
	*x = BattleSpacePlayerSimple{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BattleSpacePlayerSimple) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BattleSpacePlayerSimple) ProtoMessage() {}

func (x *BattleSpacePlayerSimple) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BattleSpacePlayerSimple.ProtoReflect.Descriptor instead.
func (*BattleSpacePlayerSimple) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{2}
}

func (x *BattleSpacePlayerSimple) GetDisplay() string {
	if x != nil {
		return x.Display
	}
	return ""
}

func (x *BattleSpacePlayerSimple) GetPos() int32 {
	if x != nil {
		return x.Pos
	}
	return 0
}

type BattleSpaceDataSimple struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId       string                     `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
	MapURI        string                     `protobuf:"bytes,2,opt,name=mapURI,proto3" json:"mapURI,omitempty"`
	MasterUid     string                     `protobuf:"bytes,3,opt,name=masterUid,proto3" json:"masterUid,omitempty"`
	MasterIcon    string                     `protobuf:"bytes,4,opt,name=masterIcon,proto3" json:"masterIcon,omitempty"`
	MasterDisplay string                     `protobuf:"bytes,5,opt,name=masterDisplay,proto3" json:"masterDisplay,omitempty"`
	Players       []*BattleSpacePlayerSimple `protobuf:"bytes,6,rep,name=players,proto3" json:"players,omitempty"`
}

func (x *BattleSpaceDataSimple) Reset() {
	*x = BattleSpaceDataSimple{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BattleSpaceDataSimple) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BattleSpaceDataSimple) ProtoMessage() {}

func (x *BattleSpaceDataSimple) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BattleSpaceDataSimple.ProtoReflect.Descriptor instead.
func (*BattleSpaceDataSimple) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{3}
}

func (x *BattleSpaceDataSimple) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

func (x *BattleSpaceDataSimple) GetMapURI() string {
	if x != nil {
		return x.MapURI
	}
	return ""
}

func (x *BattleSpaceDataSimple) GetMasterUid() string {
	if x != nil {
		return x.MasterUid
	}
	return ""
}

func (x *BattleSpaceDataSimple) GetMasterIcon() string {
	if x != nil {
		return x.MasterIcon
	}
	return ""
}

func (x *BattleSpaceDataSimple) GetMasterDisplay() string {
	if x != nil {
		return x.MasterDisplay
	}
	return ""
}

func (x *BattleSpaceDataSimple) GetPlayers() []*BattleSpacePlayerSimple {
	if x != nil {
		return x.Players
	}
	return nil
}

type BattleSpacePlayer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid     string `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Display string `protobuf:"bytes,2,opt,name=display,proto3" json:"display,omitempty"`
	Icon    string `protobuf:"bytes,3,opt,name=icon,proto3" json:"icon,omitempty"`
	Pos     int32  `protobuf:"varint,4,opt,name=pos,proto3" json:"pos,omitempty"`
}

func (x *BattleSpacePlayer) Reset() {
	*x = BattleSpacePlayer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BattleSpacePlayer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BattleSpacePlayer) ProtoMessage() {}

func (x *BattleSpacePlayer) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BattleSpacePlayer.ProtoReflect.Descriptor instead.
func (*BattleSpacePlayer) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{4}
}

func (x *BattleSpacePlayer) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *BattleSpacePlayer) GetDisplay() string {
	if x != nil {
		return x.Display
	}
	return ""
}

func (x *BattleSpacePlayer) GetIcon() string {
	if x != nil {
		return x.Icon
	}
	return ""
}

func (x *BattleSpacePlayer) GetPos() int32 {
	if x != nil {
		return x.Pos
	}
	return 0
}

type BattleSpaceData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId   string               `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
	MapURI    string               `protobuf:"bytes,2,opt,name=mapURI,proto3" json:"mapURI,omitempty"`
	MasterUid string               `protobuf:"bytes,3,opt,name=masterUid,proto3" json:"masterUid,omitempty"`
	Starttime uint64               `protobuf:"fixed64,4,opt,name=starttime,proto3" json:"starttime,omitempty"`
	State     string               `protobuf:"bytes,5,opt,name=state,proto3" json:"state,omitempty"`
	Players   []*BattleSpacePlayer `protobuf:"bytes,6,rep,name=players,proto3" json:"players,omitempty"`
}

func (x *BattleSpaceData) Reset() {
	*x = BattleSpaceData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BattleSpaceData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BattleSpaceData) ProtoMessage() {}

func (x *BattleSpaceData) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BattleSpaceData.ProtoReflect.Descriptor instead.
func (*BattleSpaceData) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{5}
}

func (x *BattleSpaceData) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

func (x *BattleSpaceData) GetMapURI() string {
	if x != nil {
		return x.MapURI
	}
	return ""
}

func (x *BattleSpaceData) GetMasterUid() string {
	if x != nil {
		return x.MasterUid
	}
	return ""
}

func (x *BattleSpaceData) GetStarttime() uint64 {
	if x != nil {
		return x.Starttime
	}
	return 0
}

func (x *BattleSpaceData) GetState() string {
	if x != nil {
		return x.State
	}
	return ""
}

func (x *BattleSpaceData) GetPlayers() []*BattleSpacePlayer {
	if x != nil {
		return x.Players
	}
	return nil
}

type GetBattleSpaceList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start int32 `protobuf:"varint,1,opt,name=start,proto3" json:"start,omitempty"`
	Size  int32 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *GetBattleSpaceList) Reset() {
	*x = GetBattleSpaceList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBattleSpaceList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBattleSpaceList) ProtoMessage() {}

func (x *GetBattleSpaceList) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBattleSpaceList.ProtoReflect.Descriptor instead.
func (*GetBattleSpaceList) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{6}
}

func (x *GetBattleSpaceList) GetStart() int32 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *GetBattleSpaceList) GetSize() int32 {
	if x != nil {
		return x.Size
	}
	return 0
}

type GetBattleSpaceListResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start  int32                    `protobuf:"varint,1,opt,name=start,proto3" json:"start,omitempty"`
	Count  int32                    `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
	Spaces []*BattleSpaceDataSimple `protobuf:"bytes,3,rep,name=spaces,proto3" json:"spaces,omitempty"`
}

func (x *GetBattleSpaceListResp) Reset() {
	*x = GetBattleSpaceListResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBattleSpaceListResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBattleSpaceListResp) ProtoMessage() {}

func (x *GetBattleSpaceListResp) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBattleSpaceListResp.ProtoReflect.Descriptor instead.
func (*GetBattleSpaceListResp) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{7}
}

func (x *GetBattleSpaceListResp) GetStart() int32 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *GetBattleSpaceListResp) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *GetBattleSpaceListResp) GetSpaces() []*BattleSpaceDataSimple {
	if x != nil {
		return x.Spaces
	}
	return nil
}

type EnterBattleSpace struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId string `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
}

func (x *EnterBattleSpace) Reset() {
	*x = EnterBattleSpace{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EnterBattleSpace) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnterBattleSpace) ProtoMessage() {}

func (x *EnterBattleSpace) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnterBattleSpace.ProtoReflect.Descriptor instead.
func (*EnterBattleSpace) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{8}
}

func (x *EnterBattleSpace) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

type EnterBattleSpaceResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Space *BattleSpaceData `protobuf:"bytes,1,opt,name=space,proto3" json:"space,omitempty"`
}

func (x *EnterBattleSpaceResp) Reset() {
	*x = EnterBattleSpaceResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EnterBattleSpaceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnterBattleSpaceResp) ProtoMessage() {}

func (x *EnterBattleSpaceResp) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnterBattleSpaceResp.ProtoReflect.Descriptor instead.
func (*EnterBattleSpaceResp) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{9}
}

func (x *EnterBattleSpaceResp) GetSpace() *BattleSpaceData {
	if x != nil {
		return x.Space
	}
	return nil
}

type ReadyBattleSpace struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId string `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
	Uid     string `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`
	Ready   string `protobuf:"bytes,3,opt,name=ready,proto3" json:"ready,omitempty"`
}

func (x *ReadyBattleSpace) Reset() {
	*x = ReadyBattleSpace{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadyBattleSpace) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadyBattleSpace) ProtoMessage() {}

func (x *ReadyBattleSpace) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadyBattleSpace.ProtoReflect.Descriptor instead.
func (*ReadyBattleSpace) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{10}
}

func (x *ReadyBattleSpace) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

func (x *ReadyBattleSpace) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *ReadyBattleSpace) GetReady() string {
	if x != nil {
		return x.Ready
	}
	return ""
}

type ReadyBattleSpaceResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId string `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
	Uid     string `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`
	Ready   string `protobuf:"bytes,3,opt,name=ready,proto3" json:"ready,omitempty"`
}

func (x *ReadyBattleSpaceResp) Reset() {
	*x = ReadyBattleSpaceResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadyBattleSpaceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadyBattleSpaceResp) ProtoMessage() {}

func (x *ReadyBattleSpaceResp) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadyBattleSpaceResp.ProtoReflect.Descriptor instead.
func (*ReadyBattleSpaceResp) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{11}
}

func (x *ReadyBattleSpaceResp) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

func (x *ReadyBattleSpaceResp) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *ReadyBattleSpaceResp) GetReady() string {
	if x != nil {
		return x.Ready
	}
	return ""
}

//解散房间
type DissBattleSpaceNotify struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId string `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
}

func (x *DissBattleSpaceNotify) Reset() {
	*x = DissBattleSpaceNotify{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DissBattleSpaceNotify) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DissBattleSpaceNotify) ProtoMessage() {}

func (x *DissBattleSpaceNotify) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DissBattleSpaceNotify.ProtoReflect.Descriptor instead.
func (*DissBattleSpaceNotify) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{12}
}

func (x *DissBattleSpaceNotify) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

//开始战斗
type RequsetStartBattleSpace struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId string `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
}

func (x *RequsetStartBattleSpace) Reset() {
	*x = RequsetStartBattleSpace{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequsetStartBattleSpace) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequsetStartBattleSpace) ProtoMessage() {}

func (x *RequsetStartBattleSpace) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequsetStartBattleSpace.ProtoReflect.Descriptor instead.
func (*RequsetStartBattleSpace) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{13}
}

func (x *RequsetStartBattleSpace) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

type RequsetStartBattleSpaceResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpaceId string `protobuf:"bytes,1,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
}

func (x *RequsetStartBattleSpaceResp) Reset() {
	*x = RequsetStartBattleSpaceResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_battle_proto_msgTypes[14]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequsetStartBattleSpaceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequsetStartBattleSpaceResp) ProtoMessage() {}

func (x *RequsetStartBattleSpaceResp) ProtoReflect() protoreflect.Message {
	mi := &file_battle_proto_msgTypes[14]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequsetStartBattleSpaceResp.ProtoReflect.Descriptor instead.
func (*RequsetStartBattleSpaceResp) Descriptor() ([]byte, []int) {
	return file_battle_proto_rawDescGZIP(), []int{14}
}

func (x *RequsetStartBattleSpaceResp) GetSpaceId() string {
	if x != nil {
		return x.SpaceId
	}
	return ""
}

var File_battle_proto protoreflect.FileDescriptor

var file_battle_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04,
	0x70, 0x75, 0x62, 0x73, 0x22, 0x47, 0x0a, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x42, 0x61,
	0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x61, 0x70,
	0x55, 0x52, 0x49, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x61, 0x70, 0x55, 0x52,
	0x49, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x61, 0x78, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x07, 0x52, 0x08, 0x6d, 0x61, 0x78, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x49, 0x0a,
	0x15, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61,
	0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64,
	0x12, 0x16, 0x0a, 0x06, 0x6d, 0x61, 0x70, 0x55, 0x52, 0x49, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x6d, 0x61, 0x70, 0x55, 0x52, 0x49, 0x22, 0x45, 0x0a, 0x17, 0x42, 0x61, 0x74, 0x74,
	0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x53, 0x69, 0x6d,
	0x70, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x70, 0x6f, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x70, 0x6f, 0x73, 0x22,
	0xe6, 0x01, 0x0a, 0x15, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x44,
	0x61, 0x74, 0x61, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x70, 0x61,
	0x63, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x61, 0x70, 0x55, 0x52, 0x49, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x61, 0x70, 0x55, 0x52, 0x49, 0x12, 0x1c, 0x0a, 0x09, 0x6d,
	0x61, 0x73, 0x74, 0x65, 0x72, 0x55, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x55, 0x69, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x6d, 0x61, 0x73,
	0x74, 0x65, 0x72, 0x49, 0x63, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d,
	0x61, 0x73, 0x74, 0x65, 0x72, 0x49, 0x63, 0x6f, 0x6e, 0x12, 0x24, 0x0a, 0x0d, 0x6d, 0x61, 0x73,
	0x74, 0x65, 0x72, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x12,
	0x37, 0x0a, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1d, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x2e, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70,
	0x61, 0x63, 0x65, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52,
	0x07, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x22, 0x65, 0x0a, 0x11, 0x42, 0x61, 0x74, 0x74,
	0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12,
	0x18, 0x0a, 0x07, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x69, 0x63, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x69, 0x63, 0x6f, 0x6e, 0x12, 0x10, 0x0a,
	0x03, 0x70, 0x6f, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x70, 0x6f, 0x73, 0x22,
	0xc8, 0x01, 0x0a, 0x0f, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x12, 0x16, 0x0a,
	0x06, 0x6d, 0x61, 0x70, 0x55, 0x52, 0x49, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d,
	0x61, 0x70, 0x55, 0x52, 0x49, 0x12, 0x1c, 0x0a, 0x09, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x55,
	0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72,
	0x55, 0x69, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x06, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x74, 0x69, 0x6d,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x31, 0x0a, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x65,
	0x72, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x2e,
	0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x50, 0x6c, 0x61, 0x79, 0x65,
	0x72, 0x52, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x22, 0x3e, 0x0a, 0x12, 0x47, 0x65,
	0x74, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x4c, 0x69, 0x73, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x79, 0x0a, 0x16, 0x47, 0x65,
	0x74, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x4c, 0x69, 0x73, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x33, 0x0a, 0x06, 0x73, 0x70, 0x61, 0x63, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1b, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x2e, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70,
	0x61, 0x63, 0x65, 0x44, 0x61, 0x74, 0x61, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x06, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x73, 0x22, 0x2c, 0x0a, 0x10, 0x45, 0x6e, 0x74, 0x65, 0x72, 0x42, 0x61,
	0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x70, 0x61,
	0x63, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x49, 0x64, 0x22, 0x43, 0x0a, 0x14, 0x45, 0x6e, 0x74, 0x65, 0x72, 0x42, 0x61, 0x74, 0x74,
	0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x2b, 0x0a, 0x05, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x70, 0x75, 0x62,
	0x73, 0x2e, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x44, 0x61, 0x74,
	0x61, 0x52, 0x05, 0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x54, 0x0a, 0x10, 0x52, 0x65, 0x61, 0x64,
	0x79, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x65, 0x61, 0x64,
	0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x65, 0x61, 0x64, 0x79, 0x22, 0x58,
	0x0a, 0x14, 0x52, 0x65, 0x61, 0x64, 0x79, 0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61,
	0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64,
	0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75,
	0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x65, 0x61, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x72, 0x65, 0x61, 0x64, 0x79, 0x22, 0x31, 0x0a, 0x15, 0x44, 0x69, 0x73, 0x73,
	0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x79, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x22, 0x33, 0x0a, 0x17, 0x52,
	0x65, 0x71, 0x75, 0x73, 0x65, 0x74, 0x53, 0x74, 0x61, 0x72, 0x74, 0x42, 0x61, 0x74, 0x74, 0x6c,
	0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64,
	0x22, 0x37, 0x0a, 0x1b, 0x52, 0x65, 0x71, 0x75, 0x73, 0x65, 0x74, 0x53, 0x74, 0x61, 0x72, 0x74,
	0x42, 0x61, 0x74, 0x74, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12,
	0x18, 0x0a, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x42, 0x08, 0x5a, 0x06, 0x2e, 0x3b, 0x70,
	0x75, 0x62, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_battle_proto_rawDescOnce sync.Once
	file_battle_proto_rawDescData = file_battle_proto_rawDesc
)

func file_battle_proto_rawDescGZIP() []byte {
	file_battle_proto_rawDescOnce.Do(func() {
		file_battle_proto_rawDescData = protoimpl.X.CompressGZIP(file_battle_proto_rawDescData)
	})
	return file_battle_proto_rawDescData
}

var file_battle_proto_msgTypes = make([]protoimpl.MessageInfo, 15)
var file_battle_proto_goTypes = []interface{}{
	(*CreateBattleSpace)(nil),           // 0: pubs.CreateBattleSpace
	(*CreateBattleSpaceResp)(nil),       // 1: pubs.CreateBattleSpaceResp
	(*BattleSpacePlayerSimple)(nil),     // 2: pubs.BattleSpacePlayerSimple
	(*BattleSpaceDataSimple)(nil),       // 3: pubs.BattleSpaceDataSimple
	(*BattleSpacePlayer)(nil),           // 4: pubs.BattleSpacePlayer
	(*BattleSpaceData)(nil),             // 5: pubs.BattleSpaceData
	(*GetBattleSpaceList)(nil),          // 6: pubs.GetBattleSpaceList
	(*GetBattleSpaceListResp)(nil),      // 7: pubs.GetBattleSpaceListResp
	(*EnterBattleSpace)(nil),            // 8: pubs.EnterBattleSpace
	(*EnterBattleSpaceResp)(nil),        // 9: pubs.EnterBattleSpaceResp
	(*ReadyBattleSpace)(nil),            // 10: pubs.ReadyBattleSpace
	(*ReadyBattleSpaceResp)(nil),        // 11: pubs.ReadyBattleSpaceResp
	(*DissBattleSpaceNotify)(nil),       // 12: pubs.DissBattleSpaceNotify
	(*RequsetStartBattleSpace)(nil),     // 13: pubs.RequsetStartBattleSpace
	(*RequsetStartBattleSpaceResp)(nil), // 14: pubs.RequsetStartBattleSpaceResp
}
var file_battle_proto_depIdxs = []int32{
	2, // 0: pubs.BattleSpaceDataSimple.players:type_name -> pubs.BattleSpacePlayerSimple
	4, // 1: pubs.BattleSpaceData.players:type_name -> pubs.BattleSpacePlayer
	3, // 2: pubs.GetBattleSpaceListResp.spaces:type_name -> pubs.BattleSpaceDataSimple
	5, // 3: pubs.EnterBattleSpaceResp.space:type_name -> pubs.BattleSpaceData
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_battle_proto_init() }
func file_battle_proto_init() {
	if File_battle_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_battle_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateBattleSpace); i {
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
		file_battle_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateBattleSpaceResp); i {
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
		file_battle_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BattleSpacePlayerSimple); i {
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
		file_battle_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BattleSpaceDataSimple); i {
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
		file_battle_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BattleSpacePlayer); i {
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
		file_battle_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BattleSpaceData); i {
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
		file_battle_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBattleSpaceList); i {
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
		file_battle_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBattleSpaceListResp); i {
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
		file_battle_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EnterBattleSpace); i {
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
		file_battle_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EnterBattleSpaceResp); i {
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
		file_battle_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadyBattleSpace); i {
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
		file_battle_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadyBattleSpaceResp); i {
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
		file_battle_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DissBattleSpaceNotify); i {
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
		file_battle_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequsetStartBattleSpace); i {
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
		file_battle_proto_msgTypes[14].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequsetStartBattleSpaceResp); i {
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
			RawDescriptor: file_battle_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   15,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_battle_proto_goTypes,
		DependencyIndexes: file_battle_proto_depIdxs,
		MessageInfos:      file_battle_proto_msgTypes,
	}.Build()
	File_battle_proto = out.File
	file_battle_proto_rawDesc = nil
	file_battle_proto_goTypes = nil
	file_battle_proto_depIdxs = nil
}
