// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hapi/release/release.proto

package release

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import chart "k8s.io/helm/pkg/proto/hapi/chart"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Release describes a deployment of a chart, together with the chart
// and the variables used to deploy that chart.
type Release struct {
	// Name is the name of the release
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Info provides information about a release
	Info *Info `protobuf:"bytes,2,opt,name=info,proto3" json:"info,omitempty"`
	// Chart is the chart that was released.
	Chart *chart.Chart `protobuf:"bytes,3,opt,name=chart,proto3" json:"chart,omitempty"`
	// Config is the set of extra Values added to the chart.
	// These values override the default values inside of the chart.
	Config *chart.Config `protobuf:"bytes,4,opt,name=config,proto3" json:"config,omitempty"`
	// Manifest is the string representation of the rendered template.
	Manifest string `protobuf:"bytes,5,opt,name=manifest,proto3" json:"manifest,omitempty"`
	// Hooks are all of the hooks declared for this release.
	Hooks []*Hook `protobuf:"bytes,6,rep,name=hooks,proto3" json:"hooks,omitempty"`
	// Version is an int32 which represents the version of the release.
	Version int32 `protobuf:"varint,7,opt,name=version,proto3" json:"version,omitempty"`
	// Namespace is the kubernetes namespace of the release.
	Namespace            string   `protobuf:"bytes,8,opt,name=namespace,proto3" json:"namespace,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Release) Reset()         { *m = Release{} }
func (m *Release) String() string { return proto.CompactTextString(m) }
func (*Release) ProtoMessage()    {}
func (*Release) Descriptor() ([]byte, []int) {
	return fileDescriptor_release_fa600adfb1fffc82, []int{0}
}
func (m *Release) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Release.Unmarshal(m, b)
}
func (m *Release) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Release.Marshal(b, m, deterministic)
}
func (dst *Release) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Release.Merge(dst, src)
}
func (m *Release) XXX_Size() int {
	return xxx_messageInfo_Release.Size(m)
}
func (m *Release) XXX_DiscardUnknown() {
	xxx_messageInfo_Release.DiscardUnknown(m)
}

var xxx_messageInfo_Release proto.InternalMessageInfo

func (m *Release) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Release) GetInfo() *Info {
	if m != nil {
		return m.Info
	}
	return nil
}

func (m *Release) GetChart() *chart.Chart {
	if m != nil {
		return m.Chart
	}
	return nil
}

func (m *Release) GetConfig() *chart.Config {
	if m != nil {
		return m.Config
	}
	return nil
}

func (m *Release) GetManifest() string {
	if m != nil {
		return m.Manifest
	}
	return ""
}

func (m *Release) GetHooks() []*Hook {
	if m != nil {
		return m.Hooks
	}
	return nil
}

func (m *Release) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *Release) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func init() {
	proto.RegisterType((*Release)(nil), "hapi.release.Release")
}

func init() {
	proto.RegisterFile("hapi/release/release.proto", fileDescriptor_release_fa600adfb1fffc82)
}

var fileDescriptor_release_fa600adfb1fffc82 = []byte{
	// 256 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0xbf, 0x4e, 0xc3, 0x40,
	0x0c, 0xc6, 0x95, 0x36, 0x7f, 0x1a, 0xc3, 0x82, 0x07, 0xb0, 0x22, 0x86, 0x88, 0x01, 0x22, 0x86,
	0x54, 0x82, 0x37, 0x80, 0x05, 0xd6, 0x1b, 0xd9, 0x8e, 0xe8, 0x42, 0x4e, 0xa5, 0xe7, 0x28, 0x17,
	0xf1, 0x2c, 0x3c, 0x2e, 0xba, 0x3f, 0x85, 0x94, 0x2e, 0x4e, 0xec, 0xdf, 0xa7, 0xcf, 0xdf, 0x19,
	0xaa, 0x41, 0x8e, 0x7a, 0x3b, 0xa9, 0x4f, 0x25, 0xad, 0x3a, 0x7c, 0xdb, 0x71, 0xe2, 0x99, 0xf1,
	0xdc, 0xb1, 0x36, 0xce, 0xaa, 0xab, 0x23, 0xe5, 0xc0, 0xbc, 0x0b, 0xb2, 0x7f, 0x40, 0x9b, 0x9e,
	0x8f, 0x40, 0x37, 0xc8, 0x69, 0xde, 0x76, 0x6c, 0x7a, 0xfd, 0x11, 0xc1, 0xe5, 0x12, 0xb8, 0x1a,
	0xe6, 0x37, 0xdf, 0x2b, 0x28, 0x44, 0xf0, 0x41, 0x84, 0xd4, 0xc8, 0xbd, 0xa2, 0xa4, 0x4e, 0x9a,
	0x52, 0xf8, 0x7f, 0xbc, 0x85, 0xd4, 0xd9, 0xd3, 0xaa, 0x4e, 0x9a, 0xb3, 0x07, 0x6c, 0x97, 0xf9,
	0xda, 0x57, 0xd3, 0xb3, 0xf0, 0x1c, 0xef, 0x20, 0xf3, 0xb6, 0xb4, 0xf6, 0xc2, 0x8b, 0x20, 0x0c,
	0x9b, 0x9e, 0x5d, 0x15, 0x81, 0xe3, 0x3d, 0xe4, 0x21, 0x18, 0xa5, 0x4b, 0xcb, 0xa8, 0xf4, 0x44,
	0x44, 0x05, 0x56, 0xb0, 0xd9, 0x4b, 0xa3, 0x7b, 0x65, 0x67, 0xca, 0x7c, 0xa8, 0xdf, 0x1e, 0x1b,
	0xc8, 0xdc, 0x41, 0x2c, 0xe5, 0xf5, 0xfa, 0x34, 0xd9, 0x0b, 0xf3, 0x4e, 0x04, 0x01, 0x12, 0x14,
	0x5f, 0x6a, 0xb2, 0x9a, 0x0d, 0x15, 0x75, 0xd2, 0x64, 0xe2, 0xd0, 0xe2, 0x35, 0x94, 0xee, 0x91,
	0x76, 0x94, 0x9d, 0xa2, 0x8d, 0x5f, 0xf0, 0x37, 0x78, 0x2a, 0xdf, 0x8a, 0x68, 0xf7, 0x9e, 0xfb,
	0x63, 0x3d, 0xfe, 0x04, 0x00, 0x00, 0xff, 0xff, 0xc8, 0x8f, 0xec, 0x97, 0xbb, 0x01, 0x00, 0x00,
}
