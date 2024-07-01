// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: userstore.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RemoteMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id            string           `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	OperationType string           `protobuf:"bytes,2,opt,name=operationType,proto3" json:"operationType,omitempty"`
	Organization  string           `protobuf:"bytes,3,opt,name=organization,proto3" json:"organization,omitempty"`
	Data          *structpb.Struct `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *RemoteMessage) Reset() {
	*x = RemoteMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_userstore_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoteMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoteMessage) ProtoMessage() {}

func (x *RemoteMessage) ProtoReflect() protoreflect.Message {
	mi := &file_userstore_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoteMessage.ProtoReflect.Descriptor instead.
func (*RemoteMessage) Descriptor() ([]byte, []int) {
	return file_userstore_proto_rawDescGZIP(), []int{0}
}

func (x *RemoteMessage) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RemoteMessage) GetOperationType() string {
	if x != nil {
		return x.OperationType
	}
	return ""
}

func (x *RemoteMessage) GetOrganization() string {
	if x != nil {
		return x.Organization
	}
	return ""
}

func (x *RemoteMessage) GetData() *structpb.Struct {
	if x != nil {
		return x.Data
	}
	return nil
}

type UserStoreRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OperationType string           `protobuf:"bytes,1,opt,name=operationType,proto3" json:"operationType,omitempty"`
	Organization  string           `protobuf:"bytes,2,opt,name=organization,proto3" json:"organization,omitempty"`
	Data          *structpb.Struct `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *UserStoreRequest) Reset() {
	*x = UserStoreRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_userstore_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserStoreRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserStoreRequest) ProtoMessage() {}

func (x *UserStoreRequest) ProtoReflect() protoreflect.Message {
	mi := &file_userstore_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserStoreRequest.ProtoReflect.Descriptor instead.
func (*UserStoreRequest) Descriptor() ([]byte, []int) {
	return file_userstore_proto_rawDescGZIP(), []int{1}
}

func (x *UserStoreRequest) GetOperationType() string {
	if x != nil {
		return x.OperationType
	}
	return ""
}

func (x *UserStoreRequest) GetOrganization() string {
	if x != nil {
		return x.Organization
	}
	return ""
}

func (x *UserStoreRequest) GetData() *structpb.Struct {
	if x != nil {
		return x.Data
	}
	return nil
}

type UserStoreResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OperationType string           `protobuf:"bytes,1,opt,name=operationType,proto3" json:"operationType,omitempty"`
	Organization  string           `protobuf:"bytes,2,opt,name=organization,proto3" json:"organization,omitempty"`
	Data          *structpb.Struct `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *UserStoreResponse) Reset() {
	*x = UserStoreResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_userstore_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserStoreResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserStoreResponse) ProtoMessage() {}

func (x *UserStoreResponse) ProtoReflect() protoreflect.Message {
	mi := &file_userstore_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserStoreResponse.ProtoReflect.Descriptor instead.
func (*UserStoreResponse) Descriptor() ([]byte, []int) {
	return file_userstore_proto_rawDescGZIP(), []int{2}
}

func (x *UserStoreResponse) GetOperationType() string {
	if x != nil {
		return x.OperationType
	}
	return ""
}

func (x *UserStoreResponse) GetOrganization() string {
	if x != nil {
		return x.Organization
	}
	return ""
}

func (x *UserStoreResponse) GetData() *structpb.Struct {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_userstore_proto protoreflect.FileDescriptor

var file_userstore_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x75, 0x73, 0x65, 0x72, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x96, 0x01, 0x0a, 0x0d, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x24, 0x0a, 0x0d, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x6f, 0x72, 0x67, 0x61, 0x6e,
	0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f,
	0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2b, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x89, 0x01, 0x0a, 0x10, 0x55, 0x73, 0x65,
	0x72, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x24, 0x0a,
	0x0d, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x6f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f, 0x72, 0x67, 0x61, 0x6e,
	0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2b, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x22, 0x8a, 0x01, 0x0a, 0x11, 0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x6f,
	0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x6f, 0x70,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x22, 0x0a, 0x0c, 0x6f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2b, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x32, 0x44, 0x0a, 0x0f, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x55, 0x73, 0x65, 0x72, 0x53,
	0x74, 0x6f, 0x72, 0x65, 0x12, 0x31, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x75, 0x6e, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x12, 0x0e, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x1a, 0x0e, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x28, 0x01, 0x30, 0x01, 0x32, 0x48, 0x0a, 0x0c, 0x52, 0x65, 0x6d, 0x6f, 0x74,
	0x65, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x38, 0x0a, 0x0f, 0x69, 0x6e, 0x76, 0x6f, 0x6b,
	0x65, 0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x11, 0x2e, 0x55, 0x73, 0x65,
	0x72, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e,
	0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x08, 0x5a, 0x06, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_userstore_proto_rawDescOnce sync.Once
	file_userstore_proto_rawDescData = file_userstore_proto_rawDesc
)

func file_userstore_proto_rawDescGZIP() []byte {
	file_userstore_proto_rawDescOnce.Do(func() {
		file_userstore_proto_rawDescData = protoimpl.X.CompressGZIP(file_userstore_proto_rawDescData)
	})
	return file_userstore_proto_rawDescData
}

var file_userstore_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_userstore_proto_goTypes = []any{
	(*RemoteMessage)(nil),     // 0: RemoteMessage
	(*UserStoreRequest)(nil),  // 1: UserStoreRequest
	(*UserStoreResponse)(nil), // 2: UserStoreResponse
	(*structpb.Struct)(nil),   // 3: google.protobuf.Struct
}
var file_userstore_proto_depIdxs = []int32{
	3, // 0: RemoteMessage.data:type_name -> google.protobuf.Struct
	3, // 1: UserStoreRequest.data:type_name -> google.protobuf.Struct
	3, // 2: UserStoreResponse.data:type_name -> google.protobuf.Struct
	0, // 3: RemoteUserStore.communicate:input_type -> RemoteMessage
	1, // 4: RemoteServer.invokeUserStore:input_type -> UserStoreRequest
	0, // 5: RemoteUserStore.communicate:output_type -> RemoteMessage
	2, // 6: RemoteServer.invokeUserStore:output_type -> UserStoreResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_userstore_proto_init() }
func file_userstore_proto_init() {
	if File_userstore_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_userstore_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*RemoteMessage); i {
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
		file_userstore_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*UserStoreRequest); i {
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
		file_userstore_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*UserStoreResponse); i {
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
			RawDescriptor: file_userstore_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_userstore_proto_goTypes,
		DependencyIndexes: file_userstore_proto_depIdxs,
		MessageInfos:      file_userstore_proto_msgTypes,
	}.Build()
	File_userstore_proto = out.File
	file_userstore_proto_rawDesc = nil
	file_userstore_proto_goTypes = nil
	file_userstore_proto_depIdxs = nil
}