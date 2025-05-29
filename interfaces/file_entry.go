package interfaces

import (
	"time"

	"github.com/google/uuid"
)

type IFileEntry interface {
	GetID() uuid.UUID
	GetParentID() uuid.UUID
	GetParent() IFileEntry
	GetType() string
	GetName() string
	GetOriginName() string
	GetPath() string
	GetDepth() int
	GetSize() int64
	GetCreatedAt() time.Time
	GetCreatedBy() string
	GetModifiedAt() *time.Time
	GetModifiedBy() string
	GetPermissions() string
	GetChecksum() string
	GetComments() string
	GetMetadata() IJsonB
	SetID(id uuid.UUID)
	SetParentID(parentID uuid.UUID)
	SetType(entryType string)
	SetName(name string)
	SetOriginName(originName string)
	SetDepth(depth int)
	SetSize(size int64)
	SetCreatedAt(createdAt time.Time)
	SetCreatedBy(createdBy string)
	SetModifiedAt(modifiedAt *time.Time)
	SetModifiedBy(modifiedBy string)
	SetPermissions(permissions string)
	SetChecksum(checksum string)
	SetComments(comments string)
	SetMetadata(metadata IJsonB)
	SetParent(parent IFileEntry)
	String() string
}
