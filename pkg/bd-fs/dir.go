package bd_fs

/*
#cgo pkg-config: libbluray
#include <stdlib.h>

#include <libbluray/bluray.h>
*/
import "C"
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var _ = fs.Node(&dir{})
var _ = fs.NodeRequestLookuper(&dir{})
var _ = fs.HandleReadDirAller(&dir{})

type dir struct {
	device string
	sizeMap map[int]uint64
}

func (d *dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	 attr.Mode = os.ModeDir | 0755
	 attr.Valid = 24*7*365*time.Hour
	 return nil
}

func (d *dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if !strings.HasSuffix(req.Name, ".m2ts") {
		return nil, syscall.ENOENT
	}

	playlist, err := strconv.Atoi(strings.TrimSuffix(req.Name, filepath.Ext(".m2ts")))
	if err != nil {
		return nil, syscall.ENOENT
	}

	if size, ok := d.sizeMap[playlist]; !ok {
		return nil, syscall.ENOENT
	} else {
		return &m2tsFile{device: d.device, playlist: playlist, size: size}, nil
	}
}

func (d *dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var ret []fuse.Dirent

	for d, _ := range d.sizeMap {
		var de = fuse.Dirent{
			Name:  fmt.Sprintf("%v.m2ts", d),
			Type:  fuse.DT_File,
		}

		ret = append(ret, de)
	}

	return ret, nil
}