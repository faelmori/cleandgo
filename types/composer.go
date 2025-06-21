package types

import (
	"fmt"
	"os"

	it "github.com/rafa-mori/cleandgo/interfaces"
	utl "github.com/rafa-mori/cleandgo/utils"
)

type TreeComposer struct {
	*FileTree
}

func NewTreeComposer(fileTree it.IFileTree) (it.ITreeComposer, error) {
	if _, ok := fileTree.GetFileTreeType().(*FileTree); !ok {
		return nil, fmt.Errorf("invalid file tree type")
	}
	return &TreeComposer{
		FileTree: fileTree.GetFileTreeType().(*FileTree),
	}, nil
}
func (tc *TreeComposer) MakeTreeDirectories() error {
	entries := tc.FileTree.GetEntries()
	for _, entry := range entries {
		if entry.GetType() == "directory" {
			if utl.CheckFileExists(entry.GetPath()) {
				continue
			}
			if err := os.MkdirAll(entry.GetPath(), os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", entry.GetPath(), err)
			}
		}
	}
	return nil
}
func (tc *TreeComposer) MakeTreeFiles() error {
	entries := tc.FileTree.GetEntries()
	for _, entry := range entries {
		if entry.GetType() == "file" {
			if utl.CheckFileExists(entry.GetPath()) {
				continue
			}
			file, err := os.Create(entry.GetPath())
			if err != nil {
				return fmt.Errorf("failed to create file '%s': %w", entry.GetPath(), err)
			}
			defer file.Close()
		}
	}
	return nil
}
func (tc *TreeComposer) MakeTreeSymlinks() error {
	entries := tc.FileTree.GetEntries()
	for _, entry := range entries {
		if entry.GetType() == "symlink" {
			if utl.CheckFileExists(entry.GetPath()) {
				continue
			}
			if err := os.Symlink(entry.GetOriginName(), entry.GetPath()); err != nil {
				return fmt.Errorf("failed to create symlink '%s' -> '%s': %w", entry.GetPath(), entry.GetOriginName(), err)
			}
		}
	}
	return nil
}
func (tc *TreeComposer) MakeTree() error {
	if err := tc.MakeTreeDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	if err := tc.MakeTreeFiles(); err != nil {
		return fmt.Errorf("failed to create files: %w", err)
	}
	if err := tc.MakeTreeSymlinks(); err != nil {
		return fmt.Errorf("failed to create symlinks: %w", err)
	}
	return nil
}
func (tc *TreeComposer) SetFilePermissions(path, permissions string) error {
	perms, err := utl.ParsePermissions(permissions)
	if err != nil {
		return fmt.Errorf("failed to parse permissions for '%s': %w", path, err)
	}
	if err := os.Chmod(path, perms); err != nil {
		return fmt.Errorf("failed to set permissions for '%s': %w", path, err)
	}
	return nil
}
func (tc *TreeComposer) SetFileChecksum(path, checksum string) error {
	if err := utl.SetFileChecksum(path, checksum); err != nil {
		return fmt.Errorf("failed to set checksum for '%s': %w", path, err)
	}
	return nil
}
func (tc *TreeComposer) EnsureTreePermissions() error {
	entries := tc.FileTree.GetEntries()
	for _, entry := range entries {
		if entry.GetPermissions() == "" {
			continue
		}
		if err := tc.SetFilePermissions(entry.GetPath(), entry.GetPermissions()); err != nil {
			return fmt.Errorf("failed to set permissions for '%s': %w", entry.GetPath(), err)
		}
	}
	return nil
}
func (tc *TreeComposer) EnsureTreeChecksums() error {
	entries := tc.FileTree.GetEntries()
	for _, entry := range entries {
		if entry.GetChecksum() == "" {
			continue
		}
		if ok, err := utl.CheckFileChecksum(entry.GetPath(), entry.GetChecksum()); err != nil {
			return fmt.Errorf("failed to check checksum for '%s': %w", entry.GetPath(), err)
		} else if !ok {
			return fmt.Errorf("checksum mismatch for '%s'", entry.GetPath())
		}
	}
	return nil
}
