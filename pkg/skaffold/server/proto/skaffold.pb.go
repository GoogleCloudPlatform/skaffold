// Code generated by protoc-gen-go. DO NOT EDIT.
// source: skaffold.proto

package proto

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type StateResponse struct {
	State                *State   `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StateResponse) Reset()         { *m = StateResponse{} }
func (m *StateResponse) String() string { return proto.CompactTextString(m) }
func (*StateResponse) ProtoMessage()    {}
func (*StateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{0}
}

func (m *StateResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StateResponse.Unmarshal(m, b)
}
func (m *StateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StateResponse.Marshal(b, m, deterministic)
}
func (m *StateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateResponse.Merge(m, src)
}
func (m *StateResponse) XXX_Size() int {
	return xxx_messageInfo_StateResponse.Size(m)
}
func (m *StateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_StateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_StateResponse proto.InternalMessageInfo

func (m *StateResponse) GetState() *State {
	if m != nil {
		return m.State
	}
	return nil
}

type Response struct {
	Msg                  string   `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{1}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

type Request struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{2}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type State struct {
	BuildState           *BuildState           `protobuf:"bytes,1,opt,name=buildState,proto3" json:"buildState,omitempty"`
	DeployState          *DeployState          `protobuf:"bytes,2,opt,name=deployState,proto3" json:"deployState,omitempty"`
	ForwardedPorts       map[string]*PortEvent `protobuf:"bytes,3,rep,name=forwardedPorts,proto3" json:"forwardedPorts,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *State) Reset()         { *m = State{} }
func (m *State) String() string { return proto.CompactTextString(m) }
func (*State) ProtoMessage()    {}
func (*State) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{3}
}

func (m *State) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_State.Unmarshal(m, b)
}
func (m *State) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_State.Marshal(b, m, deterministic)
}
func (m *State) XXX_Merge(src proto.Message) {
	xxx_messageInfo_State.Merge(m, src)
}
func (m *State) XXX_Size() int {
	return xxx_messageInfo_State.Size(m)
}
func (m *State) XXX_DiscardUnknown() {
	xxx_messageInfo_State.DiscardUnknown(m)
}

var xxx_messageInfo_State proto.InternalMessageInfo

func (m *State) GetBuildState() *BuildState {
	if m != nil {
		return m.BuildState
	}
	return nil
}

func (m *State) GetDeployState() *DeployState {
	if m != nil {
		return m.DeployState
	}
	return nil
}

func (m *State) GetForwardedPorts() map[string]*PortEvent {
	if m != nil {
		return m.ForwardedPorts
	}
	return nil
}

// BuildState contains a map of all skaffold artifacts to their current build
// states
type BuildState struct {
	Artifacts            map[string]string `protobuf:"bytes,1,rep,name=artifacts,proto3" json:"artifacts,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *BuildState) Reset()         { *m = BuildState{} }
func (m *BuildState) String() string { return proto.CompactTextString(m) }
func (*BuildState) ProtoMessage()    {}
func (*BuildState) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{4}
}

func (m *BuildState) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BuildState.Unmarshal(m, b)
}
func (m *BuildState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BuildState.Marshal(b, m, deterministic)
}
func (m *BuildState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BuildState.Merge(m, src)
}
func (m *BuildState) XXX_Size() int {
	return xxx_messageInfo_BuildState.Size(m)
}
func (m *BuildState) XXX_DiscardUnknown() {
	xxx_messageInfo_BuildState.DiscardUnknown(m)
}

var xxx_messageInfo_BuildState proto.InternalMessageInfo

func (m *BuildState) GetArtifacts() map[string]string {
	if m != nil {
		return m.Artifacts
	}
	return nil
}

