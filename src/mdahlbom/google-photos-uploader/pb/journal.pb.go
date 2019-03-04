// Code generated by protoc-gen-go. DO NOT EDIT.
// source: journal.proto

package pb // import "mdahlbom/google-photos-uploader/pb"

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

type Journal struct {
	// List of files / subdirectories
	Entries              []*JournalEntry `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Journal) Reset()         { *m = Journal{} }
func (m *Journal) String() string { return proto.CompactTextString(m) }
func (*Journal) ProtoMessage()    {}
func (*Journal) Descriptor() ([]byte, []int) {
	return fileDescriptor_journal_f6f874f3804ca75d, []int{0}
}
func (m *Journal) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Journal.Unmarshal(m, b)
}
func (m *Journal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Journal.Marshal(b, m, deterministic)
}
func (dst *Journal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Journal.Merge(dst, src)
}
func (m *Journal) XXX_Size() int {
	return xxx_messageInfo_Journal.Size(m)
}
func (m *Journal) XXX_DiscardUnknown() {
	xxx_messageInfo_Journal.DiscardUnknown(m)
}

var xxx_messageInfo_Journal proto.InternalMessageInfo

func (m *Journal) GetEntries() []*JournalEntry {
	if m != nil {
		return m.Entries
	}
	return nil
}

type JournalEntry struct {
	// Name (relative) of the file or directory
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Upload token for this journal entry
	UploadToken string `protobuf:"bytes,2,opt,name=upload_token,json=uploadToken,proto3" json:"upload_token,omitempty"`
	// Whether this a media item has been successfully created for this item
	MediaItemCreated bool `protobuf:"varint,3,opt,name=media_item_created,json=mediaItemCreated,proto3" json:"media_item_created,omitempty"`
	// Whether this item has successfully been added to an album
	AddedToAlbum         bool     `protobuf:"varint,4,opt,name=added_to_album,json=addedToAlbum,proto3" json:"added_to_album,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *JournalEntry) Reset()         { *m = JournalEntry{} }
func (m *JournalEntry) String() string { return proto.CompactTextString(m) }
func (*JournalEntry) ProtoMessage()    {}
func (*JournalEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_journal_f6f874f3804ca75d, []int{1}
}
func (m *JournalEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_JournalEntry.Unmarshal(m, b)
}
func (m *JournalEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_JournalEntry.Marshal(b, m, deterministic)
}
func (dst *JournalEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_JournalEntry.Merge(dst, src)
}
func (m *JournalEntry) XXX_Size() int {
	return xxx_messageInfo_JournalEntry.Size(m)
}
func (m *JournalEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_JournalEntry.DiscardUnknown(m)
}

var xxx_messageInfo_JournalEntry proto.InternalMessageInfo

func (m *JournalEntry) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *JournalEntry) GetUploadToken() string {
	if m != nil {
		return m.UploadToken
	}
	return ""
}

func (m *JournalEntry) GetMediaItemCreated() bool {
	if m != nil {
		return m.MediaItemCreated
	}
	return false
}

func (m *JournalEntry) GetAddedToAlbum() bool {
	if m != nil {
		return m.AddedToAlbum
	}
	return false
}

func init() {
	proto.RegisterType((*Journal)(nil), "uploader.Journal")
	proto.RegisterType((*JournalEntry)(nil), "uploader.JournalEntry")
}

func init() { proto.RegisterFile("journal.proto", fileDescriptor_journal_f6f874f3804ca75d) }

var fileDescriptor_journal_f6f874f3804ca75d = []byte{
	// 232 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x90, 0x31, 0x4f, 0xc3, 0x30,
	0x10, 0x46, 0x65, 0x5a, 0xd1, 0xe2, 0x06, 0x84, 0x3c, 0x20, 0x8f, 0x21, 0xea, 0x90, 0x81, 0xa6,
	0x08, 0x46, 0x26, 0x40, 0x0c, 0x30, 0x46, 0x9d, 0x58, 0xa2, 0x0b, 0x3e, 0xb5, 0x01, 0xdb, 0x17,
	0xb9, 0x97, 0x81, 0x9f, 0xc2, 0xbf, 0x45, 0xb1, 0x89, 0xc4, 0x66, 0xbd, 0xf7, 0xf4, 0x49, 0x3e,
	0x79, 0xfe, 0x49, 0x43, 0xf0, 0x60, 0xab, 0x3e, 0x10, 0x93, 0x5a, 0x0e, 0xbd, 0x25, 0x30, 0x18,
	0x8a, 0x07, 0xb9, 0x78, 0x4b, 0x4a, 0xdd, 0xca, 0x05, 0x7a, 0x0e, 0x1d, 0x1e, 0xb5, 0xc8, 0x67,
	0xe5, 0xea, 0xee, 0xaa, 0x9a, 0xb2, 0xea, 0xaf, 0x79, 0xf1, 0x1c, 0xbe, 0xeb, 0x29, 0x2b, 0x7e,
	0x84, 0xcc, 0xfe, 0x1b, 0xa5, 0xe4, 0xdc, 0x83, 0x43, 0x2d, 0x72, 0x51, 0x9e, 0xd5, 0xf1, 0xad,
	0xae, 0x65, 0x96, 0x66, 0x1a, 0xa6, 0x2f, 0xf4, 0xfa, 0x24, 0xba, 0x55, 0x62, 0xbb, 0x11, 0xa9,
	0x1b, 0xa9, 0x1c, 0x9a, 0x0e, 0x9a, 0x8e, 0xd1, 0x35, 0x1f, 0x01, 0x81, 0xd1, 0xe8, 0x59, 0x2e,
	0xca, 0x65, 0x7d, 0x19, 0xcd, 0x2b, 0xa3, 0x7b, 0x4e, 0x5c, 0xad, 0xe5, 0x05, 0x18, 0x83, 0xe3,
	0x5e, 0x03, 0xb6, 0x1d, 0x9c, 0x9e, 0xc7, 0x32, 0x8b, 0x74, 0x47, 0x8f, 0x23, 0x7b, 0x5a, 0xbf,
	0x17, 0xce, 0xc0, 0xc1, 0xb6, 0xe4, 0xb6, 0x7b, 0xa2, 0xbd, 0xc5, 0x4d, 0x7f, 0x20, 0xa6, 0xe3,
	0x66, 0xfa, 0xd4, 0xb6, 0x6f, 0xdb, 0xd3, 0x78, 0x8f, 0xfb, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff,
	0xa0, 0xdb, 0xca, 0x42, 0x20, 0x01, 0x00, 0x00,
}
