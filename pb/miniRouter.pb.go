// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.27.0
// source: miniRouter.proto

package pb

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

type InstanceInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceID int64  `protobuf:"varint,1,opt,name=InstanceID,proto3" json:"InstanceID,omitempty"`
	Name       string `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	Host       string `protobuf:"bytes,3,opt,name=Host,proto3" json:"Host,omitempty"`
	Port       int32  `protobuf:"varint,4,opt,name=Port,proto3" json:"Port,omitempty"`
	Weight     int32  `protobuf:"varint,5,opt,name=Weight,proto3" json:"Weight,omitempty"`
	CreateTime int64  `protobuf:"varint,6,opt,name=CreateTime,proto3" json:"CreateTime,omitempty"`
}

func (x *InstanceInfo) Reset() {
	*x = InstanceInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InstanceInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InstanceInfo) ProtoMessage() {}

func (x *InstanceInfo) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InstanceInfo.ProtoReflect.Descriptor instead.
func (*InstanceInfo) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{0}
}

func (x *InstanceInfo) GetInstanceID() int64 {
	if x != nil {
		return x.InstanceID
	}
	return 0
}

func (x *InstanceInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *InstanceInfo) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *InstanceInfo) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *InstanceInfo) GetWeight() int32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *InstanceInfo) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

type TargetRouterInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RouterID        int64  `protobuf:"varint,1,opt,name=RouterID,proto3" json:"RouterID,omitempty"`
	SrcInstanceName string `protobuf:"bytes,2,opt,name=SrcInstanceName,proto3" json:"SrcInstanceName,omitempty"`
	DstInstanceName string `protobuf:"bytes,3,opt,name=DstInstanceName,proto3" json:"DstInstanceName,omitempty"`
	Timeout         int64  `protobuf:"varint,4,opt,name=Timeout,proto3" json:"Timeout,omitempty"`
	CreateTime      int64  `protobuf:"varint,6,opt,name=CreateTime,proto3" json:"CreateTime,omitempty"`
}

func (x *TargetRouterInfo) Reset() {
	*x = TargetRouterInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TargetRouterInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TargetRouterInfo) ProtoMessage() {}

func (x *TargetRouterInfo) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TargetRouterInfo.ProtoReflect.Descriptor instead.
func (*TargetRouterInfo) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{1}
}

func (x *TargetRouterInfo) GetRouterID() int64 {
	if x != nil {
		return x.RouterID
	}
	return 0
}

func (x *TargetRouterInfo) GetSrcInstanceName() string {
	if x != nil {
		return x.SrcInstanceName
	}
	return ""
}

func (x *TargetRouterInfo) GetDstInstanceName() string {
	if x != nil {
		return x.DstInstanceName
	}
	return ""
}

func (x *TargetRouterInfo) GetTimeout() int64 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *TargetRouterInfo) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

type KVRouterInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RouterID        int64             `protobuf:"varint,1,opt,name=RouterID,proto3" json:"RouterID,omitempty"`
	Weight          int32             `protobuf:"varint,2,opt,name=weight,proto3" json:"weight,omitempty"`
	Dic             map[string]string `protobuf:"bytes,3,rep,name=Dic,proto3" json:"Dic,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	DstInstanceName []string          `protobuf:"bytes,4,rep,name=DstInstanceName,proto3" json:"DstInstanceName,omitempty"`
	Timeout         int64             `protobuf:"varint,5,opt,name=Timeout,proto3" json:"Timeout,omitempty"`
	CreateTime      int64             `protobuf:"varint,6,opt,name=CreateTime,proto3" json:"CreateTime,omitempty"`
	// 如果有多个，那么允许指定更进一步的筛选方式
	RouterType int32 `protobuf:"varint,7,opt,name=RouterType,proto3" json:"RouterType,omitempty"`
}

func (x *KVRouterInfo) Reset() {
	*x = KVRouterInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVRouterInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVRouterInfo) ProtoMessage() {}

func (x *KVRouterInfo) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVRouterInfo.ProtoReflect.Descriptor instead.
func (*KVRouterInfo) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{2}
}

