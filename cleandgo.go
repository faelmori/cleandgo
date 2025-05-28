package cleandgo

import (
	it "github.com/faelmori/cleandgo/interfaces"
	t "github.com/faelmori/cleandgo/types"
	l "github.com/faelmori/logz"
	"github.com/google/uuid"
)

type FileTree = it.IFileTree
type FileTreeType = t.FileTree

type FileEntry = it.IFileEntry
type FileEntryType = t.FileEntry

func NewFileTree(treeFileSource, composerTargetPath string, printTree bool, logger l.Logger, debug bool) (it.IFileTree, error) {
	return t.NewFileTree(treeFileSource, composerTargetPath, printTree, logger, debug)
}

func NewFileEntry(id, parentID uuid.UUID, entryType, name, originName string, depth int, size int64, comments string) (it.IFileEntry, error) {
	return t.NewFileEntry(id, parentID, entryType, name, originName, depth, size, comments)
}
