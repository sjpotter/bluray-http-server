package bd_fs

/*
#cgo pkg-config: libbluray
#include <stdlib.h>

#include <libbluray/bluray.h>
*/
import "C"
import (
	"bazil.org/fuse/fs"
	"fmt"
)

type FS struct {
	Device string
}
var _ fs.FS = (*FS)(nil)

func (f *FS) Root() (fs.Node, error) {
	n := &dir{
		device: f.Device,
		sizeMap: getPlaylistsSizes(f.Device),
	}

	return n, nil
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