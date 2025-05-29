package interfaces

type ITreeComposer interface {
	MakeTreeDirectories() error
	MakeTreeFiles() error
	MakeTreeSymlinks() error
	MakeTree() error
	SetFilePermissions(path, permissions string) error
	EnsureTreePermissions() error
	EnsureTreeChecksums() error
}
