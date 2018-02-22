// Code generated by protoc-gen-go. DO NOT EDIT.
// source: journal.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	journal.proto

It has these top-level messages:
	Journal
	JournalEntry
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Journal struct {
	// List of files / subdirectories
	Entries []*JournalEntry `protobuf:"bytes,1,rep,name=entries" json:"entries,omitempty"`
	// Album title for this directory - set when album successfully created
	AlbumTitle string `protobuf:"bytes,2,opt,name=album_title,json=albumTitle" json:"album_title,omitempty"`
	// Album ID for this directory - set when album successfully created
	AlbumId string `protobuf:"bytes,3,opt,name=album_id,json=albumId" json:"album_id,omitempty"`
}

func (m *Journal) Reset()                    { *m = Journal{} }
func (m *Journal) String() string            { return proto.CompactTextString(m) }
func (*Journal) ProtoMessage()               {}
func (*Journal) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Journal) GetEntries() []*JournalEntry {
	if m != nil {
		return m.Entries
	}
	return nil
}

func (m *Journal) GetAlbumTitle() string {
	if m != nil {
		return m.AlbumTitle
	}
	return ""
}

func (m *Journal) GetAlbumId() string {
	if m != nil {
		return m.AlbumId
	}
	return ""
}

type JournalEntry struct {
	// Name (relative) of the file or directory
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// Whether the entry is a directory
	IsDirectory bool `protobuf:"varint,2,opt,name=is_directory,json=isDirectory" json:"is_directory,omitempty"`
	// When the entry was completed (uploaded)
	Completed *google_protobuf.Timestamp `protobuf:"bytes,3,opt,name=completed" json:"completed,omitempty"`
}

func (m *JournalEntry) Reset()                    { *m = JournalEntry{} }
func (m *JournalEntry) String() string            { return proto.CompactTextString(m) }
func (*JournalEntry) ProtoMessage()               {}
func (*JournalEntry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *JournalEntry) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *JournalEntry) GetIsDirectory() bool {
	if m != nil {
		return m.IsDirectory
	}
	return false
}

func (m *JournalEntry) GetCompleted() *google_protobuf.Timestamp {
	if m != nil {
		return m.Completed
	}
	return nil
}

func init() {
	proto.RegisterType((*Journal)(nil), "uploader.Journal")
	proto.RegisterType((*JournalEntry)(nil), "uploader.JournalEntry")
}

func init() { proto.RegisterFile("journal.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 255 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x8f, 0x31, 0x4f, 0xc3, 0x30,
	0x10, 0x85, 0x15, 0x8a, 0x48, 0xea, 0x94, 0xc5, 0x03, 0x0a, 0x5d, 0x5a, 0x22, 0x86, 0x2e, 0x75,
	0x50, 0x59, 0x98, 0x11, 0x0c, 0x30, 0x46, 0x9d, 0x58, 0x2a, 0xa7, 0x39, 0x5a, 0x23, 0x3b, 0x67,
	0xd9, 0x17, 0xa1, 0x4e, 0xfc, 0x75, 0x84, 0x8d, 0x05, 0x9b, 0xfd, 0xee, 0xbb, 0x77, 0xfa, 0xd8,
	0xe5, 0x07, 0x8e, 0x6e, 0x90, 0x5a, 0x58, 0x87, 0x84, 0xbc, 0x18, 0xad, 0x46, 0xd9, 0x83, 0x9b,
	0x2f, 0x0e, 0x88, 0x07, 0x0d, 0x4d, 0xc8, 0xbb, 0xf1, 0xbd, 0x21, 0x65, 0xc0, 0x93, 0x34, 0x36,
	0xa2, 0xf5, 0x27, 0xcb, 0x5f, 0xe3, 0x2e, 0xbf, 0x63, 0x39, 0x0c, 0xe4, 0x14, 0xf8, 0x2a, 0x5b,
	0x4e, 0x56, 0xe5, 0xe6, 0x4a, 0xa4, 0x1e, 0xf1, 0xcb, 0x3c, 0x0f, 0xe4, 0x4e, 0x6d, 0xc2, 0xf8,
	0x82, 0x95, 0x52, 0x77, 0xa3, 0xd9, 0x91, 0x22, 0x0d, 0xd5, 0xd9, 0x32, 0x5b, 0x4d, 0x5b, 0x16,
	0xa2, 0xed, 0x4f, 0xc2, 0xaf, 0x59, 0x11, 0x01, 0xd5, 0x57, 0x93, 0x30, 0xcd, 0xc3, 0xff, 0xa5,
	0xaf, 0xbf, 0xd8, 0xec, 0x7f, 0x29, 0xe7, 0xec, 0x7c, 0x90, 0x06, 0xaa, 0x2c, 0x60, 0xe1, 0xcd,
	0x6f, 0xd8, 0x4c, 0xf9, 0x5d, 0xaf, 0x1c, 0xec, 0x09, 0xdd, 0x29, 0x1c, 0x28, 0xda, 0x52, 0xf9,
	0xa7, 0x14, 0xf1, 0x07, 0x36, 0xdd, 0xa3, 0xb1, 0x1a, 0x08, 0xe2, 0x89, 0x72, 0x33, 0x17, 0x51,
	0x5a, 0x24, 0x69, 0xb1, 0x4d, 0xd2, 0xed, 0x1f, 0xfc, 0x78, 0xfb, 0x56, 0x9b, 0x5e, 0x1e, 0x75,
	0x87, 0xa6, 0x89, 0x0b, 0x6b, 0x7b, 0x44, 0x42, 0xbf, 0x4e, 0xd6, 0x8d, 0xed, 0xba, 0x8b, 0x50,
	0x72, 0xff, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x7c, 0x99, 0x23, 0x17, 0x62, 0x01, 0x00, 0x00,
}
