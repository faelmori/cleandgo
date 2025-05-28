package types

import (
	"fmt"
	"time"

	gl "github.com/faelmori/cleandgo/logger"
	"github.com/google/uuid"
)

type FileEntry struct {
	*Mutexes
	ID          uuid.UUID  `json:"id" yaml:"id" xml:"id" toml:"id" gorm:"type:uuid,default:uuid_generate_v4()"`                 // ID único do arquivo ou diretório
	ParentID    uuid.UUID  `json:"parentId" yaml:"parentId" xml:"parentId" toml:"parentId" gorm:"foreignkey:ID;type:uuid"`      // ID do diretório pai (`00000000-0000-0000-0000-000000000000` se for raiz)
	Type        string     `json:"type" yaml:"type" xml:"type" toml:"type" gorm:"type"`                                         // "file" ou "directory"
	Name        string     `json:"name" yaml:"name" xml:"name" toml:"name" gorm:"name"`                                         // Nome real do artefato
	OriginName  string     `json:"originName" yaml:"originName" xml:"originName" toml:"originName" gorm:"omitempty,originName"` // Nome original do arquivo de árvore
	Depth       int        `json:"depth" yaml:"depth" xml:"depth" toml:"depth" gorm:"depth"`                                    // Profundidade hierárquica
	Size        int64      `json:"size" yaml:"size" xml:"size" toml:"size" gorm:"omitempty,size"`                               // Tamanho do arquivo em bytes
	CreatedAt   time.Time  `json:"createdAt" yaml:"createdAt" xml:"createdAt" toml:"createdAt" gorm:"createdAt"`                // Data de criação
	CreatedBy   string     `json:"createdBy" yaml:"createdBy" xml:"createdBy" toml:"createdBy" gorm:"createdBy"`                // Usuário que criou o arquivo
	ModifiedAt  *time.Time `json:"modifiedAt" yaml:"modifiedAt" xml:"modifiedAt" toml:"modifiedAt" gorm:"omitempty,modifiedAt"` // Data de modificação
	ModifiedBy  string     `json:"modifiedBy" yaml:"modifiedBy" xml:"modifiedBy" toml:"modifiedBy" gorm:"omitempty,modifiedBy"` // Usuário que modificou o arquivo
	Permissions string     `json:"permissions" yaml:"permissions" xml:"permissions" toml:"permissions" gorm:"permissions"`      // Permissões do arquivo (ex: "rwxr-xr-x")
	Checksum    string     `json:"checksum" yaml:"checksum" xml:"checksum" toml:"checksum" gorm:"omitempty,checksum"`           // Checksum do arquivo para integridade
	Metadata    JsonB      `json:"metadata" yaml:"metadata" xml:"metadata" toml:"metadata" gorm:"omitempty,type:jsonb"`         // Metadados adicionais em formato JSON
	Parent      *FileEntry `json:"parent" yaml:"parent" xml:"parent" toml:"parent" gorm:"foreignkey:ID;association_foreignkey:ParentID"`
}

func NewFileEntry(id, parentID uuid.UUID, entryType, name, originName string, depth int, size int64) (*FileEntry, error) {
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
	}, nil
}
