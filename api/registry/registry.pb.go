// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.13.0
// source: api/registry/registry.proto

package registry

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type NamedPort struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The actual numerical listening port
	Port int32 `protobuf:"varint,1,opt,name=port,proto3" json:"port,omitempty"`
	// The name of the port, e.g. "https"
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *NamedPort) Reset() {
	*x = NamedPort{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_registry_registry_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NamedPort) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NamedPort) ProtoMessage() {}

func (x *NamedPort) ProtoReflect() protoreflect.Message {
	mi := &file_api_registry_registry_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NamedPort.ProtoReflect.Descriptor instead.
func (*NamedPort) Descriptor() ([]byte, []int) {
	return file_api_registry_registry_proto_rawDescGZIP(), []int{0}
}

func (x *NamedPort) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *NamedPort) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// The request message with the service info.
type ServiceInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// namespace the service is registered in
	Namespace string `protobuf:"bytes,1,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// name of the service - this determines which namespace this service uses
	ServiceName string `protobuf:"bytes,2,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	// DNS name of the service host not including subdomain or domain and no dots
	HostName string `protobuf:"bytes,3,opt,name=host_name,json=hostName,proto3" json:"host_name,omitempty"`
	// IP address of that specific host name
	Ipaddress string `protobuf:"bytes,4,opt,name=ipaddress,proto3" json:"ipaddress,omitempty"`
	// Physical node name where the service instance is running.
	// For bare metal or VM services, this may the the FQN of the service including the host_name
	NodeName string `protobuf:"bytes,5,opt,name=node_name,json=nodeName,proto3" json:"node_name,omitempty"`
	// Array of named ports for the service
	Ports []*NamedPort `protobuf:"bytes,6,rep,name=ports,proto3" json:"ports,omitempty"`
	// A weight this service instance should be given
	// NOTE: this information may not be utilized by the consumer of this information
	Weight float32 `protobuf:"fixed32,7,opt,name=weight,proto3" json:"weight,omitempty"`
}

func (x *ServiceInfo) Reset() {
	*x = ServiceInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_registry_registry_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceInfo) ProtoMessage() {}

func (x *ServiceInfo) ProtoReflect() protoreflect.Message {
	mi := &file_api_registry_registry_proto_msgTypes[1]
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
	return file_api_registry_registry_proto_rawDescGZIP(), []int{1}
}

func (x *ServiceInfo) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *ServiceInfo) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *ServiceInfo) GetHostName() string {
	if x != nil {
		return x.HostName
	}
	return ""
}

func (x *ServiceInfo) GetIpaddress() string {
	if x != nil {
		return x.Ipaddress
	}
	return ""
}

func (x *ServiceInfo) GetNodeName() string {
	if x != nil {
		return x.NodeName
	}
	return ""
}

func (x *ServiceInfo) GetPorts() []*NamedPort {
	if x != nil {
		return x.Ports
	}
	return nil
}

func (x *ServiceInfo) GetWeight() float32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

// The response message containing the result of the registration
// In case of a successful registration, the input will be in the response.
// If the registration failed, all values except for the status code and a reason
// will be empty.
type RegistrationResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// namespace the service is registered in
	Namespace string `protobuf:"bytes,1,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// name of the service
	ServiceName string `protobuf:"bytes,2,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	// IP address of that specific host name
	Ipaddress string `protobuf:"bytes,3,opt,name=ipaddress,proto3" json:"ipaddress,omitempty"`
	// Array of named ports for the service
	Ports []*NamedPort `protobuf:"bytes,4,rep,name=ports,proto3" json:"ports,omitempty"`
	// Status code of the operation
	// We will reuse the HTTP status code, e.g. a 201 will signal that the entry
	// has been successfully created.
	Status uint32 `protobuf:"varint,5,opt,name=status,proto3" json:"status,omitempty"`
	// Details message explaining the status
	StatusDetails string `protobuf:"bytes,6,opt,name=status_details,json=statusDetails,proto3" json:"status_details,omitempty"`
}

func (x *RegistrationResult) Reset() {
	*x = RegistrationResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_registry_registry_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegistrationResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegistrationResult) ProtoMessage() {}

func (x *RegistrationResult) ProtoReflect() protoreflect.Message {
	mi := &file_api_registry_registry_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegistrationResult.ProtoReflect.Descriptor instead.
func (*RegistrationResult) Descriptor() ([]byte, []int) {
	return file_api_registry_registry_proto_rawDescGZIP(), []int{2}
}

func (x *RegistrationResult) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *RegistrationResult) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *RegistrationResult) GetIpaddress() string {
	if x != nil {
		return x.Ipaddress
	}
	return ""
}

