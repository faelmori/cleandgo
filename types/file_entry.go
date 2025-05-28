package types

import (
	"fmt"
	"time"

	it "github.com/faelmori/cleandgo/interfaces"
	gl "github.com/faelmori/cleandgo/logger"

	"github.com/google/uuid"
)

type FileEntry struct {
	*Mutexes
	ID          uuid.UUID     `json:"id" yaml:"id" xml:"id" toml:"id" gorm:"type:uuid,default:uuid_generate_v4()"`                 // ID único do arquivo ou diretório
	ParentID    uuid.UUID     `json:"parentId" yaml:"parentId" xml:"parentId" toml:"parentId" gorm:"foreignkey:ID;type:uuid"`      // ID do diretório pai (`00000000-0000-0000-0000-000000000000` se for raiz)
	Type        string        `json:"type" yaml:"type" xml:"type" toml:"type" gorm:"type"`                                         // "file" ou "directory"
	Name        string        `json:"name" yaml:"name" xml:"name" toml:"name" gorm:"name"`                                         // Nome real do artefato
	OriginName  string        `json:"originName" yaml:"originName" xml:"originName" toml:"originName" gorm:"omitempty,originName"` // Nome original do arquivo de árvore
	Depth       int           `json:"depth" yaml:"depth" xml:"depth" toml:"depth" gorm:"depth"`                                    // Profundidade hierárquica
	Size        int64         `json:"size" yaml:"size" xml:"size" toml:"size" gorm:"omitempty,size"`                               // Tamanho do arquivo em bytes
	CreatedAt   time.Time     `json:"createdAt" yaml:"createdAt" xml:"createdAt" toml:"createdAt" gorm:"createdAt"`                // Data de criação
	CreatedBy   string        `json:"createdBy" yaml:"createdBy" xml:"createdBy" toml:"createdBy" gorm:"createdBy"`                // Usuário que criou o arquivo
	ModifiedAt  *time.Time    `json:"modifiedAt" yaml:"modifiedAt" xml:"modifiedAt" toml:"modifiedAt" gorm:"omitempty,modifiedAt"` // Data de modificação
	ModifiedBy  string        `json:"modifiedBy" yaml:"modifiedBy" xml:"modifiedBy" toml:"modifiedBy" gorm:"omitempty,modifiedBy"` // Usuário que modificou o arquivo
	Permissions string        `json:"permissions" yaml:"permissions" xml:"permissions" toml:"permissions" gorm:"permissions"`      // Permissões do arquivo (ex: "rwxr-xr-x")
	Checksum    string        `json:"checksum" yaml:"checksum" xml:"checksum" toml:"checksum" gorm:"omitempty,checksum"`           // Checksum do arquivo para integridade
	Comments    string        `json:"comments" yaml:"comments" xml:"comments" toml:"comments" gorm:"omitempty,comments"`           // Comentários adicionais sobre o arquivo
	Metadata    it.IJsonB     `json:"metadata" yaml:"metadata" xml:"metadata" toml:"metadata" gorm:"omitempty,type:jsonb"`         // Metadados adicionais em formato JSON
	Parent      it.IFileEntry `json:"parent" yaml:"parent" xml:"parent" toml:"parent" gorm:"foreignkey:ID;association_foreignkey:ParentID"`
}

func NewFileEntry(id, parentID uuid.UUID, entryType, name, originName string, depth int, size int64, comments string) (it.IFileEntry, error) {
	if originName == "" {
		gl.Log("error", "Origin name cannot be empty")
		return nil, fmt.Errorf("origin name cannot be empty")
	}

	if id == uuid.Nil {
		id = uuid.Must(uuid.NewRandom())
	}

	if entryType != "file" && entryType != "directory" && entryType != "symlink" && entryType != "unknown" {
		gl.Log("error", "Entry type must be 'file' or 'directory'")
		return nil, fmt.Errorf("entry type must be 'file' or 'directory'")
	}

	return &FileEntry{
		Mutexes:    NewMutexesType(),
		ID:         id,
		ParentID:   parentID,
		Type:       entryType,
		Name:       name,
		OriginName: originName,
		Depth:      depth,
		Size:       size,
		Comments:   comments,
	}, nil
}

func (fe *FileEntry) GetID() uuid.UUID       { return fe.ID }
func (fe *FileEntry) GetParentID() uuid.UUID { return fe.ParentID }
func (fe *FileEntry) GetType() string        { return fe.Type }
func (fe *FileEntry) GetName() string        { return fe.Name }
func (fe *FileEntry) GetOriginName() string  { return fe.OriginName }
func (fe *FileEntry) GetDepth() int          { return fe.Depth }
func (fe *FileEntry) GetSize() int64         { return fe.Size }
func (fe *FileEntry) GetCreatedAt() time.Time {
	if fe.CreatedAt.IsZero() {
		return time.Now()
	}
	return fe.CreatedAt
}
func (fe *FileEntry) GetCreatedBy() string {
	if fe.CreatedBy == "" {
		return "system"
	}
	return fe.CreatedBy
}
func (fe *FileEntry) GetModifiedAt() *time.Time {
	if fe.ModifiedAt == nil {
		return nil
	}
	if fe.ModifiedAt.IsZero() {
		return nil
	}
	return fe.ModifiedAt
}
func (fe *FileEntry) GetModifiedBy() string {
	if fe.ModifiedBy == "" {
		return "system"
	}
	return fe.ModifiedBy
}
func (fe *FileEntry) GetPermissions() string {
	if fe.Permissions == "" {
		return "rwxr-xr-x" // Default permissions
	}
	return fe.Permissions
}
func (fe *FileEntry) GetChecksum() string {
	if fe.Checksum == "" {
		return "none" // Default checksum
	}
	return fe.Checksum
}
func (fe *FileEntry) GetComments() string {
	if fe.Comments == "" {
		return "No comments"
	}
	return fe.Comments
}
func (fe *FileEntry) GetMetadata() it.IJsonB {
	if fe.Metadata == nil {
		return nil
	}
	return fe.Metadata
}

