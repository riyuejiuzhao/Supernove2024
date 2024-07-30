// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.27.0
// source: miniRouter.proto

package miniRouterProto

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

	InstanceID string `protobuf:"bytes,1,opt,name=InstanceID,proto3" json:"InstanceID,omitempty"`
	Host       string `protobuf:"bytes,2,opt,name=Host,proto3" json:"Host,omitempty"`
	Port       int32  `protobuf:"varint,3,opt,name=Port,proto3" json:"Port,omitempty"`
	Weight     int32  `protobuf:"varint,4,opt,name=Weight,proto3" json:"Weight,omitempty"`
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

func (x *InstanceInfo) GetInstanceID() string {
	if x != nil {
		return x.InstanceID
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

type ServiceInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName string          `protobuf:"bytes,1,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	Revision    int64           `protobuf:"varint,2,opt,name=Revision,proto3" json:"Revision,omitempty"`
	Instances   []*InstanceInfo `protobuf:"bytes,3,rep,name=Instances,proto3" json:"Instances,omitempty"`
}

func (x *ServiceInfo) Reset() {
	*x = ServiceInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceInfo) ProtoMessage() {}

func (x *ServiceInfo) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use ServiceInfo.ProtoReflect.Descriptor instead.
func (*ServiceInfo) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{1}
}

func (x *ServiceInfo) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *ServiceInfo) GetRevision() int64 {
	if x != nil {
		return x.Revision
	}
	return 0
}

func (x *ServiceInfo) GetInstances() []*InstanceInfo {
	if x != nil {
		return x.Instances
	}
	return nil
}

type RegisterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName string `protobuf:"bytes,1,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	Host        string `protobuf:"bytes,2,opt,name=Host,proto3" json:"Host,omitempty"`
	Port        int32  `protobuf:"varint,3,opt,name=Port,proto3" json:"Port,omitempty"`
	Weight      int32  `protobuf:"varint,4,opt,name=Weight,proto3" json:"Weight,omitempty"` // 权重
	TTL         int64  `protobuf:"varint,5,opt,name=TTL,proto3" json:"TTL,omitempty"`       //上报时间间隔
}

func (x *RegisterRequest) Reset() {
	*x = RegisterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterRequest) ProtoMessage() {}

func (x *RegisterRequest) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use RegisterRequest.ProtoReflect.Descriptor instead.
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{2}
}

func (x *RegisterRequest) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *RegisterRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *RegisterRequest) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *RegisterRequest) GetWeight() int32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *RegisterRequest) GetTTL() int64 {
	if x != nil {
		return x.TTL
	}
	return 0
}

type RegisterReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceID string `protobuf:"bytes,1,opt,name=InstanceID,proto3" json:"InstanceID,omitempty"`
	Existed    bool   `protobuf:"varint,2,opt,name=Existed,proto3" json:"Existed,omitempty"`
}

func (x *RegisterReply) Reset() {
	*x = RegisterReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterReply) ProtoMessage() {}

func (x *RegisterReply) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use RegisterReply.ProtoReflect.Descriptor instead.
func (*RegisterReply) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{3}
}

func (x *RegisterReply) GetInstanceID() string {
	if x != nil {
		return x.InstanceID
	}
	return ""
}

func (x *RegisterReply) GetExisted() bool {
	if x != nil {
		return x.Existed
	}
	return false
}

type DeregisterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName string `protobuf:"bytes,1,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	InstanceID  string `protobuf:"bytes,2,opt,name=InstanceID,proto3" json:"InstanceID,omitempty"`
	Host        string `protobuf:"bytes,3,opt,name=Host,proto3" json:"Host,omitempty"`
	Port        int32  `protobuf:"varint,4,opt,name=Port,proto3" json:"Port,omitempty"`
}

func (x *DeregisterRequest) Reset() {
	*x = DeregisterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeregisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeregisterRequest) ProtoMessage() {}

func (x *DeregisterRequest) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use DeregisterRequest.ProtoReflect.Descriptor instead.
func (*DeregisterRequest) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{4}
}

func (x *DeregisterRequest) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *DeregisterRequest) GetInstanceID() string {
	if x != nil {
		return x.InstanceID
	}
	return ""
}

func (x *DeregisterRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *DeregisterRequest) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

type DeregisterReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeregisterReply) Reset() {
	*x = DeregisterReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeregisterReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeregisterReply) ProtoMessage() {}

func (x *DeregisterReply) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use DeregisterReply.ProtoReflect.Descriptor instead.
func (*DeregisterReply) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{5}
}

type GetInstancesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName string `protobuf:"bytes,1,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	Revision    int64  `protobuf:"varint,2,opt,name=Revision,proto3" json:"Revision,omitempty"`
	CheckHealth bool   `protobuf:"varint,3,opt,name=CheckHealth,proto3" json:"CheckHealth,omitempty"`
}