func (x *RegistrationResult) GetPorts() []*NamedPort {
	if x != nil {
		return x.Ports
	}
	return nil
}

func (x *RegistrationResult) GetStatus() uint32 {
	if x != nil {
		return x.Status
	}
	return 0
}

func (x *RegistrationResult) GetStatusDetails() string {
	if x != nil {
		return x.StatusDetails
	}
	return ""
}

var File_api_registry_registry_proto protoreflect.FileDescriptor

var file_api_registry_registry_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x22, 0x33, 0x0a, 0x09, 0x4e, 0x61, 0x6d, 0x65, 0x64,
	0x50, 0x6f, 0x72, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0xe9, 0x01, 0x0a,
	0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1c, 0x0a, 0x09,
	0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1b, 0x0a,
	0x09, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x70,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x69,
	0x70, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x6e, 0x6f, 0x64, 0x65,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6e, 0x6f, 0x64,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x29, 0x0a, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x06,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x4e, 0x61, 0x6d, 0x65, 0x64, 0x50, 0x6f, 0x72, 0x74, 0x52, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73,
	0x12, 0x16, 0x0a, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x02,
	0x52, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x22, 0xdd, 0x01, 0x0a, 0x12, 0x52, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12,
	0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x21, 0x0a,
	0x0c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x1c, 0x0a, 0x09, 0x69, 0x70, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x70, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x29,
	0x0a, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x4e, 0x61, 0x6d, 0x65, 0x64, 0x50, 0x6f,
	0x72, 0x74, 0x52, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x25, 0x0a, 0x0e, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x64, 0x65, 0x74, 0x61,
	0x69, 0x6c, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x32, 0x99, 0x01, 0x0a, 0x0f, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x12, 0x41, 0x0a, 0x08,
	0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x15, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x1a,
	0x1c, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x00, 0x12,
	0x43, 0x0a, 0x0a, 0x55, 0x6e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x15, 0x2e,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x1c, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x22, 0x00, 0x42, 0x47, 0x0a, 0x23, 0x63, 0x6f, 0x6d, 0x2e, 0x62, 0x6f, 0x78, 0x2e,
	0x6b, 0x38, 0x73, 0x73, 0x76, 0x63, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x42, 0x10, 0x53, 0x76, 0x63,
	0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x0c, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_registry_registry_proto_rawDescOnce sync.Once
	file_api_registry_registry_proto_rawDescData = file_api_registry_registry_proto_rawDesc
)

func file_api_registry_registry_proto_rawDescGZIP() []byte {
	file_api_registry_registry_proto_rawDescOnce.Do(func() {
		file_api_registry_registry_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_registry_registry_proto_rawDescData)
	})
	return file_api_registry_registry_proto_rawDescData
}

var file_api_registry_registry_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_api_registry_registry_proto_goTypes = []interface{}{
	(*NamedPort)(nil),          // 0: registry.NamedPort
	(*ServiceInfo)(nil),        // 1: registry.ServiceInfo
	(*RegistrationResult)(nil), // 2: registry.RegistrationResult
}
var file_api_registry_registry_proto_depIdxs = []int32{
	0, // 0: registry.ServiceInfo.ports:type_name -> registry.NamedPort
	0, // 1: registry.RegistrationResult.ports:type_name -> registry.NamedPort
	1, // 2: registry.ServiceRegistry.Register:input_type -> registry.ServiceInfo
	1, // 3: registry.ServiceRegistry.UnRegister:input_type -> registry.ServiceInfo
	2, // 4: registry.ServiceRegistry.Register:output_type -> registry.RegistrationResult
	2, // 5: registry.ServiceRegistry.UnRegister:output_type -> registry.RegistrationResult
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_registry_registry_proto_init() }
func file_api_registry_registry_proto_init() {
	if File_api_registry_registry_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_registry_registry_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NamedPort); i {
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
		file_api_registry_registry_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_api_registry_registry_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegistrationResult); i {
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
			RawDescriptor: file_api_registry_registry_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_registry_registry_proto_goTypes,
		DependencyIndexes: file_api_registry_registry_proto_depIdxs,
		MessageInfos:      file_api_registry_registry_proto_msgTypes,
	}.Build()
	File_api_registry_registry_proto = out.File
	file_api_registry_registry_proto_rawDesc = nil
	file_api_registry_registry_proto_goTypes = nil
	file_api_registry_registry_proto_depIdxs = nil
}
