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
	"sync"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var _ = fs.Node(&dir{})
var _ = fs.NodeRequestLookuper(&dir{})
var _ = fs.HandleReadDirAller(&dir{})

type dir struct {
	device  string
	sizeMap map[int]uint64
	lock    sync.Mutex
}

func newDir(device string) *dir {
	dir := &dir{
		device:  device,
		sizeMap: getPlaylistsSizes(device),
	}

	return dir
}

func (d *dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0755
	attr.Valid = 24 * 7 * 365 * time.Hour
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
			Name: fmt.Sprintf("%05d.m2ts", d),
			Type: fuse.DT_File,
		}

		ret = append(ret, de)
	}

	return ret, nil
}

func getPlaylistsSizes(device string) map[int]uint64 {
	playListSizeMap := make(map[int]uint64)

	bd := C.bd_open(C.CString(device), nil)
	if bd == nil {
		fmt.Errorf("Error opening %s\n", device)
		return playListSizeMap
	}
	defer C.bd_close(bd)

	numTitles := C.bd_get_titles(bd, C.TITLES_RELEVANT, 0)
	var i C.uint

	for i = 0; i < numTitles; i++ {
		ti := C.bd_get_title_info(bd, i, 0)
		if ti == nil {
			fmt.Printf("couldn't get title info for %v:%v", device, i)
			continue
		}

		if C.bd_select_title(bd, i) <= 0 {
			fmt.Printf("error opening title %v", i)
			continue
		}

		playListSizeMap[int(ti.playlist)] = uint64(C.bd_get_title_size(bd))
		C.bd_free_title_info(ti)
	}

	return playListSizeMap
}
