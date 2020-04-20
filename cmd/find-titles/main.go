package main

/*
#cgo pkg-config: libbluray
#include <stdlib.h>

#include <libbluray/bluray.h>
*/
import "C"
import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	m2ts_fs "github.com/sjpotter/bluray-http-server/pkg/m2ts-fs"
)

var (
	iso = flag.String("iso", "", "iso file to inspect")
	minLength = flag.Int64("minimum length (minutes)", 60, "minimum length to enumerate")
	findOne   = flag.Bool("find-one", false, "only enumerate one")
)

func main() {
	flag.Parse()

	if *iso == "" {
		flag.Usage()
		return
	}

	dir := filepath.Dir(*iso)
	name := filepath.Base(*iso)

	bd := C.bd_open(C.CString(*iso), nil)
	if bd == nil {
		fmt.Printf("Error opening %s\n", *iso)
		os.Exit(1)
	}
	defer C.bd_close(bd)

	numTitles := C.bd_get_titles(bd, C.TITLES_RELEVANT, 0)

	var i uint32 = 0
	var playlists []uint32

	for i := 0; i < int(numTitles); i++ {
		ti := C.bd_get_title_info(bd, C.uint(i), 0)
		if ti == nil {
			continue
		}
		if (int64(ti.duration) / 90000) / 60 > *minLength {
			playlists = append(playlists, uint32(ti.playlist))
		}
		C.bd_free_title_info(ti)
	}

	if len(playlists) > 1 && *findOne {
		fmt.Printf("Found %v titles over %v minutes long\n", len(playlists), *minLength)
		return
	}

	for _, i = range playlists {
		info := m2ts_fs.M2TSInfo{
			File: name,
			Playlist: int(i),
		}

		data, err := yaml.Marshal(info)
		if err != nil {
			fmt.Printf("failed to marshal %+v: %v\n", info, err)
			return
		}

		m2tsFileName := fmt.Sprintf("%v-%v.m2ts", name, i)
		m2tsFileName = filepath.Join(dir, m2tsFileName)
		f, err := os.Create(m2tsFileName)
		if err != nil {
			fmt.Printf("failed to create %v: %v\n", m2tsFileName, err)
			return
		}

		f.Write(data)
		f.Close()
	}
}
