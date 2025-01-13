// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.29.3
// source: messages.proto

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

type Flag int32

const (
	Flag_NONE       Flag = 0
	Flag_COMMAND    Flag = 1
	Flag_MSG_STDIN  Flag = 2
	Flag_MSG_STDOUT Flag = 3
	Flag_MSG_STDERR Flag = 4
	Flag_EOF_STDIN  Flag = 5
	Flag_EOF_STDOUT Flag = 6
	Flag_EOF_STDERR Flag = 7
)

// Enum value maps for Flag.
var (
	Flag_name = map[int32]string{
		0: "NONE",
		1: "COMMAND",
		2: "MSG_STDIN",
		3: "MSG_STDOUT",
		4: "MSG_STDERR",
		5: "EOF_STDIN",
		6: "EOF_STDOUT",
		7: "EOF_STDERR",
	}
	Flag_value = map[string]int32{
		"NONE":       0,
		"COMMAND":    1,
		"MSG_STDIN":  2,
		"MSG_STDOUT": 3,
		"MSG_STDERR": 4,
		"EOF_STDIN":  5,
		"EOF_STDOUT": 6,
		"EOF_STDERR": 7,
	}
)

func (x Flag) Enum() *Flag {
	p := new(Flag)
	*p = x
	return p
}

func (x Flag) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Flag) Descriptor() protoreflect.EnumDescriptor {
	return file_messages_proto_enumTypes[0].Descriptor()
}

func (Flag) Type() protoreflect.EnumType {
	return &file_messages_proto_enumTypes[0]
}

func (x Flag) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Flag.Descriptor instead.
func (Flag) EnumDescriptor() ([]byte, []int) {
	return file_messages_proto_rawDescGZIP(), []int{0}
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To   string `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Flag Flag   `protobuf:"varint,3,opt,name=flag,proto3,enum=grpcsh.Flag" json:"flag,omitempty"`
	Data []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_messages_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_messages_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_messages_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *Message) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *Message) GetFlag() Flag {
	if x != nil {
		return x.Flag
	}
	return Flag_NONE
}

func (x *Message) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Result struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To   string `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Flag Flag   `protobuf:"varint,3,opt,name=flag,proto3,enum=grpcsh.Flag" json:"flag,omitempty"`
	Data []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Result) Reset() {
	*x = Result{}
	if protoimpl.UnsafeEnabled {
		mi := &file_messages_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Result) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Result) ProtoMessage() {}

func (x *Result) ProtoReflect() protoreflect.Message {
	mi := &file_messages_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Result.ProtoReflect.Descriptor instead.
func (*Result) Descriptor() ([]byte, []int) {
	return file_messages_proto_rawDescGZIP(), []int{1}
}

func (x *Result) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *Result) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *Result) GetFlag() Flag {
	if x != nil {
		return x.Flag
	}
	return Flag_NONE
}