func (x *KVRouterInfo) GetRouterID() int64 {
	if x != nil {
		return x.RouterID
	}
	return 0
}

func (x *KVRouterInfo) GetWeight() int32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *KVRouterInfo) GetDic() map[string]string {
	if x != nil {
		return x.Dic
	}
	return nil
}

func (x *KVRouterInfo) GetDstInstanceName() []string {
	if x != nil {
		return x.DstInstanceName
	}
	return nil
}

func (x *KVRouterInfo) GetTimeout() int64 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *KVRouterInfo) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

func (x *KVRouterInfo) GetRouterType() int32 {
	if x != nil {
		return x.RouterType
	}
	return 0
}

type ClusterInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name    string   `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Address []string `protobuf:"bytes,2,rep,name=Address,proto3" json:"Address,omitempty"`
}

func (x *ClusterInfo) Reset() {
	*x = ClusterInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterInfo) ProtoMessage() {}

func (x *ClusterInfo) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterInfo.ProtoReflect.Descriptor instead.
func (*ClusterInfo) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{3}
}

func (x *ClusterInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ClusterInfo) GetAddress() []string {
	if x != nil {
		return x.Address
	}
	return nil
}

type InitArgv struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *InitArgv) Reset() {
	*x = InitArgv{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InitArgv) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitArgv) ProtoMessage() {}

func (x *InitArgv) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitArgv.ProtoReflect.Descriptor instead.
func (*InitArgv) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{4}
}

// 路由表没有时间限制
// 说到底一个服务一个也不会很多，一直不断的用heartbeat续也不合理
type RouterTableInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName string   `protobuf:"bytes,2,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	Tags        []string `protobuf:"bytes,3,rep,name=Tags,proto3" json:"Tags,omitempty"`
}

func (x *RouterTableInfo) Reset() {
	*x = RouterTableInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RouterTableInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RouterTableInfo) ProtoMessage() {}

func (x *RouterTableInfo) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RouterTableInfo.ProtoReflect.Descriptor instead.
func (*RouterTableInfo) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{5}
}

func (x *RouterTableInfo) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *RouterTableInfo) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

type InitResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Instances []*ClusterInfo `protobuf:"bytes,1,rep,name=Instances,proto3" json:"Instances,omitempty"`
	Routers   []*ClusterInfo `protobuf:"bytes,2,rep,name=Routers,proto3" json:"Routers,omitempty"`
}

func (x *InitResult) Reset() {
	*x = InitResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InitResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitResult) ProtoMessage() {}

func (x *InitResult) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitResult.ProtoReflect.Descriptor instead.
func (*InitResult) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{6}
}

func (x *InitResult) GetInstances() []*ClusterInfo {
	if x != nil {
		return x.Instances
	}
	return nil
}

func (x *InitResult) GetRouters() []*ClusterInfo {
	if x != nil {
		return x.Routers
	}
	return nil
}

var File_miniRouter_proto protoreflect.FileDescriptor