func (fe *FileEntry) SetID(id uuid.UUID) {
	if id == uuid.Nil {
		gl.Log("error", "ID cannot be nil")
		return
	}
	fe.ID = id
}
func (fe *FileEntry) SetParentID(parentID uuid.UUID) {
	if parentID == uuid.Nil {
		gl.Log("error", "Parent ID cannot be nil")
		return
	}
	fe.ParentID = parentID
}
func (fe *FileEntry) SetType(entryType string) {
	if entryType != "file" && entryType != "directory" && entryType != "symlink" && entryType != "unknown" {
		gl.Log("error", "Entry type must be 'file', 'directory', 'symlink' or 'unknown'")
		return
	}
	fe.Type = entryType
}
func (fe *FileEntry) SetName(name string) {
	if name == "" {
		gl.Log("error", "Name cannot be empty")
		return
	}
	fe.Name = name
}
func (fe *FileEntry) SetOriginName(originName string) {
	if originName == "" {
		gl.Log("error", "Origin name cannot be empty")
		return
	}
	fe.OriginName = originName
}
func (fe *FileEntry) SetDepth(depth int) {
	if depth < 0 {
		gl.Log("error", "Depth cannot be negative")
		return
	}
	fe.Depth = depth
}
func (fe *FileEntry) SetSize(size int64) {
	if size < 0 {
		gl.Log("error", "Size cannot be negative")
		return
	}
	fe.Size = size
}
func (fe *FileEntry) SetCreatedAt(createdAt time.Time) {
	if createdAt.IsZero() {
		gl.Log("error", "CreatedAt cannot be zero")
		return
	}
	fe.CreatedAt = createdAt
}
func (fe *FileEntry) SetCreatedBy(createdBy string) {
	if createdBy == "" {
		gl.Log("error", "CreatedBy cannot be empty")
		return
	}
	fe.CreatedBy = createdBy
}
func (fe *FileEntry) SetModifiedAt(modifiedAt *time.Time) {
	if modifiedAt != nil && modifiedAt.IsZero() {
		gl.Log("error", "ModifiedAt cannot be zero")
		return
	}
	fe.ModifiedAt = modifiedAt
}
func (fe *FileEntry) SetModifiedBy(modifiedBy string) {
	if modifiedBy == "" {
		gl.Log("error", "ModifiedBy cannot be empty")
		return
	}
	fe.ModifiedBy = modifiedBy
}
func (fe *FileEntry) SetPermissions(permissions string) {
	if permissions == "" {
		gl.Log("error", "Permissions cannot be empty")
		return
	}
	fe.Permissions = permissions
}
func (fe *FileEntry) SetChecksum(checksum string) {
	if checksum == "" {
		gl.Log("error", "Checksum cannot be empty")
		return
	}
	fe.Checksum = checksum
}
func (fe *FileEntry) SetComments(comments string) {
	if comments == "" {
		gl.Log("error", "Comments cannot be empty")
		return
	}
	fe.Comments = comments
}
func (fe *FileEntry) SetMetadata(metadata it.IJsonB) {
	if metadata == nil {
		gl.Log("error", "Metadata cannot be nil")
		return
	}
	fe.Metadata = metadata
}
func (fe *FileEntry) SetParent(parent it.IFileEntry) {
	if parent == nil {
		gl.Log("error", "Parent cannot be nil")
		return
	}
	if parent.GetID() == uuid.Nil {
		gl.Log("error", "Parent ID cannot be nil")
		return
	}
	fe.Parent = parent
	fe.ParentID = parent.GetID()
}
func (fe *FileEntry) String() string {
	return fmt.Sprintf("FileEntry(ID: %s, ParentID: %s, Type: %s, Name: %s, OriginName: %s, Depth: %d, Size: %d, CreatedAt: %s, CreatedBy: %s, ModifiedAt: %v, ModifiedBy: %s, Permissions: %s, Checksum: %s, Comments: %s)",
		fe.ID.String(), fe.ParentID.String(), fe.Type, fe.Name, fe.OriginName, fe.Depth, fe.Size,
		fe.CreatedAt.Format(time.RFC3339), fe.CreatedBy,
		func() *string {
			if fe.ModifiedAt != nil {
				modifiedAt := fe.ModifiedAt.Format(time.RFC3339)
				return &modifiedAt
			}
			return nil
		}(),
		fe.ModifiedBy, fe.Permissions, fe.Checksum, fe.Comments)
}
