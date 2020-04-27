package bd_fs

import "C"
import (
	"bazil.org/fuse/fs"
)

type FS struct {
	Device string
}
var _ fs.FS = (*FS)(nil)

func (f *FS) Root() (fs.Node, error) {
	return newDir(f.Device), nil
}