func (x *GetInstancesRequest) Reset() {
	*x = GetInstancesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetInstancesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInstancesRequest) ProtoMessage() {}

func (x *GetInstancesRequest) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use GetInstancesRequest.ProtoReflect.Descriptor instead.
func (*GetInstancesRequest) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{6}
}

func (x *GetInstancesRequest) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *GetInstancesRequest) GetRevision() int64 {
	if x != nil {
		return x.Revision
	}
	return 0
}

func (x *GetInstancesRequest) GetCheckHealth() bool {
	if x != nil {
		return x.CheckHealth
	}
	return false
}

type GetInstancesReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Revision  int64           `protobuf:"varint,1,opt,name=Revision,proto3" json:"Revision,omitempty"`
	Instances []*InstanceInfo `protobuf:"bytes,2,rep,name=Instances,proto3" json:"Instances,omitempty"`
}

func (x *GetInstancesReply) Reset() {
	*x = GetInstancesReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetInstancesReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInstancesReply) ProtoMessage() {}

func (x *GetInstancesReply) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInstancesReply.ProtoReflect.Descriptor instead.
func (*GetInstancesReply) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{7}
}

func (x *GetInstancesReply) GetRevision() int64 {
	if x != nil {
		return x.Revision
	}
	return 0
}

func (x *GetInstancesReply) GetInstances() []*InstanceInfo {
	if x != nil {
		return x.Instances
	}
	return nil
}

type HeartBeatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName string `protobuf:"bytes,1,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	InstanceID  string `protobuf:"bytes,2,opt,name=InstanceID,proto3" json:"InstanceID,omitempty"`
}

func (x *HeartBeatRequest) Reset() {
	*x = HeartBeatRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeartBeatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartBeatRequest) ProtoMessage() {}

func (x *HeartBeatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartBeatRequest.ProtoReflect.Descriptor instead.
func (*HeartBeatRequest) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{8}
}

func (x *HeartBeatRequest) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *HeartBeatRequest) GetInstanceID() string {
	if x != nil {
		return x.InstanceID
	}
	return ""
}

type HeartBeatReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HeartBeatReply) Reset() {
	*x = HeartBeatReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeartBeatReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartBeatReply) ProtoMessage() {}

func (x *HeartBeatReply) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartBeatReply.ProtoReflect.Descriptor instead.
func (*HeartBeatReply) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{9}
}

type InstanceHealthInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceID    string `protobuf:"bytes,1,opt,name=InstanceID,proto3" json:"InstanceID,omitempty"`
	TTL           int64  `protobuf:"varint,2,opt,name=TTL,proto3" json:"TTL,omitempty"`
	LastHeartBeat int64  `protobuf:"varint,3,opt,name=LastHeartBeat,proto3" json:"LastHeartBeat,omitempty"`
}

func (x *InstanceHealthInfo) Reset() {
	*x = InstanceHealthInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InstanceHealthInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InstanceHealthInfo) ProtoMessage() {}

func (x *InstanceHealthInfo) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InstanceHealthInfo.ProtoReflect.Descriptor instead.
func (*InstanceHealthInfo) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{10}
}

func (x *InstanceHealthInfo) GetInstanceID() string {
	if x != nil {
		return x.InstanceID
	}
	return ""
}

func (x *InstanceHealthInfo) GetTTL() int64 {
	if x != nil {
		return x.TTL
	}
	return 0
}

func (x *InstanceHealthInfo) GetLastHeartBeat() int64 {
	if x != nil {
		return x.LastHeartBeat
	}
	return 0
}

type ServiceHealthInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName        string                `protobuf:"bytes,1,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	InstanceHealthInfo []*InstanceHealthInfo `protobuf:"bytes,2,rep,name=InstanceHealthInfo,proto3" json:"InstanceHealthInfo,omitempty"`
}

