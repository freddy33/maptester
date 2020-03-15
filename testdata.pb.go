// Code generated by protoc-gen-go. DO NOT EDIT.
// source: testdata.proto

package maptester

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type TestValue struct {
	SVal                 string   `protobuf:"bytes,1,opt,name=sVal,proto3" json:"sVal,omitempty"`
	Idx                  int64    `protobuf:"varint,2,opt,name=idx,proto3" json:"idx,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TestValue) Reset()         { *m = TestValue{} }
func (m *TestValue) String() string { return proto.CompactTextString(m) }
func (*TestValue) ProtoMessage()    {}
func (*TestValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_40c4782d007dfce9, []int{0}
}

func (m *TestValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TestValue.Unmarshal(m, b)
}
func (m *TestValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TestValue.Marshal(b, m, deterministic)
}
func (m *TestValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TestValue.Merge(m, src)
}
func (m *TestValue) XXX_Size() int {
	return xxx_messageInfo_TestValue.Size(m)
}
func (m *TestValue) XXX_DiscardUnknown() {
	xxx_messageInfo_TestValue.DiscardUnknown(m)
}

var xxx_messageInfo_TestValue proto.InternalMessageInfo

func (m *TestValue) GetSVal() string {
	if m != nil {
		return m.SVal
	}
	return ""
}

func (m *TestValue) GetIdx() int64 {
	if m != nil {
		return m.Idx
	}
	return 0
}

type IntTestLine struct {
	Key                  []int64    `protobuf:"varint,1,rep,packed,name=key,proto3" json:"key,omitempty"`
	Value                *TestValue `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *IntTestLine) Reset()         { *m = IntTestLine{} }
func (m *IntTestLine) String() string { return proto.CompactTextString(m) }
func (*IntTestLine) ProtoMessage()    {}
func (*IntTestLine) Descriptor() ([]byte, []int) {
	return fileDescriptor_40c4782d007dfce9, []int{1}
}

func (m *IntTestLine) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IntTestLine.Unmarshal(m, b)
}
func (m *IntTestLine) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IntTestLine.Marshal(b, m, deterministic)
}
func (m *IntTestLine) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IntTestLine.Merge(m, src)
}
func (m *IntTestLine) XXX_Size() int {
	return xxx_messageInfo_IntTestLine.Size(m)
}
func (m *IntTestLine) XXX_DiscardUnknown() {
	xxx_messageInfo_IntTestLine.DiscardUnknown(m)
}

var xxx_messageInfo_IntTestLine proto.InternalMessageInfo

func (m *IntTestLine) GetKey() []int64 {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *IntTestLine) GetValue() *TestValue {
	if m != nil {
		return m.Value
	}
	return nil
}

type StringTestLine struct {
	Key                  string     `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                *TestValue `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *StringTestLine) Reset()         { *m = StringTestLine{} }
func (m *StringTestLine) String() string { return proto.CompactTextString(m) }
func (*StringTestLine) ProtoMessage()    {}
func (*StringTestLine) Descriptor() ([]byte, []int) {
	return fileDescriptor_40c4782d007dfce9, []int{2}
}

func (m *StringTestLine) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StringTestLine.Unmarshal(m, b)
}
func (m *StringTestLine) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StringTestLine.Marshal(b, m, deterministic)
}
func (m *StringTestLine) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StringTestLine.Merge(m, src)
}
func (m *StringTestLine) XXX_Size() int {
	return xxx_messageInfo_StringTestLine.Size(m)
}
func (m *StringTestLine) XXX_DiscardUnknown() {
	xxx_messageInfo_StringTestLine.DiscardUnknown(m)
}

var xxx_messageInfo_StringTestLine proto.InternalMessageInfo

func (m *StringTestLine) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *StringTestLine) GetValue() *TestValue {
	if m != nil {
		return m.Value
	}
	return nil
}

