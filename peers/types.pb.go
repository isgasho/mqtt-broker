// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

/*
Package peers is a generated protocol buffer package.

It is generated from these files:
	types.proto

It has these top-level messages:
	Peer
	ComputeUsage
	MemoryUsage
	PeerList
*/
package peers

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Peer struct {
	ID           string        `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	MeshID       uint64        `protobuf:"varint,2,opt,name=MeshID" json:"MeshID,omitempty"`
	Hostname     string        `protobuf:"bytes,3,opt,name=Hostname" json:"Hostname,omitempty"`
	Address      string        `protobuf:"bytes,4,opt,name=Address" json:"Address,omitempty"`
	LastAdded    int64         `protobuf:"varint,5,opt,name=LastAdded" json:"LastAdded,omitempty"`
	LastDeleted  int64         `protobuf:"varint,6,opt,name=LastDeleted" json:"LastDeleted,omitempty"`
	MemoryUsage  *MemoryUsage  `protobuf:"bytes,7,opt,name=MemoryUsage" json:"MemoryUsage,omitempty"`
	ComputeUsage *ComputeUsage `protobuf:"bytes,8,opt,name=ComputeUsage" json:"ComputeUsage,omitempty"`
	Runtime      string        `protobuf:"bytes,9,opt,name=Runtime" json:"Runtime,omitempty"`
}

func (m *Peer) Reset()                    { *m = Peer{} }
func (m *Peer) String() string            { return proto.CompactTextString(m) }
func (*Peer) ProtoMessage()               {}
func (*Peer) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Peer) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Peer) GetMeshID() uint64 {
	if m != nil {
		return m.MeshID
	}
	return 0
}

func (m *Peer) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

func (m *Peer) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Peer) GetLastAdded() int64 {
	if m != nil {
		return m.LastAdded
	}
	return 0
}

func (m *Peer) GetLastDeleted() int64 {
	if m != nil {
		return m.LastDeleted
	}
	return 0
}

func (m *Peer) GetMemoryUsage() *MemoryUsage {
	if m != nil {
		return m.MemoryUsage
	}
	return nil
}

func (m *Peer) GetComputeUsage() *ComputeUsage {
	if m != nil {
		return m.ComputeUsage
	}
	return nil
}

func (m *Peer) GetRuntime() string {
	if m != nil {
		return m.Runtime
	}
	return ""
}

type ComputeUsage struct {
	Cores      int64 `protobuf:"varint,1,opt,name=Cores" json:"Cores,omitempty"`
	Goroutines int64 `protobuf:"varint,2,opt,name=Goroutines" json:"Goroutines,omitempty"`
}

func (m *ComputeUsage) Reset()                    { *m = ComputeUsage{} }
func (m *ComputeUsage) String() string            { return proto.CompactTextString(m) }
func (*ComputeUsage) ProtoMessage()               {}
func (*ComputeUsage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ComputeUsage) GetCores() int64 {
	if m != nil {
		return m.Cores
	}
	return 0
}

func (m *ComputeUsage) GetGoroutines() int64 {
	if m != nil {
		return m.Goroutines
	}
	return 0
}

type MemoryUsage struct {
	Alloc      uint64 `protobuf:"varint,1,opt,name=Alloc" json:"Alloc,omitempty"`
	TotalAlloc uint64 `protobuf:"varint,2,opt,name=TotalAlloc" json:"TotalAlloc,omitempty"`
	Sys        uint64 `protobuf:"varint,3,opt,name=Sys" json:"Sys,omitempty"`
	NumGC      uint32 `protobuf:"varint,4,opt,name=NumGC" json:"NumGC,omitempty"`
}

func (m *MemoryUsage) Reset()                    { *m = MemoryUsage{} }
func (m *MemoryUsage) String() string            { return proto.CompactTextString(m) }
func (*MemoryUsage) ProtoMessage()               {}
func (*MemoryUsage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *MemoryUsage) GetAlloc() uint64 {
	if m != nil {
		return m.Alloc
	}
	return 0
}

func (m *MemoryUsage) GetTotalAlloc() uint64 {
	if m != nil {
		return m.TotalAlloc
	}
	return 0
}

func (m *MemoryUsage) GetSys() uint64 {
	if m != nil {
		return m.Sys
	}
	return 0
}

func (m *MemoryUsage) GetNumGC() uint32 {
	if m != nil {
		return m.NumGC
	}
	return 0
}

type PeerList struct {
	Peers []*Peer `protobuf:"bytes,1,rep,name=Peers" json:"Peers,omitempty"`
}

func (m *PeerList) Reset()                    { *m = PeerList{} }
func (m *PeerList) String() string            { return proto.CompactTextString(m) }
func (*PeerList) ProtoMessage()               {}
func (*PeerList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *PeerList) GetPeers() []*Peer {
	if m != nil {
		return m.Peers
	}
	return nil
}

func init() {
	proto.RegisterType((*Peer)(nil), "peers.Peer")
	proto.RegisterType((*ComputeUsage)(nil), "peers.ComputeUsage")
	proto.RegisterType((*MemoryUsage)(nil), "peers.MemoryUsage")
	proto.RegisterType((*PeerList)(nil), "peers.PeerList")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 337 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x92, 0x4d, 0x4b, 0xfb, 0x40,
	0x10, 0xc6, 0xc9, 0x5b, 0x5f, 0x26, 0xff, 0xbf, 0xc8, 0x28, 0xb2, 0x88, 0x48, 0xcc, 0x29, 0x17,
	0x7b, 0xa8, 0x82, 0xe7, 0xd2, 0x40, 0x2d, 0xb4, 0x22, 0xab, 0x7e, 0x80, 0x68, 0x06, 0x2d, 0x26,
	0xdd, 0xb0, 0xbb, 0x39, 0xf4, 0x43, 0xf9, 0x1d, 0x65, 0x77, 0xa3, 0x26, 0xb7, 0xfd, 0x3d, 0xcf,
	0xcc, 0x43, 0x66, 0x26, 0x10, 0xeb, 0x43, 0x43, 0x6a, 0xd6, 0x48, 0xa1, 0x05, 0x46, 0x0d, 0x91,
	0x54, 0xe9, 0x97, 0x0f, 0xe1, 0x23, 0x91, 0xc4, 0x23, 0xf0, 0xd7, 0x39, 0xf3, 0x12, 0x2f, 0x9b,
	0x72, 0x7f, 0x9d, 0xe3, 0x19, 0x8c, 0xb6, 0xa4, 0x3e, 0xd6, 0x39, 0xf3, 0x13, 0x2f, 0x0b, 0x79,
	0x47, 0x78, 0x0e, 0x93, 0x7b, 0xa1, 0xf4, 0xbe, 0xa8, 0x89, 0x05, 0xb6, 0xfa, 0x97, 0x91, 0xc1,
	0x78, 0x51, 0x96, 0x92, 0x94, 0x62, 0xa1, 0xb5, 0x7e, 0x10, 0x2f, 0x60, 0xba, 0x29, 0x94, 0x5e,
	0x94, 0x25, 0x95, 0x2c, 0x4a, 0xbc, 0x2c, 0xe0, 0x7f, 0x02, 0x26, 0x10, 0x1b, 0xc8, 0xa9, 0x22,
	0x4d, 0x25, 0x1b, 0x59, 0xbf, 0x2f, 0xe1, 0x2d, 0xc4, 0x5b, 0xaa, 0x85, 0x3c, 0xbc, 0xa8, 0xe2,
	0x9d, 0xd8, 0x38, 0xf1, 0xb2, 0x78, 0x8e, 0x33, 0x3b, 0xc3, 0xac, 0xe7, 0xf0, 0x7e, 0x19, 0xde,
	0xc1, 0xbf, 0xa5, 0xa8, 0x9b, 0x56, 0x93, 0x6b, 0x9b, 0xd8, 0xb6, 0x93, 0xae, 0xad, 0x6f, 0xf1,
	0x41, 0xa1, 0x19, 0x84, 0xb7, 0x7b, 0xbd, 0xab, 0x89, 0x4d, 0xdd, 0x20, 0x1d, 0xa6, 0xf9, 0x30,
	0x12, 0x4f, 0x21, 0x5a, 0x0a, 0x49, 0xca, 0x6e, 0x2e, 0xe0, 0x0e, 0xf0, 0x12, 0x60, 0x25, 0xa4,
	0x68, 0xf5, 0x6e, 0x4f, 0xca, 0x2e, 0x30, 0xe0, 0x3d, 0x25, 0xfd, 0x1c, 0x8c, 0x63, 0x42, 0x16,
	0x55, 0x25, 0xde, 0x6c, 0x48, 0xc8, 0x1d, 0x98, 0x90, 0x67, 0xa1, 0x8b, 0xca, 0x59, 0xee, 0x0a,
	0x3d, 0x05, 0x8f, 0x21, 0x78, 0x3a, 0x28, 0x7b, 0x84, 0x90, 0x9b, 0xa7, 0xc9, 0x79, 0x68, 0xeb,
	0xd5, 0xd2, 0x6e, 0xff, 0x3f, 0x77, 0x90, 0x5e, 0xc3, 0xc4, 0x5c, 0x78, 0xb3, 0x53, 0x1a, 0xaf,
	0x20, 0x32, 0x6f, 0xf3, 0xb9, 0x41, 0x16, 0xcf, 0xe3, 0x6e, 0x15, 0x46, 0xe3, 0xce, 0x79, 0x1d,
	0xd9, 0xff, 0xe3, 0xe6, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x08, 0x50, 0x67, 0x36, 0x2e, 0x02, 0x00,
	0x00,
}