func (x *ServiceHealthInfo) Reset() {
	*x = ServiceHealthInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceHealthInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceHealthInfo) ProtoMessage() {}

func (x *ServiceHealthInfo) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceHealthInfo.ProtoReflect.Descriptor instead.
func (*ServiceHealthInfo) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{11}
}

func (x *ServiceHealthInfo) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *ServiceHealthInfo) GetInstanceHealthInfo() []*InstanceHealthInfo {
	if x != nil {
		return x.InstanceHealthInfo
	}
	return nil
}

type GetHealthInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceNames []string `protobuf:"bytes,1,rep,name=ServiceNames,proto3" json:"ServiceNames,omitempty"`
}

func (x *GetHealthInfoRequest) Reset() {
	*x = GetHealthInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetHealthInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetHealthInfoRequest) ProtoMessage() {}

func (x *GetHealthInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetHealthInfoRequest.ProtoReflect.Descriptor instead.
func (*GetHealthInfoRequest) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{12}
}

func (x *GetHealthInfoRequest) GetServiceNames() []string {
	if x != nil {
		return x.ServiceNames
	}
	return nil
}

type GetHealthInfoReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HealthInfos []*ServiceHealthInfo `protobuf:"bytes,1,rep,name=HealthInfos,proto3" json:"HealthInfos,omitempty"`
}

func (x *GetHealthInfoReply) Reset() {
	*x = GetHealthInfoReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniRouter_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetHealthInfoReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetHealthInfoReply) ProtoMessage() {}