type MapTestResult struct {
	NbLines              int32    `protobuf:"varint,1,opt,name=nbLines,proto3" json:"nbLines,omitempty"`
	NbEntries            int32    `protobuf:"varint,2,opt,name=nbEntries,proto3" json:"nbEntries,omitempty"`
	NbSameKeys           int32    `protobuf:"varint,3,opt,name=nbSameKeys,proto3" json:"nbSameKeys,omitempty"`
	NbOfTimesSameKey     []int32  `protobuf:"varint,4,rep,packed,name=nbOfTimesSameKey,proto3" json:"nbOfTimesSameKey,omitempty"`
	OffsetsPerThreads    []int32  `protobuf:"varint,5,rep,packed,name=offsetsPerThreads,proto3" json:"offsetsPerThreads,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MapTestResult) Reset()         { *m = MapTestResult{} }
func (m *MapTestResult) String() string { return proto.CompactTextString(m) }
func (*MapTestResult) ProtoMessage()    {}
func (*MapTestResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_40c4782d007dfce9, []int{3}
}

func (m *MapTestResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MapTestResult.Unmarshal(m, b)
}
func (m *MapTestResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MapTestResult.Marshal(b, m, deterministic)
}
func (m *MapTestResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MapTestResult.Merge(m, src)
}
func (m *MapTestResult) XXX_Size() int {
	return xxx_messageInfo_MapTestResult.Size(m)
}
func (m *MapTestResult) XXX_DiscardUnknown() {
	xxx_messageInfo_MapTestResult.DiscardUnknown(m)
}

var xxx_messageInfo_MapTestResult proto.InternalMessageInfo

func (m *MapTestResult) GetNbLines() int32 {
	if m != nil {
		return m.NbLines
	}
	return 0
}

func (m *MapTestResult) GetNbEntries() int32 {
	if m != nil {
		return m.NbEntries
	}
	return 0
}

func (m *MapTestResult) GetNbSameKeys() int32 {
	if m != nil {
		return m.NbSameKeys
	}
	return 0
}

func (m *MapTestResult) GetNbOfTimesSameKey() []int32 {
	if m != nil {
		return m.NbOfTimesSameKey
	}
	return nil
}

func (m *MapTestResult) GetOffsetsPerThreads() []int32 {
	if m != nil {
		return m.OffsetsPerThreads
	}
	return nil
}

func init() {
	proto.RegisterType((*TestValue)(nil), "maptester.TestValue")
	proto.RegisterType((*IntTestLine)(nil), "maptester.IntTestLine")
	proto.RegisterType((*StringTestLine)(nil), "maptester.StringTestLine")
	proto.RegisterType((*MapTestResult)(nil), "maptester.MapTestResult")
}

func init() {
	proto.RegisterFile("testdata.proto", fileDescriptor_40c4782d007dfce9)
}

var fileDescriptor_40c4782d007dfce9 = []byte{
	// 271 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x91, 0x4f, 0x4b, 0xf3, 0x40,
	0x10, 0x87, 0xc9, 0x9b, 0xe6, 0x95, 0x4c, 0xb1, 0xd4, 0xc5, 0xc3, 0x1e, 0x44, 0x42, 0x4e, 0xa1,
	0x48, 0x40, 0xfd, 0x0c, 0x1e, 0xa4, 0xfe, 0x63, 0x1b, 0x7a, 0xdf, 0x90, 0x89, 0x86, 0x26, 0x9b,
	0xb0, 0x33, 0x15, 0xf3, 0xf1, 0xfc, 0x66, 0xb2, 0xab, 0x56, 0x21, 0x27, 0x6f, 0xb3, 0xcf, 0xfc,
	0xe6, 0x99, 0x85, 0x81, 0x05, 0x23, 0x71, 0xa5, 0x59, 0xe7, 0x83, 0xed, 0xb9, 0x17, 0x71, 0xa7,
	0x07, 0x87, 0xd0, 0xa6, 0x97, 0x10, 0x17, 0x48, 0xbc, 0xd5, 0xed, 0x1e, 0x85, 0x80, 0x19, 0x6d,
	0x75, 0x2b, 0x83, 0x24, 0xc8, 0x62, 0xe5, 0x6b, 0xb1, 0x84, 0xb0, 0xa9, 0xde, 0xe4, 0xbf, 0x24,
	0xc8, 0x42, 0xe5, 0xca, 0x74, 0x0d, 0xf3, 0x5b, 0xc3, 0x6e, 0xea, 0xae, 0x31, 0xe8, 0x02, 0x3b,
	0x1c, 0x65, 0x90, 0x84, 0x2e, 0xb0, 0xc3, 0x51, 0xac, 0x20, 0x7a, 0x75, 0x3e, 0x3f, 0x34, 0xbf,
	0x3a, 0xcd, 0x0f, 0xeb, 0xf2, 0xc3, 0x2e, 0xf5, 0x19, 0x49, 0x1f, 0x60, 0xb1, 0x61, 0xdb, 0x98,
	0xe7, 0xa9, 0xcf, 0xfd, 0xe1, 0xcf, 0xbe, 0xf7, 0x00, 0x8e, 0xef, 0xf5, 0xe0, 0xb8, 0x42, 0xda,
	0xb7, 0x2c, 0x24, 0x1c, 0x99, 0xd2, 0x99, 0xc9, 0x3b, 0x23, 0xf5, 0xfd, 0x14, 0x67, 0x10, 0x9b,
	0xf2, 0xc6, 0xb0, 0x6d, 0x90, 0xbc, 0x3b, 0x52, 0x3f, 0x40, 0x9c, 0x03, 0x98, 0x72, 0xa3, 0x3b,
	0x5c, 0xe3, 0x48, 0x32, 0xf4, 0xed, 0x5f, 0x44, 0xac, 0x60, 0x69, 0xca, 0xc7, 0xba, 0x68, 0x3a,
	0xa4, 0x2f, 0x28, 0x67, 0x49, 0x98, 0x45, 0x6a, 0xc2, 0xc5, 0x05, 0x9c, 0xf4, 0x75, 0x4d, 0xc8,
	0xf4, 0x84, 0xb6, 0x78, 0xb1, 0xa8, 0x2b, 0x92, 0x91, 0x0f, 0x4f, 0x1b, 0xe5, 0x7f, 0x7f, 0xa5,
	0xeb, 0x8f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x2d, 0x2a, 0xe6, 0xe5, 0xb7, 0x01, 0x00, 0x00,
}