// DeployState contains the status of the current deploy
type DeployState struct {
	Status               string   `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeployState) Reset()         { *m = DeployState{} }
func (m *DeployState) String() string { return proto.CompactTextString(m) }
func (*DeployState) ProtoMessage()    {}
func (*DeployState) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{5}
}

func (m *DeployState) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeployState.Unmarshal(m, b)
}
func (m *DeployState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeployState.Marshal(b, m, deterministic)
}
func (m *DeployState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeployState.Merge(m, src)
}
func (m *DeployState) XXX_Size() int {
	return xxx_messageInfo_DeployState.Size(m)
}
func (m *DeployState) XXX_DiscardUnknown() {
	xxx_messageInfo_DeployState.DiscardUnknown(m)
}

var xxx_messageInfo_DeployState proto.InternalMessageInfo

func (m *DeployState) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

type Event struct {
	// Types that are valid to be assigned to EventType:
	//	*Event_MetaEvent
	//	*Event_BuildEvent
	//	*Event_DeployEvent
	//	*Event_PortEvent
	EventType            isEvent_EventType `protobuf_oneof:"event_type"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Event) Reset()         { *m = Event{} }
func (m *Event) String() string { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()    {}
func (*Event) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{6}
}

func (m *Event) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Event.Unmarshal(m, b)
}
func (m *Event) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Event.Marshal(b, m, deterministic)
}
func (m *Event) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Event.Merge(m, src)
}
func (m *Event) XXX_Size() int {
	return xxx_messageInfo_Event.Size(m)
}
func (m *Event) XXX_DiscardUnknown() {
	xxx_messageInfo_Event.DiscardUnknown(m)
}

var xxx_messageInfo_Event proto.InternalMessageInfo

type isEvent_EventType interface {
	isEvent_EventType()
}

type Event_MetaEvent struct {
	MetaEvent *MetaEvent `protobuf:"bytes,1,opt,name=metaEvent,proto3,oneof"`
}

type Event_BuildEvent struct {
	BuildEvent *BuildEvent `protobuf:"bytes,2,opt,name=buildEvent,proto3,oneof"`
}

type Event_DeployEvent struct {
	DeployEvent *DeployEvent `protobuf:"bytes,3,opt,name=deployEvent,proto3,oneof"`
}

type Event_PortEvent struct {
	PortEvent *PortEvent `protobuf:"bytes,4,opt,name=portEvent,proto3,oneof"`
}

func (*Event_MetaEvent) isEvent_EventType() {}

func (*Event_BuildEvent) isEvent_EventType() {}

func (*Event_DeployEvent) isEvent_EventType() {}

func (*Event_PortEvent) isEvent_EventType() {}

func (m *Event) GetEventType() isEvent_EventType {
	if m != nil {
		return m.EventType
	}
	return nil
}

func (m *Event) GetMetaEvent() *MetaEvent {
	if x, ok := m.GetEventType().(*Event_MetaEvent); ok {
		return x.MetaEvent
	}
	return nil
}

func (m *Event) GetBuildEvent() *BuildEvent {
	if x, ok := m.GetEventType().(*Event_BuildEvent); ok {
		return x.BuildEvent
	}
	return nil
}

func (m *Event) GetDeployEvent() *DeployEvent {
	if x, ok := m.GetEventType().(*Event_DeployEvent); ok {
		return x.DeployEvent
	}
	return nil
}

func (m *Event) GetPortEvent() *PortEvent {
	if x, ok := m.GetEventType().(*Event_PortEvent); ok {
		return x.PortEvent
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Event) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Event_MetaEvent)(nil),
		(*Event_BuildEvent)(nil),
		(*Event_DeployEvent)(nil),
		(*Event_PortEvent)(nil),
	}
}