func (x *Result) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type PeerMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Channel string `protobuf:"bytes,1,opt,name=channel,proto3" json:"channel,omitempty"`
	From    string `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	To      string `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	Flag    Flag   `protobuf:"varint,4,opt,name=flag,proto3,enum=grpcsh.Flag" json:"flag,omitempty"`
	Data    []byte `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *PeerMessage) Reset() {
	*x = PeerMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_messages_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerMessage) ProtoMessage() {}

func (x *PeerMessage) ProtoReflect() protoreflect.Message {
	mi := &file_messages_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerMessage.ProtoReflect.Descriptor instead.
func (*PeerMessage) Descriptor() ([]byte, []int) {
	return file_messages_proto_rawDescGZIP(), []int{2}
}

func (x *PeerMessage) GetChannel() string {
	if x != nil {
		return x.Channel
	}
	return ""
}

func (x *PeerMessage) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *PeerMessage) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *PeerMessage) GetFlag() Flag {
	if x != nil {
		return x.Flag
	}
	return Flag_NONE
}

func (x *PeerMessage) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_messages_proto protoreflect.FileDescriptor

var file_messages_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x06, 0x67, 0x72, 0x70, 0x63, 0x73, 0x68, 0x22, 0x63, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12, 0x20, 0x0a, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x73, 0x68, 0x2e, 0x46,
	0x6c, 0x61, 0x67, 0x52, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x62, 0x0a,
	0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74,
	0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12, 0x20, 0x0a, 0x04, 0x66,
	0x6c, 0x61, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x73, 0x68, 0x2e, 0x46, 0x6c, 0x61, 0x67, 0x52, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x12, 0x12, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x22, 0x81, 0x01, 0x0a, 0x0b, 0x50, 0x65, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x66,
	0x72, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12,
	0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12,
	0x20, 0x0a, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e,
	0x67, 0x72, 0x70, 0x63, 0x73, 0x68, 0x2e, 0x46, 0x6c, 0x61, 0x67, 0x52, 0x04, 0x66, 0x6c, 0x61,
	0x67, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x2a, 0x7b, 0x0a, 0x04, 0x46, 0x6c, 0x61, 0x67, 0x12, 0x08, 0x0a,
	0x04, 0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x4d, 0x4d, 0x41,
	0x4e, 0x44, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x4d, 0x53, 0x47, 0x5f, 0x53, 0x54, 0x44, 0x49,
	0x4e, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x4d, 0x53, 0x47, 0x5f, 0x53, 0x54, 0x44, 0x4f, 0x55,
	0x54, 0x10, 0x03, 0x12, 0x0e, 0x0a, 0x0a, 0x4d, 0x53, 0x47, 0x5f, 0x53, 0x54, 0x44, 0x45, 0x52,
	0x52, 0x10, 0x04, 0x12, 0x0d, 0x0a, 0x09, 0x45, 0x4f, 0x46, 0x5f, 0x53, 0x54, 0x44, 0x49, 0x4e,
	0x10, 0x05, 0x12, 0x0e, 0x0a, 0x0a, 0x45, 0x4f, 0x46, 0x5f, 0x53, 0x54, 0x44, 0x4f, 0x55, 0x54,
	0x10, 0x06, 0x12, 0x0e, 0x0a, 0x0a, 0x45, 0x4f, 0x46, 0x5f, 0x53, 0x54, 0x44, 0x45, 0x52, 0x52,
	0x10, 0x07, 0x42, 0x0b, 0x5a, 0x09, 0x67, 0x72, 0x70, 0x63, 0x73, 0x68, 0x2f, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_messages_proto_rawDescOnce sync.Once
	file_messages_proto_rawDescData = file_messages_proto_rawDesc
)

func file_messages_proto_rawDescGZIP() []byte {
	file_messages_proto_rawDescOnce.Do(func() {
		file_messages_proto_rawDescData = protoimpl.X.CompressGZIP(file_messages_proto_rawDescData)
	})
	return file_messages_proto_rawDescData
}

var file_messages_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_messages_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_messages_proto_goTypes = []any{
	(Flag)(0),           // 0: grpcsh.Flag
	(*Message)(nil),     // 1: grpcsh.Message
	(*Result)(nil),      // 2: grpcsh.Result
	(*PeerMessage)(nil), // 3: grpcsh.PeerMessage
}
var file_messages_proto_depIdxs = []int32{
	0, // 0: grpcsh.Message.flag:type_name -> grpcsh.Flag
	0, // 1: grpcsh.Result.flag:type_name -> grpcsh.Flag
	0, // 2: grpcsh.PeerMessage.flag:type_name -> grpcsh.Flag
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_messages_proto_init() }
func file_messages_proto_init() {
	if File_messages_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_messages_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Message); i {
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
		file_messages_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Result); i {
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
		file_messages_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*PeerMessage); i {
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
			RawDescriptor: file_messages_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_messages_proto_goTypes,
		DependencyIndexes: file_messages_proto_depIdxs,
		EnumInfos:         file_messages_proto_enumTypes,
		MessageInfos:      file_messages_proto_msgTypes,
	}.Build()
	File_messages_proto = out.File
	file_messages_proto_rawDesc = nil
	file_messages_proto_goTypes = nil
	file_messages_proto_depIdxs = nil
}
