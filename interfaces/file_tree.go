package interfaces

import "github.com/google/uuid"

type IFileTree interface {
	GetDirectoriesIcons() []string
	GetFilesIcons() []string
	GetDrawedMap() map[string]string
	GetEntries() []IFileEntry
	GetEntryByName(name string) IFileEntry
	AddEntry(entry IFileEntry)
	SetEntriesDepth()
	GetEntryByID(id uuid.UUID) IFileEntry
	GetChildren(parentID uuid.UUID) []IFileEntry
	Sanitize(dirtyData []byte) error
	ParseTree() error
	SerializeToFile(format string) error
	LoadFromFile(format string) error
	BackupTreeFile() error
	RestoreTreeFile() error
	SetMaxDepth(depth int)
	GetRootID() uuid.UUID
	SetRootID(id uuid.UUID)
}