type MetaEvent struct {
	Entry                string   `protobuf:"bytes,1,opt,name=entry,proto3" json:"entry,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MetaEvent) Reset()         { *m = MetaEvent{} }
func (m *MetaEvent) String() string { return proto.CompactTextString(m) }
func (*MetaEvent) ProtoMessage()    {}
func (*MetaEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{7}
}

func (m *MetaEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MetaEvent.Unmarshal(m, b)
}
func (m *MetaEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MetaEvent.Marshal(b, m, deterministic)
}
func (m *MetaEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MetaEvent.Merge(m, src)
}
func (m *MetaEvent) XXX_Size() int {
	return xxx_messageInfo_MetaEvent.Size(m)
}
func (m *MetaEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_MetaEvent.DiscardUnknown(m)
}

var xxx_messageInfo_MetaEvent proto.InternalMessageInfo

func (m *MetaEvent) GetEntry() string {
	if m != nil {
		return m.Entry
	}
	return ""
}

type BuildEvent struct {
	Artifact             string   `protobuf:"bytes,1,opt,name=artifact,proto3" json:"artifact,omitempty"`
	Status               string   `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Err                  string   `protobuf:"bytes,3,opt,name=err,proto3" json:"err,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BuildEvent) Reset()         { *m = BuildEvent{} }
func (m *BuildEvent) String() string { return proto.CompactTextString(m) }
func (*BuildEvent) ProtoMessage()    {}
func (*BuildEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{8}
}

func (m *BuildEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BuildEvent.Unmarshal(m, b)
}
func (m *BuildEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BuildEvent.Marshal(b, m, deterministic)
}
func (m *BuildEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BuildEvent.Merge(m, src)
}
func (m *BuildEvent) XXX_Size() int {
	return xxx_messageInfo_BuildEvent.Size(m)
}
func (m *BuildEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_BuildEvent.DiscardUnknown(m)
}

var xxx_messageInfo_BuildEvent proto.InternalMessageInfo

func (m *BuildEvent) GetArtifact() string {
	if m != nil {
		return m.Artifact
	}
	return ""
}

func (m *BuildEvent) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *BuildEvent) GetErr() string {
	if m != nil {
		return m.Err
	}
	return ""
}

type DeployEvent struct {
	Status               string   `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Err                  string   `protobuf:"bytes,2,opt,name=err,proto3" json:"err,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeployEvent) Reset()         { *m = DeployEvent{} }
func (m *DeployEvent) String() string { return proto.CompactTextString(m) }
func (*DeployEvent) ProtoMessage()    {}
func (*DeployEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{9}
}

func (m *DeployEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeployEvent.Unmarshal(m, b)
}
func (m *DeployEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeployEvent.Marshal(b, m, deterministic)
}
func (m *DeployEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeployEvent.Merge(m, src)
}
func (m *DeployEvent) XXX_Size() int {
	return xxx_messageInfo_DeployEvent.Size(m)
}
func (m *DeployEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_DeployEvent.DiscardUnknown(m)
}

var xxx_messageInfo_DeployEvent proto.InternalMessageInfo

func (m *DeployEvent) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *DeployEvent) GetErr() string {
	if m != nil {
		return m.Err
	}
	return ""
}

type PortEvent struct {
	LocalPort            int32    `protobuf:"varint,1,opt,name=localPort,proto3" json:"localPort,omitempty"`
	RemotePort           int32    `protobuf:"varint,2,opt,name=remotePort,proto3" json:"remotePort,omitempty"`
	PodName              string   `protobuf:"bytes,3,opt,name=podName,proto3" json:"podName,omitempty"`
	ContainerName        string   `protobuf:"bytes,4,opt,name=containerName,proto3" json:"containerName,omitempty"`
	Namespace            string   `protobuf:"bytes,5,opt,name=namespace,proto3" json:"namespace,omitempty"`
	PortName             string   `protobuf:"bytes,6,opt,name=portName,proto3" json:"portName,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PortEvent) Reset()         { *m = PortEvent{} }
func (m *PortEvent) String() string { return proto.CompactTextString(m) }
func (*PortEvent) ProtoMessage()    {}
func (*PortEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{10}
}

func (m *PortEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PortEvent.Unmarshal(m, b)
}
func (m *PortEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PortEvent.Marshal(b, m, deterministic)
}
func (m *PortEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PortEvent.Merge(m, src)
}
func (m *PortEvent) XXX_Size() int {
	return xxx_messageInfo_PortEvent.Size(m)
}
func (m *PortEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_PortEvent.DiscardUnknown(m)
}

var xxx_messageInfo_PortEvent proto.InternalMessageInfo

func (m *PortEvent) GetLocalPort() int32 {
	if m != nil {
		return m.LocalPort
	}
	return 0
}

func (m *PortEvent) GetRemotePort() int32 {
	if m != nil {
		return m.RemotePort
	}
	return 0
}

func (m *PortEvent) GetPodName() string {
	if m != nil {
		return m.PodName
	}
	return ""
}

func (m *PortEvent) GetContainerName() string {
	if m != nil {
		return m.ContainerName
	}
	return ""
}

func (m *PortEvent) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *PortEvent) GetPortName() string {
	if m != nil {
		return m.PortName
	}
	return ""
}