var file_miniRouter_proto_rawDesc = []byte{
	0x0a, 0x10, 0x6d, 0x69, 0x6e, 0x69, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xa2, 0x01, 0x0a, 0x0c, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x1e, 0x0a, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49,
	0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x50,
	0x6f, 0x72, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x57, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x06, 0x57, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x22, 0xbc, 0x01, 0x0a, 0x10, 0x54, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08,
	0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08,
	0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x44, 0x12, 0x28, 0x0a, 0x0f, 0x53, 0x72, 0x63, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0f, 0x53, 0x72, 0x63, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x28, 0x0a, 0x0f, 0x44, 0x73, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x44, 0x73, 0x74,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x54,
	0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x54, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x22, 0xa8, 0x02, 0x0a, 0x0c, 0x4b, 0x56, 0x52, 0x6f, 0x75,
	0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x52, 0x6f, 0x75, 0x74, 0x65,
	0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x52, 0x6f, 0x75, 0x74, 0x65,
	0x72, 0x49, 0x44, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x28, 0x0a, 0x03, 0x44,
	0x69, 0x63, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x4b, 0x56, 0x52, 0x6f, 0x75,
	0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x44, 0x69, 0x63, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x03, 0x44, 0x69, 0x63, 0x12, 0x28, 0x0a, 0x0f, 0x44, 0x73, 0x74, 0x49, 0x6e, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f,
	0x44, 0x73, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x52, 0x6f, 0x75,
	0x74, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x52,
	0x6f, 0x75, 0x74, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x1a, 0x36, 0x0a, 0x08, 0x44, 0x69, 0x63,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0x3b, 0x0a, 0x0b, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f,
	0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x0a,
	0x0a, 0x08, 0x49, 0x6e, 0x69, 0x74, 0x41, 0x72, 0x67, 0x76, 0x22, 0x47, 0x0a, 0x0f, 0x52, 0x6f,
	0x75, 0x74, 0x65, 0x72, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x20, 0x0a,
	0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x54, 0x61, 0x67, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x54,
	0x61, 0x67, 0x73, 0x22, 0x60, 0x0a, 0x0a, 0x49, 0x6e, 0x69, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x12, 0x2a, 0x0a, 0x09, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x6e,
	0x66, 0x6f, 0x52, 0x09, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x12, 0x26, 0x0a,
	0x07, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c,
	0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x07, 0x52, 0x6f,
	0x75, 0x74, 0x65, 0x72, 0x73, 0x32, 0x31, 0x0a, 0x0d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x20, 0x0a, 0x04, 0x49, 0x6e, 0x69, 0x74, 0x12, 0x09,
	0x2e, 0x49, 0x6e, 0x69, 0x74, 0x41, 0x72, 0x67, 0x76, 0x1a, 0x0b, 0x2e, 0x49, 0x6e, 0x69, 0x74,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x00, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x3b, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_miniRouter_proto_rawDescOnce sync.Once
	file_miniRouter_proto_rawDescData = file_miniRouter_proto_rawDesc
)

func file_miniRouter_proto_rawDescGZIP() []byte {
	file_miniRouter_proto_rawDescOnce.Do(func() {
		file_miniRouter_proto_rawDescData = protoimpl.X.CompressGZIP(file_miniRouter_proto_rawDescData)
	})
	return file_miniRouter_proto_rawDescData
}

var file_miniRouter_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_miniRouter_proto_goTypes = []interface{}{
	(*InstanceInfo)(nil),     // 0: InstanceInfo
	(*TargetRouterInfo)(nil), // 1: TargetRouterInfo
	(*KVRouterInfo)(nil),     // 2: KVRouterInfo
	(*ClusterInfo)(nil),      // 3: ClusterInfo
	(*InitArgv)(nil),         // 4: InitArgv
	(*RouterTableInfo)(nil),  // 5: RouterTableInfo
	(*InitResult)(nil),       // 6: InitResult
	nil,                      // 7: KVRouterInfo.DicEntry
}
var file_miniRouter_proto_depIdxs = []int32{
	7, // 0: KVRouterInfo.Dic:type_name -> KVRouterInfo.DicEntry
	3, // 1: InitResult.Instances:type_name -> ClusterInfo
	3, // 2: InitResult.Routers:type_name -> ClusterInfo
	4, // 3: ConfigService.Init:input_type -> InitArgv
	6, // 4: ConfigService.Init:output_type -> InitResult
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_miniRouter_proto_init() }
func file_miniRouter_proto_init() {
	if File_miniRouter_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_miniRouter_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InstanceInfo); i {
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
		file_miniRouter_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TargetRouterInfo); i {
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
		file_miniRouter_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KVRouterInfo); i {
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
		file_miniRouter_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterInfo); i {
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
		file_miniRouter_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InitArgv); i {
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
		file_miniRouter_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RouterTableInfo); i {
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
		file_miniRouter_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InitResult); i {
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
			RawDescriptor: file_miniRouter_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_miniRouter_proto_goTypes,
		DependencyIndexes: file_miniRouter_proto_depIdxs,
		MessageInfos:      file_miniRouter_proto_msgTypes,
	}.Build()
	File_miniRouter_proto = out.File
	file_miniRouter_proto_rawDesc = nil
	file_miniRouter_proto_goTypes = nil
	file_miniRouter_proto_depIdxs = nil
}
