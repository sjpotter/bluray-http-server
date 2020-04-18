package m2ts_fs

import (
	"bazil.org/fuse/fs"
	"os"
)

type FS struct {}
var _ fs.FS = (*FS)(nil)

func (f *FS) Root() (fs.Node, error) {
	root, _ := os.Getwd()
	n := &dir{
		path: root,
	}

	return n, nil
}