func (x *GetHealthInfoReply) ProtoReflect() protoreflect.Message {
	mi := &file_miniRouter_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetHealthInfoReply.ProtoReflect.Descriptor instead.
func (*GetHealthInfoReply) Descriptor() ([]byte, []int) {
	return file_miniRouter_proto_rawDescGZIP(), []int{13}
}

func (x *GetHealthInfoReply) GetHealthInfos() []*ServiceHealthInfo {
	if x != nil {
		return x.HealthInfos
	}
	return nil
}

var File_miniRouter_proto protoreflect.FileDescriptor

var file_miniRouter_proto_rawDesc = []byte{
	0x0a, 0x10, 0x6d, 0x69, 0x6e, 0x69, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x6e, 0x0a, 0x0c, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x1e, 0x0a, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x57, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x57, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x22, 0x78, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x2b, 0x0a, 0x09, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x09, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x22, 0x85, 0x01, 0x0a,
	0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x57, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x57, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x54, 0x54, 0x4c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x03, 0x54, 0x54, 0x4c, 0x22, 0x49, 0x0a, 0x0d, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x1e, 0x0a, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x45, 0x78, 0x69, 0x73, 0x74, 0x65, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x45, 0x78, 0x69, 0x73, 0x74, 0x65, 0x64, 0x22,
	0x7d, 0x0a, 0x11, 0x44, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x49, 0x6e, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x50, 0x6f,
	0x72, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x22, 0x11,
	0x0a, 0x0f, 0x44, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x75, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x52, 0x65,
	0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x52, 0x65,
	0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x43, 0x68, 0x65,
	0x63, 0x6b, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x22, 0x5c, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x1a, 0x0a,
	0x08, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x08, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x2b, 0x0a, 0x09, 0x49, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x09, 0x49, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x22, 0x54, 0x0a, 0x10, 0x48, 0x65, 0x61, 0x72, 0x74, 0x42,
	0x65, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x44, 0x22, 0x10, 0x0a, 0x0e,
	0x48, 0x65, 0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x6c,
	0x0a, 0x12, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1e, 0x0a, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x49, 0x44, 0x12, 0x10, 0x0a, 0x03, 0x54, 0x54, 0x4c, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x03, 0x54, 0x54, 0x4c, 0x12, 0x24, 0x0a, 0x0d, 0x4c, 0x61, 0x73, 0x74, 0x48, 0x65,
	0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x4c,
	0x61, 0x73, 0x74, 0x48, 0x65, 0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x22, 0x7a, 0x0a, 0x11,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x43, 0x0a, 0x12, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x12, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x3a, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x22, 0x0a, 0x0c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x73, 0x22, 0x4a, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x34, 0x0a, 0x0b, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x0b, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x73,
	0x32, 0x77, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x2e, 0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12,
	0x10, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x0e, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x0a, 0x44, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65,
	0x72, 0x12, 0x12, 0x2e, 0x44, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e, 0x44, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x32, 0x4e, 0x0a, 0x10, 0x44, 0x69, 0x73,
	0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3a, 0x0a,
	0x0c, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x12, 0x14, 0x2e,
	0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x32, 0x81, 0x01, 0x0a, 0x0d, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x31, 0x0a, 0x09, 0x48,
	0x65, 0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x12, 0x11, 0x2e, 0x48, 0x65, 0x61, 0x72, 0x74,
	0x42, 0x65, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x48, 0x65,
	0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x3d,
	0x0a, 0x0d, 0x47, 0x65, 0x74, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x15, 0x2e, 0x47, 0x65, 0x74, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x47, 0x65, 0x74, 0x48, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x13, 0x5a,
	0x11, 0x2e, 0x3b, 0x6d, 0x69, 0x6e, 0x69, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_miniRouter_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_miniRouter_proto_goTypes = []interface{}{
	(*InstanceInfo)(nil),         // 0: InstanceInfo
	(*ServiceInfo)(nil),          // 1: ServiceInfo
	(*RegisterRequest)(nil),      // 2: RegisterRequest
	(*RegisterReply)(nil),        // 3: RegisterReply
	(*DeregisterRequest)(nil),    // 4: DeregisterRequest
	(*DeregisterReply)(nil),      // 5: DeregisterReply
	(*GetInstancesRequest)(nil),  // 6: GetInstancesRequest
	(*GetInstancesReply)(nil),    // 7: GetInstancesReply
	(*HeartBeatRequest)(nil),     // 8: HeartBeatRequest
	(*HeartBeatReply)(nil),       // 9: HeartBeatReply
	(*InstanceHealthInfo)(nil),   // 10: InstanceHealthInfo
	(*ServiceHealthInfo)(nil),    // 11: ServiceHealthInfo
	(*GetHealthInfoRequest)(nil), // 12: GetHealthInfoRequest
	(*GetHealthInfoReply)(nil),   // 13: GetHealthInfoReply
}
var file_miniRouter_proto_depIdxs = []int32{
	0,  // 0: ServiceInfo.Instances:type_name -> InstanceInfo
	0,  // 1: GetInstancesReply.Instances:type_name -> InstanceInfo
	10, // 2: ServiceHealthInfo.InstanceHealthInfo:type_name -> InstanceHealthInfo
	11, // 3: GetHealthInfoReply.HealthInfos:type_name -> ServiceHealthInfo
	2,  // 4: RegisterService.Register:input_type -> RegisterRequest
	4,  // 5: RegisterService.Deregister:input_type -> DeregisterRequest
	6,  // 6: DiscoveryService.GetInstances:input_type -> GetInstancesRequest
	8,  // 7: HealthService.HeartBeat:input_type -> HeartBeatRequest
	12, // 8: HealthService.GetHealthInfo:input_type -> GetHealthInfoRequest
	3,  // 9: RegisterService.Register:output_type -> RegisterReply
	5,  // 10: RegisterService.Deregister:output_type -> DeregisterReply
	7,  // 11: DiscoveryService.GetInstances:output_type -> GetInstancesReply
	9,  // 12: HealthService.HeartBeat:output_type -> HeartBeatReply
	13, // 13: HealthService.GetHealthInfo:output_type -> GetHealthInfoReply
	9,  // [9:14] is the sub-list for method output_type
	4,  // [4:9] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
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
			switch v := v.(*ServiceInfo); i {
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
			switch v := v.(*RegisterRequest); i {
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
			switch v := v.(*RegisterReply); i {
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
			switch v := v.(*DeregisterRequest); i {
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
			switch v := v.(*DeregisterReply); i {
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
			switch v := v.(*GetInstancesRequest); i {
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
		file_miniRouter_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetInstancesReply); i {
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
		file_miniRouter_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeartBeatRequest); i {
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
		file_miniRouter_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeartBeatReply); i {
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
		file_miniRouter_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InstanceHealthInfo); i {
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
		file_miniRouter_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceHealthInfo); i {
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
		file_miniRouter_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetHealthInfoRequest); i {
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
		file_miniRouter_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetHealthInfoReply); i {
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
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   3,
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