type LogEntry struct {
	Timestamp            *timestamp.Timestamp `protobuf:"bytes,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Event                *Event               `protobuf:"bytes,2,opt,name=event,proto3" json:"event,omitempty"`
	Entry                string               `protobuf:"bytes,3,opt,name=entry,proto3" json:"entry,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *LogEntry) Reset()         { *m = LogEntry{} }
func (m *LogEntry) String() string { return proto.CompactTextString(m) }
func (*LogEntry) ProtoMessage()    {}
func (*LogEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{11}
}

func (m *LogEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LogEntry.Unmarshal(m, b)
}
func (m *LogEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LogEntry.Marshal(b, m, deterministic)
}
func (m *LogEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LogEntry.Merge(m, src)
}
func (m *LogEntry) XXX_Size() int {
	return xxx_messageInfo_LogEntry.Size(m)
}
func (m *LogEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_LogEntry.DiscardUnknown(m)
}

var xxx_messageInfo_LogEntry proto.InternalMessageInfo

func (m *LogEntry) GetTimestamp() *timestamp.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *LogEntry) GetEvent() *Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (m *LogEntry) GetEntry() string {
	if m != nil {
		return m.Entry
	}
	return ""
}

type ApiResponse struct {
	Response             string   `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ApiResponse) Reset()         { *m = ApiResponse{} }
func (m *ApiResponse) String() string { return proto.CompactTextString(m) }
func (*ApiResponse) ProtoMessage()    {}
func (*ApiResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4f2d38e344f9dbf5, []int{12}
}

func (m *ApiResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApiResponse.Unmarshal(m, b)
}
func (m *ApiResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApiResponse.Marshal(b, m, deterministic)
}
func (m *ApiResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApiResponse.Merge(m, src)
}
func (m *ApiResponse) XXX_Size() int {
	return xxx_messageInfo_ApiResponse.Size(m)
}
func (m *ApiResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ApiResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ApiResponse proto.InternalMessageInfo

func (m *ApiResponse) GetResponse() string {
	if m != nil {
		return m.Response
	}
	return ""
}

func init() {
	proto.RegisterType((*StateResponse)(nil), "proto.StateResponse")
	proto.RegisterType((*Response)(nil), "proto.Response")
	proto.RegisterType((*Request)(nil), "proto.Request")
	proto.RegisterType((*State)(nil), "proto.State")
	proto.RegisterMapType((map[string]*PortEvent)(nil), "proto.State.ForwardedPortsEntry")
	proto.RegisterType((*BuildState)(nil), "proto.BuildState")
	proto.RegisterMapType((map[string]string)(nil), "proto.BuildState.ArtifactsEntry")
	proto.RegisterType((*DeployState)(nil), "proto.DeployState")
	proto.RegisterType((*Event)(nil), "proto.Event")
	proto.RegisterType((*MetaEvent)(nil), "proto.MetaEvent")
	proto.RegisterType((*BuildEvent)(nil), "proto.BuildEvent")
	proto.RegisterType((*DeployEvent)(nil), "proto.DeployEvent")
	proto.RegisterType((*PortEvent)(nil), "proto.PortEvent")
	proto.RegisterType((*LogEntry)(nil), "proto.LogEntry")
	proto.RegisterType((*ApiResponse)(nil), "proto.ApiResponse")
}

func init() { proto.RegisterFile("skaffold.proto", fileDescriptor_4f2d38e344f9dbf5) }

var fileDescriptor_4f2d38e344f9dbf5 = []byte{
	// 794 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xcd, 0x6e, 0x13, 0x49,
	0x10, 0xde, 0x19, 0x67, 0x1c, 0x4f, 0x39, 0x71, 0x92, 0xde, 0x6c, 0x64, 0xcd, 0x7a, 0x77, 0xb3,
	0x2d, 0x40, 0x21, 0x07, 0x3b, 0x3f, 0x08, 0xa2, 0x08, 0x21, 0x25, 0x24, 0x10, 0x44, 0x40, 0x68,
	0xcc, 0x1d, 0x75, 0xec, 0xb6, 0x19, 0x65, 0x3c, 0x3d, 0xcc, 0xb4, 0x8d, 0x7c, 0xe1, 0x90, 0x23,
	0x17, 0x0e, 0x3c, 0x12, 0x8f, 0xc0, 0x2b, 0x20, 0x9e, 0x03, 0xf5, 0xdf, 0x4c, 0x3b, 0x71, 0x0e,
	0x9c, 0xa6, 0xab, 0xea, 0xfb, 0x6a, 0xaa, 0xea, 0xeb, 0x2e, 0x68, 0xe4, 0x97, 0x64, 0x30, 0x60,
	0x71, 0xbf, 0x9d, 0x66, 0x8c, 0x33, 0xe4, 0xc9, 0x4f, 0xd0, 0x1a, 0x32, 0x36, 0x8c, 0x69, 0x87,
	0xa4, 0x51, 0x87, 0x24, 0x09, 0xe3, 0x84, 0x47, 0x2c, 0xc9, 0x15, 0x28, 0xf8, 0x4f, 0x47, 0xa5,
	0x75, 0x31, 0x1e, 0x74, 0x78, 0x34, 0xa2, 0x39, 0x27, 0xa3, 0x54, 0x03, 0xfe, 0xbe, 0x0e, 0xa0,
	0xa3, 0x94, 0x4f, 0x55, 0x10, 0xef, 0xc3, 0x72, 0x97, 0x13, 0x4e, 0x43, 0x9a, 0xa7, 0x2c, 0xc9,
	0x29, 0xc2, 0xe0, 0xe5, 0xc2, 0xd1, 0x74, 0x36, 0x9d, 0xad, 0xfa, 0xde, 0x92, 0xc2, 0xb5, 0x15,
	0x48, 0x85, 0x70, 0x0b, 0x6a, 0x05, 0x7e, 0x15, 0x2a, 0xa3, 0x7c, 0x28, 0xd1, 0x7e, 0x28, 0x8e,
	0xf8, 0x1f, 0x58, 0x0c, 0xe9, 0x87, 0x31, 0xcd, 0x39, 0x42, 0xb0, 0x90, 0x90, 0x11, 0xd5, 0x51,
	0x79, 0xc6, 0x5f, 0x5c, 0xf0, 0x64, 0x36, 0xb4, 0x0b, 0x70, 0x31, 0x8e, 0xe2, 0x7e, 0xd7, 0xfa,
	0xdf, 0x9a, 0xfe, 0xdf, 0x71, 0x11, 0x08, 0x2d, 0x10, 0x7a, 0x00, 0xf5, 0x3e, 0x4d, 0x63, 0x36,
	0x55, 0x1c, 0x57, 0x72, 0x90, 0xe6, 0x9c, 0x94, 0x91, 0xd0, 0x86, 0xa1, 0x33, 0x68, 0x0c, 0x58,
	0xf6, 0x91, 0x64, 0x7d, 0xda, 0x7f, 0xc3, 0x32, 0x9e, 0x37, 0x2b, 0x9b, 0x95, 0xad, 0xfa, 0xde,
	0xa6, 0xdd, 0x5c, 0xfb, 0xd9, 0x0c, 0xe4, 0x34, 0xe1, 0xd9, 0x34, 0xbc, 0xc6, 0x0b, 0xba, 0xf0,
	0xe7, 0x1c, 0x98, 0x18, 0xc2, 0x25, 0x9d, 0x9a, 0x21, 0x5c, 0xd2, 0x29, 0xba, 0x07, 0xde, 0x84,
	0xc4, 0x63, 0x53, 0xe2, 0xaa, 0xfe, 0x93, 0xe0, 0x9c, 0x4e, 0x68, 0xc2, 0x43, 0x15, 0x3e, 0x74,
	0x0f, 0x1c, 0xfc, 0xd9, 0x01, 0x28, 0xfb, 0x45, 0x4f, 0xc0, 0x27, 0x19, 0x8f, 0x06, 0xa4, 0xc7,
	0xf3, 0xa6, 0x33, 0x53, 0x68, 0x89, 0x6a, 0x1f, 0x19, 0x88, 0x2a, 0xb4, 0xa4, 0x04, 0x8f, 0xa1,
	0x31, 0x1b, 0x9c, 0x53, 0xde, 0xba, 0x5d, 0x9e, 0x6f, 0x17, 0x73, 0x17, 0xea, 0xd6, 0x1c, 0xd1,
	0x06, 0x54, 0x85, 0xe6, 0xe3, 0x5c, 0xb3, 0xb5, 0x85, 0x7f, 0x3a, 0xe0, 0xc9, 0x46, 0xd0, 0x0e,
	0xf8, 0x23, 0xca, 0x89, 0x34, 0xb4, 0x88, 0xa6, 0xdb, 0x57, 0xc6, 0x7f, 0xf6, 0x47, 0x58, 0x82,
	0xd0, 0xbe, 0xd6, 0x5d, 0x51, 0xdc, 0x9b, 0xba, 0x1b, 0x8e, 0x05, 0x43, 0x0f, 0x8d, 0xf2, 0x8a,
	0x55, 0x99, 0xa3, 0xbc, 0xa1, 0xd9, 0x40, 0x51, 0x5e, 0x6a, 0x86, 0xde, 0x5c, 0x98, 0x2f, 0x86,
	0x28, 0xaf, 0x00, 0x1d, 0x2f, 0x01, 0x50, 0x71, 0x78, 0xc7, 0xa7, 0x29, 0xc5, 0xff, 0x83, 0x5f,
	0xb4, 0x21, 0xc6, 0x46, 0xc5, 0x44, 0xf5, 0x30, 0x94, 0x81, 0x43, 0x2d, 0x9f, 0xc2, 0x04, 0x50,
	0x33, 0x5a, 0x68, 0x58, 0x61, 0x5b, 0xd3, 0x74, 0xed, 0x69, 0x0a, 0x81, 0x68, 0x96, 0xc9, 0xa6,
	0xfc, 0x50, 0x1c, 0xf1, 0x23, 0x23, 0x83, 0x4a, 0x7a, 0x8b, 0x0c, 0x86, 0xe8, 0x96, 0xc4, 0x6f,
	0x0e, 0xf8, 0x45, 0x63, 0xa8, 0x05, 0x7e, 0xcc, 0x7a, 0x24, 0x16, 0x1e, 0x49, 0xf5, 0xc2, 0xd2,
	0x81, 0xfe, 0x05, 0xc8, 0xe8, 0x88, 0x71, 0x2a, 0xc3, 0xae, 0x0c, 0x5b, 0x1e, 0xd4, 0x84, 0xc5,
	0x94, 0xf5, 0x5f, 0x8b, 0x17, 0xac, 0x4a, 0x33, 0x26, 0xba, 0x03, 0xcb, 0x3d, 0x96, 0x70, 0x12,
	0x25, 0x34, 0x93, 0xf1, 0x05, 0x19, 0x9f, 0x75, 0x8a, 0xbf, 0x8b, 0x27, 0x9f, 0xa7, 0xa4, 0x47,
	0x9b, 0x9e, 0x44, 0x94, 0x0e, 0x31, 0x28, 0x31, 0x74, 0x49, 0xaf, 0xaa, 0x41, 0x19, 0x1b, 0x7f,
	0x82, 0xda, 0x39, 0x1b, 0xaa, 0xdb, 0x7b, 0x00, 0x7e, 0xb1, 0xd2, 0xf4, 0x05, 0x0b, 0xda, 0x6a,
	0xa7, 0xb5, 0xcd, 0x4e, 0x6b, 0xbf, 0x35, 0x88, 0xb0, 0x04, 0x8b, 0x5d, 0x46, 0xad, 0x3b, 0x66,
	0x76, 0x99, 0x7e, 0x80, 0x74, 0x56, 0xd2, 0x8a, 0x2d, 0xe9, 0x7d, 0xa8, 0x1f, 0xa5, 0x51, 0xb1,
	0xe4, 0x02, 0xa8, 0x65, 0xfa, 0x6c, 0x34, 0x35, 0xf6, 0xde, 0x55, 0x05, 0x56, 0xba, 0x7a, 0x6f,
	0x77, 0x69, 0x36, 0x89, 0x7a, 0x14, 0x3d, 0x85, 0xda, 0x73, 0xca, 0xf5, 0x0b, 0xba, 0x51, 0xeb,
	0xa9, 0xd8, 0xbf, 0xc1, 0xcc, 0x66, 0xc5, 0x6b, 0x57, 0xdf, 0x7f, 0x7c, 0x75, 0xeb, 0xc8, 0xef,
	0x4c, 0x76, 0x3b, 0x72, 0xcb, 0xa2, 0x13, 0xa8, 0xc9, 0x4a, 0xcf, 0xd9, 0x10, 0xad, 0x68, 0xb0,
	0x19, 0x4a, 0x70, 0xdd, 0x81, 0x91, 0x4c, 0xb0, 0x84, 0x40, 0x24, 0x90, 0xad, 0xe5, 0x5b, 0xce,
	0x8e, 0x83, 0xce, 0xa1, 0x7a, 0x46, 0x92, 0x7e, 0x4c, 0xd1, 0x4c, 0xfb, 0xc1, 0x2d, 0x65, 0xe1,
	0x96, 0xcc, 0xb3, 0x81, 0xd7, 0xca, 0x3c, 0x9d, 0xf7, 0x32, 0xc1, 0xa1, 0xb3, 0x8d, 0x5e, 0x80,
	0x27, 0xaf, 0xfa, 0xad, 0x5d, 0x99, 0x17, 0x69, 0x4d, 0x0f, 0xaf, 0xcb, 0x94, 0x0d, 0x2c, 0x7b,
	0x93, 0x4f, 0x5a, 0xa4, 0x7a, 0x09, 0x55, 0x75, 0xc3, 0x7f, 0x2b, 0xd7, 0x5f, 0x32, 0xd7, 0x0a,
	0x96, 0x6d, 0xaa, 0x77, 0x7e, 0xe8, 0x6c, 0x5f, 0x54, 0x25, 0x72, 0xff, 0x57, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xb1, 0x29, 0xf5, 0xe0, 0x42, 0x07, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SkaffoldServiceClient is the client API for SkaffoldService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SkaffoldServiceClient interface {
	GetState(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*State, error)
	EventLog(ctx context.Context, opts ...grpc.CallOption) (SkaffoldService_EventLogClient, error)
	Handle(ctx context.Context, in *Event, opts ...grpc.CallOption) (*empty.Empty, error)
	Build(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ApiResponse, error)
	Deploy(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ApiResponse, error)
}

type skaffoldServiceClient struct {
	cc *grpc.ClientConn
}

func NewSkaffoldServiceClient(cc *grpc.ClientConn) SkaffoldServiceClient {
	return &skaffoldServiceClient{cc}
}

func (c *skaffoldServiceClient) GetState(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*State, error) {
	out := new(State)
	err := c.cc.Invoke(ctx, "/proto.SkaffoldService/GetState", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *skaffoldServiceClient) EventLog(ctx context.Context, opts ...grpc.CallOption) (SkaffoldService_EventLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &_SkaffoldService_serviceDesc.Streams[0], "/proto.SkaffoldService/EventLog", opts...)
	if err != nil {
		return nil, err
	}
	x := &skaffoldServiceEventLogClient{stream}
	return x, nil
}

type SkaffoldService_EventLogClient interface {
	Send(*LogEntry) error
	Recv() (*LogEntry, error)
	grpc.ClientStream
}

type skaffoldServiceEventLogClient struct {
	grpc.ClientStream
}

func (x *skaffoldServiceEventLogClient) Send(m *LogEntry) error {
	return x.ClientStream.SendMsg(m)
}

func (x *skaffoldServiceEventLogClient) Recv() (*LogEntry, error) {
	m := new(LogEntry)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *skaffoldServiceClient) Handle(ctx context.Context, in *Event, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/proto.SkaffoldService/Handle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *skaffoldServiceClient) Build(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ApiResponse, error) {
	out := new(ApiResponse)
	err := c.cc.Invoke(ctx, "/proto.SkaffoldService/Build", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *skaffoldServiceClient) Deploy(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ApiResponse, error) {
	out := new(ApiResponse)
	err := c.cc.Invoke(ctx, "/proto.SkaffoldService/Deploy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SkaffoldServiceServer is the server API for SkaffoldService service.
type SkaffoldServiceServer interface {
	GetState(context.Context, *empty.Empty) (*State, error)
	EventLog(SkaffoldService_EventLogServer) error
	Handle(context.Context, *Event) (*empty.Empty, error)
	Build(context.Context, *empty.Empty) (*ApiResponse, error)
	Deploy(context.Context, *empty.Empty) (*ApiResponse, error)
}

// UnimplementedSkaffoldServiceServer can be embedded to have forward compatible implementations.
type UnimplementedSkaffoldServiceServer struct {
}

func (*UnimplementedSkaffoldServiceServer) GetState(ctx context.Context, req *empty.Empty) (*State, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetState not implemented")
}
func (*UnimplementedSkaffoldServiceServer) EventLog(srv SkaffoldService_EventLogServer) error {
	return status.Errorf(codes.Unimplemented, "method EventLog not implemented")
}
func (*UnimplementedSkaffoldServiceServer) Handle(ctx context.Context, req *Event) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Handle not implemented")
}
func (*UnimplementedSkaffoldServiceServer) Build(ctx context.Context, req *empty.Empty) (*ApiResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Build not implemented")
}
func (*UnimplementedSkaffoldServiceServer) Deploy(ctx context.Context, req *empty.Empty) (*ApiResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Deploy not implemented")
}

func RegisterSkaffoldServiceServer(s *grpc.Server, srv SkaffoldServiceServer) {
	s.RegisterService(&_SkaffoldService_serviceDesc, srv)
}

func _SkaffoldService_GetState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkaffoldServiceServer).GetState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SkaffoldService/GetState",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkaffoldServiceServer).GetState(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _SkaffoldService_EventLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SkaffoldServiceServer).EventLog(&skaffoldServiceEventLogServer{stream})
}

type SkaffoldService_EventLogServer interface {
	Send(*LogEntry) error
	Recv() (*LogEntry, error)
	grpc.ServerStream
}

type skaffoldServiceEventLogServer struct {
	grpc.ServerStream
}

func (x *skaffoldServiceEventLogServer) Send(m *LogEntry) error {
	return x.ServerStream.SendMsg(m)
}

func (x *skaffoldServiceEventLogServer) Recv() (*LogEntry, error) {
	m := new(LogEntry)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _SkaffoldService_Handle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Event)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkaffoldServiceServer).Handle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SkaffoldService/Handle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkaffoldServiceServer).Handle(ctx, req.(*Event))
	}
	return interceptor(ctx, in, info, handler)
}

func _SkaffoldService_Build_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkaffoldServiceServer).Build(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SkaffoldService/Build",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkaffoldServiceServer).Build(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _SkaffoldService_Deploy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkaffoldServiceServer).Deploy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SkaffoldService/Deploy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkaffoldServiceServer).Deploy(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _SkaffoldService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.SkaffoldService",
	HandlerType: (*SkaffoldServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetState",
			Handler:    _SkaffoldService_GetState_Handler,
		},
		{
			MethodName: "Handle",
			Handler:    _SkaffoldService_Handle_Handler,
		},
		{
			MethodName: "Build",
			Handler:    _SkaffoldService_Build_Handler,
		},
		{
			MethodName: "Deploy",
			Handler:    _SkaffoldService_Deploy_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EventLog",
			Handler:       _SkaffoldService_EventLog_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "skaffold.proto",
}
