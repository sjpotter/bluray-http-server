package readers

/*
#cgo pkg-config: libbluray
#include <stdlib.h>

#include <libbluray/bluray.h>
*/
import "C"

import (
	"fmt"
	"io"
	"unsafe"

	"github.com/sjpotter/bluray-http-server/pkg/types"
	"github.com/sjpotter/bluray-http-server/pkg/utils"
)

type BluRayReader interface {
	Read(buf []byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
	Close()
	Size() uint64
}

func NewBDReadSeeker(file string, playlist int) (*BDReadSeeker, error) {
	bd := C.bd_open(C.CString(file), nil)
	if bd == nil {
		return nil, fmt.Errorf("Error opening %s\n", file)
	}

	title, err := findTitle(bd, file, playlist)
	if err != nil {
		return nil, err
	}

	if C.bd_select_title(bd, C.uint(title)) <= 0 {
		return nil, fmt.Errorf("error opening title %v", title)
	}

	size := C.bd_get_title_size(bd)

	return &BDReadSeeker{bd: bd, file: file, title: title, size: int64(size)}, nil
}

var _ BluRayReader = &BDReadSeeker{}

type BDReadSeeker struct {
	bd    *C.BLURAY
	file  string
	title int
	size  int64
}

func (b *BDReadSeeker) Read(buf []byte) (int, error) {
	p := C.malloc(C.ulong(cap(buf)))
	if p != nil {
		defer C.free(p)
	} else {
		return 0, fmt.Errorf("couldn't allocate memory space for the read")
	}

	size := C.bd_read(b.bd, (*C.uchar)(p), C.int(cap(buf)))

	data := C.GoBytes(p, size)

	copy(buf, data)

	return int(size), nil
}

/* Seek needs to work based on an offset from where the user requested the playlist start from */
func (b *BDReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		offset += 0
	case io.SeekCurrent:
		cur := C.bd_tell(b.bd)
		offset += int64(cur)
	case io.SeekEnd:
		offset += b.size
	}

	if offset > b.size {
		offset = b.size
	} else if offset < 0 {
		offset = 0
	}

	C.bd_seek(b.bd, C.ulong(offset))

	return offset, nil
}

func (b *BDReadSeeker) Close() {
	C.bd_close(b.bd)
}

func (b *BDReadSeeker) ParseTile() (*types.BDTitle, error) {
	ti := C.bd_get_title_info(b.bd, C.uint(b.title), 0)
	if ti == nil {
		return nil, fmt.Errorf("couldn't get title info for %v:%v", b.file, b.title)
	}
	defer C.bd_free_title_info(ti)

	return utils.ParseTitle(unsafe.Pointer(ti))
}

func (b *BDReadSeeker) Size() uint64 {
	return uint64(b.size)
}

func findTitle(bd *C.BLURAY, file string, playlist int) (int, error) {
	numTitles := C.bd_get_titles(bd, C.TITLES_RELEVANT, 0)
	var i C.uint

	for i = 0; i < numTitles; i++ {
		ti := C.bd_get_title_info(bd, i, 0)
		if ti == nil {
			fmt.Printf("couldn't get title info for %v:%v", file, playlist)
		}
		if int(ti.playlist) == playlist {
			return int(ti.idx), nil
		}
	}

	return -1, fmt.Errorf("could not find a title for playlist %v", playlist)
}
