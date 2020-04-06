package handlers

/*
#cgo pkg-config: libbluray
#include <stdlib.h>

#include <libbluray/bluray.h>
*/
import "C"

import (
	"fmt"
	"io"

	"github.com/sjpotter/bluray-http-server/pkg/pkg/types"
)

func NewBDReadSeeker(file string, playlist int, seekTime int) (*BDReadSeeker, error) {
	bd := C.bd_open(C.CString(file), nil)
	if bd == nil {
		return nil, fmt.Errorf("Error opening %s\n", file)
	}

	title, err := findTitle(bd, playlist)
	if err != nil {
		return nil, err
	}

	if C.bd_select_title(bd, C.uint(title)) <= 0 {
		return nil, fmt.Errorf("error opening title %v", title)
	}

	size := C.bd_get_title_size(bd)

	start := int64(0)
	if seekTime != 0 {
		cTime := C.ulong(seekTime * 90000)
		C.bd_seek_time(bd, cTime)
		cur := C.bd_tell(bd)
		start = int64(cur)
		fmt.Printf("start is now %v\n", start)
	}

	return &BDReadSeeker{bd: bd, title: title, start: start, size: int64(size)}, nil
}

type BDReadSeeker struct {
	bd    *C.BLURAY
	title int
	start int64
	size  int64
}

func (b *BDReadSeeker) Read(buf []byte) (int, error) {
	p := C.malloc(6144)
	defer C.free(p)

	size := C.bd_read(b.bd, (*C.uchar)(p), 6144)

	data := C.GoBytes(p, size)

	copy(buf, data)

	return int(size), nil
}

/* Seek needs to work based on an offset from where the user requested the playlist start from */
func (b *BDReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		offset += b.start
	case io.SeekCurrent:
		cur := C.bd_tell(b.bd)
		offset += int64(cur)
		if offset < b.start {
			offset = b.start
		}
	case io.SeekEnd:
		offset += b.size
	}

	C.bd_seek(b.bd, C.ulong(offset))

	return offset - b.start, nil
}

func (b *BDReadSeeker) Close() {
	C.bd_close(b.bd)
}

func (b *BDReadSeeker) ParseTile() (*types.BDTitle, error) {
	ti := C.bd_get_title_info(b.bd, C.uint(b.title), 0)
	defer func() {
		C.bd_free_title_info(ti)
	}()

	return parseTitle(ti)
}

func findTitle(bd *C.BLURAY, playlist int) (int, error) {
	numTitles := C.bd_get_titles(bd, C.TITLES_RELEVANT, 0)
	var i C.uint

	for i = 0; i < numTitles; i++ {
		ti := C.bd_get_title_info(bd, i, 0)
		if int(ti.playlist) == playlist {
			return int(ti.idx), nil
		}
	}

	return -1, fmt.Errorf("could not find a title for playlist %v", playlist)
}
